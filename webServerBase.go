package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"webServerBase/config"
	"webServerBase/logging"
	"webServerBase/servermain"
)

var logger *logging.LoggerDataReference
var serverInstance *servermain.ServerInstanceData
var staticFileServer *servermain.FileServerData

func main() {
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
		logger.Fatal(err)
	}

	exec, err := os.Executable()
	if err != nil {
		exec = "UnknownModule"
	} else {
		parts := strings.Split(exec, fmt.Sprintf("%c", os.PathSeparator))
		exec = parts[len(parts)-1]
		if exec == "debug" {
			exec = parts[len(parts)-2]
		}
	}

	RunWithConfig(config.GetConfigDataInstance(), exec)
}

/*

RunWithConfig runs with a sgeneric handler
Param - config a ref to the config object
*/
func RunWithConfig(configData *config.Data, executable string) {
	/*
		Open the logs. Log name is in the congig data. If not defined default to sysout
	*/
	logging.CreateLogWithFilenameAndAppID(config.GetConfigDataInstance().DefaultLogFileName, executable+":"+strconv.Itoa(config.GetConfigDataInstance().Port), config.GetConfigDataInstance().LoggerLevels)
	/*
		Stack the defered processes (Last in First out)
	*/
	defer ExitApplication()
	defer CloseLog()
	/*
		Add loggers for each module (Makes for neater logs!)
	*/
	logger = logging.NewLogger("ServerMain")
	/*
		Log server startup info
	*/
	logger.LogInfof("Server will start on port %d\n", configData.Port)
	logger.LogInfof("OS '%s'. Static path will be:%s\n", runtime.GOOS, configData.GetConfigDataStaticFilePathForOS())
	logger.LogInfof("To stop the server http://localhost:%d/stop\n", configData.Port)

	/*
		Configure and Start the server.
	*/
	serverInstance = servermain.NewServerInstanceData(executable, configData.ContentTypeCharset)
	/*
		Set the static file data paths for the given OS. When this is done we can add the handler.
	*/
	serverInstance.SetStaticFileServerData(configData.GetConfigDataStaticFilePathForOS())
	serverInstance.SetPathToTemplates(configData.GetConfigDataTemplateFilePathForOS())
	logger.LogInfof("Availiable Templates: %s\n", serverInstance.ListTemplateNames(", "))
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
	serverInstance.AddMappedHandler("/calc/?/div/?", http.MethodGet, divHandler)
	serverInstance.AddAfterHandler(filterAfter)

	serverInstance.ListenAndServeOnPort(configData.Port)
}

/************************************************
Start of handlers section
*************************************************/

func stopServerInstance(r *http.Request, response *servermain.Response) {
	d := servermain.NewURLDetails(r)
	count, err := strconv.Atoi(d.GetNamedPart("stop","2"))
	if err != nil {
		response.ChangeResponse(400, "Invalid stop period", "Ha", err)
	} else {
		serverInstance.StopServerLater(count, fmt.Sprintf("Stopped by request. Delay %d seconds", count))
		response.ChangeResponse(200, serverInstance.GetStatusDataJSON(), "application/json", nil)	
	}
}

func divHandler(r *http.Request, response *servermain.Response) {
	d := servermain.NewURLDetails(r)
	p1 := d.GetNamedPart("calc","undefined")
	p2 := d.GetNamedPart("div","undefined")
	a1, err := strconv.Atoi(p1)
	if err != nil {
		response.ChangeResponse(400, "invalid number "+p1, "", err)
	}
	a2, err := strconv.Atoi(p2)
	if err != nil {
		response.ChangeResponse(400, "invalid number "+p2, "", err)
	}
	response.ChangeResponse(200, strconv.Itoa(a1/a2), "", nil)
}

func statusHandler(r *http.Request, response *servermain.Response) {
	response.ChangeResponse(200, serverInstance.GetStatusDataJSON(), "application/json", nil)
}

func filterBefore(r *http.Request, response *servermain.Response) {
	logger.LogDebug("IN Filter Before 1")
}

func filterAfter(r *http.Request, response *servermain.Response) {
	logger.LogDebug("IN Filter After 1")
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
		logger.LogErrorf("EXIT: Logs Closed. Exit code %d", code)
	} else {
		logger.LogInfo("EXIT: Logs Closed. Exit code 0")
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
