package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"webServerBase/dto"
	"webServerBase/handlers"
	"webServerBase/logging"
	"webServerBase/state"
)

var logger *logging.LoggerDataReference

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
	RunWithConfig(state.GetConfigDataInstance())
}

/*

RunWithConfig runs with a sgeneric handler
Param - config a ref to the config object
*/
func RunWithConfig(configData *state.ConfigData) {
	/*
		Open the logs. Log name is in the congig data. If not defined default to sysout
	*/
	logging.CreateLogWithFilenameAndAppID(configData.LogFileName, state.GetStatusDataExecutableName()+":"+strconv.Itoa(state.GetConfigDataInstance().Port), state.GetConfigDataInstance().LoggerLevel)
	defer logging.CloseLog()
	logger = logging.NewLogger("WebServerBase")
	/*
	   Say hello.
	*/
	logger.LogInfof("Server will start on port %d\n", configData.Port)
	logger.LogInfof("To stop the server http://localhost:%d/stop\n", configData.Port)
	if configData.Debug {
		logger.LogDebugf("State:%s\n", state.GetStatusDataJSON())
	}
	/*
	   Configure and Start the server.
	*/
	handlerData := handlers.NewHandlerData()
	handlerData.AddBeforeHandler(filterBefore)
	handlerData.AddMappedHandler("/stop", http.MethodGet, stopHandler)
	handlerData.AddMappedHandler("/status", http.MethodGet, statusHandler)
	handlerData.AddAfterHandler(filterAfter)
	handlerData.SetErrorHandler(errorHandler)
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(configData.Port), handlerData))
}

/************************************************
Start of handlers section
*************************************************/

func stopHandler(r *http.Request) *dto.Response {
	state.SetStatusDataState("STOPPING")
	go stopServer(false)
	return dto.NewResponse(200, state.GetStatusDataJSONWithoutConfig(), "application/json", nil)
}

func statusHandler(r *http.Request) *dto.Response {
	return dto.NewResponse(200, state.GetStatusDataJSON(), "application/json", nil)
}

func filterBefore(r *http.Request) *dto.Response {
	logger.LogDebug("IN Filter Before 1")
	return nil
}
func filterAfter(r *http.Request) *dto.Response {
	logger.LogDebug("IN Filter After 1")
	return nil
}
func errorHandler(w http.ResponseWriter, r *http.Request, e *dto.Response) {
	logger.LogDebug("IN errorHandler")
	http.Error(w, e.GetResp(), 400)
}

/************************************************
End of handlers section

Start of utility functions
*************************************************/

func stopServer(immediate bool) {
	if !immediate {
		time.Sleep(time.Millisecond * 500)
	}
	logging.CloseLog()
	os.Exit(0)
}
