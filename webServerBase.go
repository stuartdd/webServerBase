package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"webServerBase/handlers"
	"webServerBase/logging"
	"webServerBase/state"
)

var logger *logging.LoggerDataReference
var server *handlers.HandlerFunctionData

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
	err := state.LoadConfigData(configFileName)
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

	RunWithConfig(state.GetConfigDataInstance(), exec)
}

/*

RunWithConfig runs with a sgeneric handler
Param - config a ref to the config object
*/
func RunWithConfig(configData *state.ConfigData, executable string) {
	/*
		Open the logs. Log name is in the congig data. If not defined default to sysout
	*/
	logging.CreateLogWithFilenameAndAppID(state.GetConfigDataInstance().DefaultLogFileName, executable+":"+strconv.Itoa(state.GetConfigDataInstance().Port), state.GetConfigDataInstance().LoggerLevels)
	defer CloseLog()
	/*
		Add loggers for each module (Makes for neater logs!)
	*/
	logger = logging.NewLogger("ServerMain")
	/*
		Log server startup info
	*/
	logger.LogInfof("Server will start on port %d\n", configData.Port)
	logger.LogInfof("OS '%s'. Static path will be:%s\n", runtime.GOOS, state.GetConfigDataStaticPathForOS())
	logger.LogInfof("To stop the server http://localhost:%d/stop\n", configData.Port)
	logger.LogDebugf("State:%s\n", state.GetStatusDataJSON())
	/*
	   Configure and Start the server.
	*/
	server = handlers.NewHandlerData(executable, configData.ContentTypeCharset)
	server.AddContentTypeFromMap(configData.ContentTypes)
	server.SetRedirections(state.GetConfigDataInstance().Redirections)
	server.AddFileServerDataFromMap(state.GetConfigDataStaticPathForOS())
	server.AddBeforeHandler(filterBefore)
	server.AddMappedHandler("/stop", http.MethodGet, stopHandler)
	server.AddMappedHandler("/status", http.MethodGet, statusHandler)
	server.AddMappedHandler("/calc/?/add/?", http.MethodGet, adderHandler)
	server.AddAfterHandler(filterAfter)
	server.ListenAndServeOnPort(configData.Port)
}

/************************************************
Start of handlers section
*************************************************/

func stopHandler(r *http.Request) *handlers.Response {
	state.SetStatusDataState("STOPPING")
	go server.StopServer(false)
	return handlers.NewResponse(200, state.GetStatusDataJSONWithoutConfig(), "application/json", nil)
}

func adderHandler(r *http.Request) *handlers.Response {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	a, err := strconv.Atoi(parts[1])
	if err != nil {
		return handlers.NewResponse(400, "invalid number "+parts[1], "", err)
	}
	b, err := strconv.Atoi(parts[3])
	if err != nil {
		return handlers.NewResponse(400, "invalid number "+parts[3], "", err)
	}
	return handlers.NewResponse(200, strconv.Itoa(a+b), "", nil)
}

func statusHandler(r *http.Request) *handlers.Response {
	return handlers.NewResponse(200, state.GetStatusDataJSON(), "application/json", nil)
}

func filterBefore(r *http.Request, response *handlers.Response) *handlers.Response {
	logger.LogDebug("IN Filter Before 1")
	return nil
}

func filterAfter(r *http.Request, response *handlers.Response) *handlers.Response {
	logger.LogDebug("IN Filter After 1")
	return nil
}

func errorHandler(w http.ResponseWriter, r *http.Request, e *handlers.Response) {
	logger.LogDebug("IN errorHandler")
	http.Error(w, e.GetResp(), 400)
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
	logging.CloseLog()
}
