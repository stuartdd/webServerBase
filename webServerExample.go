package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/stuartdd/webServerBase/config"
	"github.com/stuartdd/webServerBase/logging"
	"github.com/stuartdd/webServerBase/servermain"
)

/*
SCStaticPath and these constants are used as unique subcodes in error responses.
They MUST start wit the highest sub code defined in  servermain SCMax
*/
const (
	SCStaticPath = iota + servermain.SCMax + 1
	SCWriteFile
	SCParamValidation
)

var log *logging.LoggerDataReference
var serverInstance *servermain.ServerInstanceData

func main() {

	/*
		derive the name of the application
	*/ 
	exec := config.GetApplicationModuleName()

	/*
		Read the configuration file. If no name is given use the default name.
	*/
	var configFileName string
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
		if !strings.HasSuffix(strings.ToLower(configFileName), ".json") {
			configFileName = configFileName + ".json"
		}
	}
	err := config.LoadConfigData(configFileName)
	if err != nil {
		log.Fatal(err)
	}

	configData := config.GetConfigDataInstance()
	/*
		Initialiase the logs. Log name is in the config data. If not defined default to sysout
	*/
	logging.CreateLogWithFilenameAndAppID(configData.DefaultLogFileName, exec+":"+strconv.Itoa(configData.Port), 1, configData.LoggerLevels)
	/*
		Stack the defered processes (Last in First out)
	*/
	defer ExitApplication()
	defer CloseLog()

	RunWithConfig(config.GetConfigDataInstance(), exec)
}

/*

RunWithConfig runs with a sgeneric handler
Param - config a ref to the config object
*/
func RunWithConfig(configData *config.Data, executable string) {
	/*
		Add loggers for each module (Makes for neater logs!)
	*/
	log = logging.NewLogger("ServerMain")
	/*
		Log server startup info
	*/
	log.LogInfof("Server will start on port %d\n", configData.Port)
	log.LogInfof("OS '%s'. Static path will be:%s\n", runtime.GOOS, configData.GetConfigDataStaticFilePathForOS())
	log.LogInfof("To stop the server http://localhost:%d/stop\n", configData.Port)
	/*
		Configure and Start the server.
	*/
	serverInstance = servermain.NewServerInstanceData(executable, configData.ContentTypeCharset)
	/*
		Set the static file data paths for the given OS. When this is done we can add the handler.
	*/
	serverInstance.SetStaticFileServerData(configData.GetConfigDataStaticFilePathForOS())
	/*
		Set up the templates directory
	*/
	serverInstance.SetPathToTemplates(configData.GetConfigDataTemplateFilePathForOS())
	/*
	Add the list of templates to the log
	*/
	log.LogInfof("Availiable Templates: %s\n", serverInstance.ListTemplateNames(", "))
	/*
		When we return template "index1.html" we need to provide some data so provide a dataProvider
	*/
	serverInstance.AddTemplateDataProvider(templateDataProvider)
	/*
		Add too or override the Default content types
	*/
	serverInstance.AddContentTypeFromMap(configData.ContentTypes)
	/*
		Add redirections from the config data
	*/
	serverInstance.SetRedirections(config.GetConfigDataInstance().Redirections)
	/*
		Set the http status code returned if a panic is thrown by any od the handlers
	*/
	serverInstance.SetPanicStatusCode(configData.PanicResponseCode)
	/*
	A before handler is executed before ALL requests are passed to Mapped handlers
	You can add multiple before handlers. These are usefull for access control and other global checks
	If the before handler vetos the request then the Mapped handlers are not called
	*/
	serverInstance.AddBeforeHandler(filterBefore)
	/*
	Add a function for all URL mappings. A ? matches ANY value. A * indicates any value after the match 
	E.G. /x/y/* will activate for x/y/1/d/3/4/5/
	E.G. /x/?/y wil activate for /x/1/y
	*/
	serverInstance.AddMappedHandler("/stop", http.MethodGet, stopServerInstance)
	serverInstance.AddMappedHandler("/stop/?", http.MethodGet, stopServerInstance)
	serverInstance.AddMappedHandler("/status", http.MethodGet, statusHandler)
	serverInstance.AddMappedHandler("/static/?", http.MethodGet, servermain.ReasonableStaticFileHandler)
	serverInstance.AddMappedHandler("/site/?", http.MethodGet, servermain.ReasonableTemplateFileHandler)
	serverInstance.AddMappedHandler("/calc/qube/?", http.MethodGet, qubeHandler)
	serverInstance.AddMappedHandler("/calc/?/div/?", http.MethodGet, divHandler)
	serverInstance.AddMappedHandler("/path/?/file/?", http.MethodPost, fileSaveHandler)
	serverInstance.AddMappedHandler("/path/?/file/?/ext/?", http.MethodPost, fileSaveHandler)
	/*
	An after handler is executed after ALL requests have been handles
	You can add multiple after handlers. 
	If the after handler vetos the request then the after handler response is returned not the mapped handler's response
	*/

	serverInstance.AddAfterHandler(filterAfter)

	serverInstance.ListenAndServeOnPort(configData.Port)
}

/************************************************
Start of handlers section
*************************************************/

