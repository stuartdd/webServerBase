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
	serverInstance.SetPathToTemplates(configData.GetConfigDataTemplateFilePathForOS())
	log.LogInfof("Availiable Templates: %s\n", serverInstance.ListTemplateNames(", "))
	/*
		Add too or override the Default content types
	*/
	serverInstance.AddContentTypeFromMap(configData.ContentTypes)
	/*
		Add redirections from the config data
	*/
	serverInstance.SetRedirections(config.GetConfigDataInstance().Redirections)
	/*
		Add static file paths.
	*/
	//	serverInstance.SetStaticFileDataFromMap(config.GetConfigDataStaticPathForOS())
	/*
		Add template file path (singular).
	*/
	//	serverInstance.SetTemplatesPath(config.GetConfigDataTemplatePathForOS())
	/*
		Set the http status code returned if a panic is thrown by any od the handlers
	*/
	serverInstance.SetPanicStatusCode(configData.PanicResponseCode)

	serverInstance.AddBeforeHandler(filterBefore)
	serverInstance.AddMappedHandler("/stop", http.MethodGet, stopServerInstance)
	serverInstance.AddMappedHandler("/stop/?", http.MethodGet, stopServerInstance)
	serverInstance.AddMappedHandler("/status", http.MethodGet, statusHandler)
	serverInstance.AddMappedHandler("/static/?", http.MethodGet, servermain.ReasonableStaticFileHandler)
	serverInstance.AddMappedHandler("/calc/qube/?", http.MethodGet, qubeHandler)
	serverInstance.AddMappedHandler("/calc/?/div/?", http.MethodGet, divHandler)
	serverInstance.AddMappedHandler("/path/?/file/?", http.MethodPost, fileSaveHandler)
	serverInstance.AddMappedHandler("/path/?/file/?/ext/?", http.MethodPost, fileSaveHandler)

	serverInstance.AddAfterHandler(filterAfter)

	serverInstance.ListenAndServeOnPort(configData.Port)
}

/************************************************
Start of handlers section
*************************************************/
func fileSaveHandler(r *http.Request, response *servermain.Response) {
	d := servermain.NewRequestHandlerHelper(r, response)
	fileName := d.GetNamedURLPart("file", "") // Not optional
	pathName := d.GetNamedURLPart("path", "") // Not optional
	ext := d.GetNamedURLPart("ext", "txt")    // Optional. Default value txt
	fullFile := filepath.Join(d.GetStaticPathForName(pathName).FilePath, fileName+"."+ext)
	err := ioutil.WriteFile(fullFile, d.GetBody(), 0644)
	if err != nil {
		servermain.ThrowPanic("E", 400, SCWriteFile, fmt.Sprintf("fileSaveHandler: static path [%s], file [%s] could not write file", pathName, fileName), err.Error())
	}
	response.SetResponse(201, "{\"Created\":\"OK\"}", "application/json")
}

func stopServerInstance(r *http.Request, response *servermain.Response) {
	d := servermain.NewRequestHandlerHelper(r, response)
	count, err := strconv.Atoi(d.GetNamedURLPart("stop", "2"))
	if err != nil {
		servermain.ThrowPanic("E", 400, SCParamValidation, "Invalid stop period", err.Error())
	} else {
		serverInstance.StopServerLater(count, fmt.Sprintf("Stopped by request. Delay %d seconds", count))
		response.SetResponse(200, serverInstance.GetStatusData(), "application/json")
	}
}

func qubeHandler(r *http.Request, response *servermain.Response) {
	d := servermain.NewRequestHandlerHelper(r, response)
	p1 := d.GetNamedURLPart("qube", "")
	a1, err := strconv.Atoi(p1)
	if err != nil {
		servermain.ThrowPanic("E", 400, SCParamValidation, "invalid number "+p1, err.Error())
	}
	response.SetResponse(200, strconv.Itoa(a1*a1*a1*a1), "")
}

func divHandler(r *http.Request, response *servermain.Response) {
	d := servermain.NewRequestHandlerHelper(r, response)
	p1 := d.GetNamedURLPart("calc", "")
	p2 := d.GetNamedURLPart("div", "")
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