package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/stuartdd/webServerBase/config"
	"github.com/stuartdd/webServerBase/logging"
	"github.com/stuartdd/webServerBase/servermain"
)

var log *logging.LoggerDataReference
var serverInstance *servermain.ServerInstanceData

type largeFileData struct {
	Name      string
	Offsets   []int64
	LineCount int
	ModTime   time.Time
	Time      time.Time
}

var pageMap = make(map[string]*largeFileData)

const banner = "\n---------------------------------------------------------------------------------------------------------------"

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

	scriptData := config.GetConfigDataInstance().GetScriptDataForOS()
	serverInstance.SetOsScriptsData(scriptData.Path, scriptData.Data)
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
	serverInstance.AddMappedHandler("/stop", http.MethodGet, servermain.StopServerInstance)
	serverInstance.AddMappedHandlerWithNames("/stop/?", http.MethodGet, servermain.StopServerInstance, []string{"seconds"})
	serverInstance.AddMappedHandler("/status", http.MethodGet, servermain.StatusHandler)
	serverInstance.AddMappedHandler("/static/*", http.MethodGet, servermain.DefaultStaticFileHandler)
	serverInstance.AddMappedHandlerWithNames("/script/?", http.MethodGet, servermain.DefaultOSScriptHandler, []string{"script"})
	serverInstance.AddMappedHandlerWithNames("/site/?", http.MethodGet, servermain.DefaultTemplateFileHandler, []string{"template"})
	serverInstance.AddMappedHandlerWithNames("/calc/qube/?", http.MethodGet, qubeHandler, []string{"qube"})
	serverInstance.AddMappedHandlerWithNames("/calc/?/div/?", http.MethodGet, divHandler, []string{"calc", "div"})
	serverInstance.AddMappedHandlerWithNames("/path/?/file/?", http.MethodPost, fileSaveHandler, []string{"path", "filename"})
	serverInstance.AddMappedHandlerWithNames("/path/?/file/?/ext/?", http.MethodPost, fileSaveHandler, []string{"path", "filename", "ext"})
	serverInstance.AddMappedHandlerWithNames("/large/?/file/?/ext/?/page/?", http.MethodPost, fileLargeHandler, []string{"path", "filename", "ext", "page"})

	/*
		An after handler is executed after ALL requests have been handled
		You can add multiple after handlers.
		If the after handler vetos the request then the after handler response is returned not the mapped handler's response
	*/
	serverInstance.AddAfterHandler(filterAfter)
	defer checkForPanicAndRecover()
	log.LogInfof(banner[2:22]+" Server starting port:%d "+banner[2:22], configData.Port)
	serverInstance.ListenAndServeOnPort(configData.Port)
}

func checkForPanicAndRecover() {
	rec := recover()
	if rec != nil {
		recStr := fmt.Sprintf("%s", rec)
		logging.LogDirectToSystemError(banner+"\nAn UN-HANDLED error occured that terminated the server:\nError message: "+recStr+banner, true)
		os.Exit(9)
	}
}

/************************************************
Start of handlers section
*************************************************/

func fileLargeHandler(r *http.Request, response *servermain.Response) {
	h := servermain.NewRequestHandlerHelper(r, response)
	fileName := h.GetNamedURLPart("filename", "") // Not optional
	pathName := h.GetNamedURLPart("path", "")     // Not optional
	ext := h.GetNamedURLPart("ext", "txt")        // Optional. Default value txt
	// line := h.GetNamedURLPart("line", "-1")         // Optional. Default value txt
	fullFile := filepath.Join(h.GetStaticPathForName(pathName).FilePath, fileName+"."+ext)
	fileID := h.GetUUID() + ".tracked"
	if pageMap[fileID] == nil {
		pageMap[fileID] = openInitial(fullFile, 5)
	}
	response.SetResponse(201, "{\"ref\":\""+fileID+"\"}", "application/json")
}

func (p *largeFileData) largeFileHandlerReader(from int, count int) string {

	info, err := os.Stat(p.Name)
	if err != nil {
		servermain.ThrowPanic("E", 404, servermain.SCFileNotFound, "Not Found", fmt.Sprintf("File %s could not be found. %s", p.Name, err.Error()))
	}
	if count == 0 {
		return ""
	}
	to := from + count

	if to >= p.LineCount {
		if info.ModTime().After(p.ModTime) {
			/*
				File has changed
			*/
		}
	}
	if to < from {
		return ""
	}

	/*
		If from is 0 then read from the start!
	*/
	var start int64 = 0
	if from > 0 {
		start = p.Offsets[from-1]
	}

	var end int64 = p.Offsets[p.LineCount-1]
	if to <= p.LineCount {
		end = p.Offsets[to-1]
	}

	bytesToRead := (end - start) + 1
	if bytesToRead < 1 {
		return ""
	}
	buf := make([]byte, bytesToRead)

	f, err := os.Open(p.Name)
	if err != nil {
		servermain.ThrowPanic("E", 417, servermain.SCOpenFileError, "Expectation Failed", fmt.Sprintf("File %s could not be opened. %s", p.Name, err.Error()))
	}
	defer f.Close()
	/*
		Read from a point in the file
	*/
	if start > 0 {
		_, err = f.Seek(start, 0)
		if err != nil {
			servermain.ThrowPanic("E", 417, servermain.SCOpenFileError, "Expectation Failed", fmt.Sprintf("File %s could not seek. %s", p.Name, err.Error()))
		}
	}

	/*
		Read the rquired number of bytes
	*/
	bytes, err := io.ReadAtLeast(f, buf, int(bytesToRead))
	checkOpenInitialError(p.Name, err)

	if bytes < 1 {
		return ""
	}
	/*
		Skip bytes from the start of the line is there!
	*/
	bufFrom := 0
	if buf[bufFrom] == 13 || buf[bufFrom] == 10 {
		bufFrom++
	}
	return string(buf[bufFrom:bytes])
}