/*
templateDataProvider (example method) - When a template is executed this method is called.
ref above: serverInstance.AddTemplateDataProvider(TemplateDataProviderOne)

The data object is returned to the template for substitution.

In this case the handler ReasonableTemplateFileHandler in file templateManager.go passes in the Query arguments
from the URL as a map as the data object.

In this method if it is a map we can assume ReasonableTemplateFileHandler has been invoked.

The test in webServerExample_test.go sends 'site/index1.html?Material=LEAD' so Material would be LEAD.
This data provider reads the confog TemplateData maps and merges the map associated with the template.
The test asserts that Soot is returned in the template confirming that this provider has been called 
and the config data has been read.
*/
func templateDataProvider(r *http.Request, templateName string, data interface{}) {
	/*
	Only get involved if the data is a map
	*/
	v, ok := data.(map[string]string)
	if ok {
		/*
		Look up a data map in the configuration data using the template name 
		*/
		if (config.GetConfigDataInstance().TemplateData != nil) {
			configData := config.GetConfigDataInstance().TemplateData[templateName]
			if (configData != nil) {
				/*
				Merge the existing data and the config data. 
				Duplicate values in config data will take presedence.
				*/
				for name, value := range configData {
					v[name] = value
				}
			}
		}
		/*
		Add the template name in for good measure!
		*/
		v["TemplateName"] = templateName
	}
}

/*
fileSaveHandler (example handler) will save the POST message body in a file
defined by the URL parameters
/path/?/file/? - Save the body at the static path ? with the file name ?
/path/?/file/?/ext/? - Save the body at the static path ? with the file name ? and extension ?
both URLs invoke this function

Note the path MUST be found in the static file mappings (via GetStaticPathForName)
So if the value is /path/data/file/fn/ext/txt and the mapping is defined as 
{"/static/":"site\\", "data":"saved\\"}
Then the file is saved as saved\\fn.txt. Otherwise a file not found is returned
*/
func fileSaveHandler(r *http.Request, response *servermain.Response) {
	h := servermain.NewRequestHandlerHelper(r, response)
	fileName := h.GetNamedURLPart("file", "") // Not optional
	pathName := h.GetNamedURLPart("path", "") // Not optional
	ext := h.GetNamedURLPart("ext", "txt")    // Optional. Default value txt
	fullFile := filepath.Join(h.GetStaticPathForName(pathName).FilePath, fileName+"."+ext)
	err := ioutil.WriteFile(fullFile, h.GetBody(), 0644)
	if err != nil {
		servermain.ThrowPanic("E", 400, SCWriteFile, fmt.Sprintf("fileSaveHandler: static path [%s], file [%s] could not write file", pathName, fileName), err.Error())
	}
	response.SetResponse(201, "{\"Created\":\"OK\"}", "application/json")
}

/*
stopServerInstance - Stops the server in N seconds defined by optional URL parameter.
Note that the delay is so the response can be processed and returned to the client (or browser)
/stop
/stop/?
both invoke this function
*/
func stopServerInstance(r *http.Request, response *servermain.Response) {
	h := servermain.NewRequestHandlerHelper(r, response)
	count, err := strconv.Atoi(h.GetNamedURLPart("stop", "2")) // Optional. Default value 2
	if err != nil {
		servermain.ThrowPanic("E", 400, SCParamValidation, "Invalid stop period", err.Error())
	} else {
		serverInstance.StopServerLater(count, fmt.Sprintf("Stopped by request. Delay %d seconds", count))
		response.SetResponse(200, serverInstance.GetStatusData(), "application/json")
	}
}

/*
qubeHandler (example function) - return the qube of the number. E.G. qube of 5 "/calc/qube/5"
*/
func qubeHandler(r *http.Request, response *servermain.Response) {
	h := servermain.NewRequestHandlerHelper(r, response)
	p1 := h.GetNamedURLPart("qube", "")
	a1, err := strconv.Atoi(p1)
	if err != nil {
		servermain.ThrowPanic("E", 400, SCParamValidation, "invalid number "+p1, err.Error())
	}
	response.SetResponse(200, strconv.Itoa(a1*a1*a1*a1), "")
}

/*
divHandler (example function) - return the a / b of the number. E.G. /calc/?/div/?
For example /calc/10/div/2 returns 5
This is used to test the exception (panic) handling by using /calc/10/div/0
*/
func divHandler(r *http.Request, response *servermain.Response) {
	h := servermain.NewRequestHandlerHelper(r, response)
	p1 := h.GetNamedURLPart("calc", "")
	p2 := h.GetNamedURLPart("div", "")
	a1, err := strconv.Atoi(p1)
	if err != nil {
		servermain.ThrowPanic("E", 400, SCParamValidation, "invalid number "+p1, err.Error())
	}
	a2, err := strconv.Atoi(p2)
	if err != nil {
		servermain.ThrowPanic("E", 400, SCParamValidation, "invalid number "+p2, err.Error())
	}
	response.SetResponse(200, strconv.Itoa(a1/a2), "")
}

func statusHandler(r *http.Request, response *servermain.Response) {
	response.SetResponse(200, serverInstance.GetStatusData(), "application/json")
}

func filterBefore(r *http.Request, response *servermain.Response) {
	log.LogDebug("IN Filter Before 1")
}

func filterAfter(r *http.Request, response *servermain.Response) {
	log.LogDebug("IN Filter After 1")
}

/************************************************
End of handlers section

Start of utility functions
*************************************************/

/*
CloseLog closes the logger file if it exists
A logger os passed to enable the CloseLog function to log that fact it has been closed!
*/
func CloseLog() {
	code := serverInstance.GetServerReturnCode()
	if code != 0 {
		log.LogErrorf("EXIT: Logs Closed. Exit code %d", code)
	} else {
		log.LogInfo("EXIT: Logs Closed. Exit code 0")
	}
	logging.CloseLog()
}

/*
ExitApplication closes the application. Make sure it happens last

Cannot use logger here as it has been closed, hopefully!
*/
func ExitApplication() {
	code := serverInstance.GetServerReturnCode()
	if code != 0 {
		os.Exit(code)
	}
}