func openInitial(name string, blocks int) *largeFileData {
	if blocks == 0 {
		servermain.ThrowPanic("E", 500, servermain.SCParamValidation, "Internal Server Error", "Internal error: openInitial-->blocks Parameter cannot be 0")
	}
	info, err := os.Stat(name)
	if err != nil {
		servermain.ThrowPanic("E", 404, servermain.SCFileNotFound, "Not Found", fmt.Sprintf("File %s could not be found. %s", name, err.Error()))
	}
	f, err := os.Open(name)
	if err != nil {
		servermain.ThrowPanic("E", 417, servermain.SCOpenFileError, "Expectation Failed", fmt.Sprintf("File %s could not be opened. %s", name, err.Error()))
	}
	defer f.Close()

	var offset int64 = 0        // Offset in to the file!
	bytesRead := 0              // The number of bytest read
	notEOF := true              // Are we at the end of the file
	buf := make([]byte, blocks) // Buffer for the file

	data := &largeFileData{
		Name:      name,
		Offsets:   make([]int64, 50), // Make room for 100 lines
		LineCount: 0,
		ModTime:   info.ModTime(), // The time the file was updated
		Time:      time.Now(),     // The time we read the file
	}
	/*
		While not at the end of the file
	*/
	for notEOF {
		bytesRead, err = io.ReadAtLeast(f, buf, blocks)
		notEOF = checkOpenInitialError(name, err)
		offset = data.parseOpenInitial(bytesRead, buf, offset)
	}
	/*
		Add an empty line to the end of the file so me now how big the file is
	*/
	offset = data.parseOpenInitial(2, []byte{10, 32}, offset)
	return data
}

/*
for each new line add the offset in the file to that line in the offsets
*/
func (p *largeFileData) parseOpenInitial(bytesRead int, b []byte, offset int64) int64 {
	for i := 0; i < bytesRead; i++ {
		if b[i] == 10 {
			if p.LineCount >= len(p.Offsets) {
				newLen := p.LineCount + 50
				sb := make([]int64, newLen)
				for i := 0; i < p.LineCount; i++ {
					sb[i] = p.Offsets[i]
				}
				p.Offsets = sb
			}
			p.Offsets[p.LineCount] = offset
			p.LineCount++
		}
		offset++
	}
	return offset
}

func checkOpenInitialError(name string, err error) bool {
	if err != nil {
		switch err {
		case io.EOF, io.ErrUnexpectedEOF:
			return false
		default:
			servermain.ThrowPanic("E", 400, servermain.SCOpenFileError, "Read File", fmt.Sprintf("File %s could not read. %s", name, err.Error()))
		}
	}
	return true
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
	fileName := h.GetNamedURLPart("filename", "") // Not optional
	pathName := h.GetNamedURLPart("path", "")     // Not optional
	ext := h.GetNamedURLPart("ext", "txt")        // Optional. Default value txt
	fullFile := filepath.Join(h.GetStaticPathForName(pathName).FilePath, fileName+"."+ext)
	err := ioutil.WriteFile(fullFile, h.GetBody(), 0644)
	if err != nil {
		servermain.ThrowPanic("E", 400, servermain.SCWriteFile, fmt.Sprintf("fileSaveHandler: static path [%s], file [%s] could not write file", pathName, fileName), err.Error())
	}
	response.SetResponse(201, "{\"Created\":\"OK\"}", "application/json")
}

/*
qubeHandler (example function) - return the qube of the number. E.G. qube of 5 "/calc/qube/5"
*/
func qubeHandler(r *http.Request, response *servermain.Response) {
	h := servermain.NewRequestHandlerHelper(r, response)
	p1 := h.GetNamedURLPart("qube", "")
	a1, err := strconv.Atoi(p1)
	if err != nil {
		servermain.ThrowPanic("E", 400, servermain.SCParamValidation, "invalid number "+p1, err.Error())
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
		servermain.ThrowPanic("E", 400, servermain.SCParamValidation, "invalid number "+p1, err.Error())
	}
	a2, err := strconv.Atoi(p2)
	if err != nil {
		servermain.ThrowPanic("E", 400, servermain.SCParamValidation, "invalid number "+p2, err.Error())
	}
	response.SetResponse(200, strconv.Itoa(a1/a2), "")
}

func filterBefore(r *http.Request, response *servermain.Response) {
	log.LogDebug("IN Filter Before 1")
}

func filterAfter(r *http.Request, response *servermain.Response) {
	log.LogDebug("IN Filter After 1")
}

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
		if config.GetConfigDataInstance().TemplateData != nil {
			configData := config.GetConfigDataInstance().TemplateData[templateName]
			if configData != nil {
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
