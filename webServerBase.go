package main

import (
	"context"
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
var server *http.Server

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
	defer CloseLog()
	/*
		Add loggers for each module (Makes for neater logs!)
	*/
	logger = logging.NewLogger("BaseHandler")
	logger = logging.NewLogger("ServerMain")
	/*
		Log server startup info
	*/
	logger.LogInfof("Server will start on port %d\n", configData.Port)
	logger.LogInfof("To stop the server http://localhost:%d/stop\n", configData.Port)
	logger.LogDebugf("State:%s\n", state.GetStatusDataJSON())
	/*
	   Configure and Start the server.
	*/
	handlerData := handlers.NewHandlerData()
	handlerData.AddFileServer("/static/", "static")
	handlerData.AddBeforeHandler(filterBefore)
	handlerData.AddMappedHandler("/stop", http.MethodGet, stopHandler)
	handlerData.AddMappedHandler("/status", http.MethodGet, statusHandler)
	handlerData.AddAfterHandler(filterAfter)

	server = &http.Server{Addr: ":" + strconv.Itoa(configData.Port)}
	server.Handler = handlerData
	err := server.ListenAndServe()
	if err != nil {
		logger.LogInfo(err.Error())
	} else {
		logger.LogInfo("http: Server closed")
	}
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

func filterBefore(r *http.Request, response *dto.Response) *dto.Response {
	logger.LogDebug("IN Filter Before 1")
	return nil
}

func filterAfter(r *http.Request, response *dto.Response) *dto.Response {
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

/*
CloseLog closes the logger file if it exists

A logger os passed to enable the CloseLog function to log that fact it has been closed!
*/
func CloseLog() {
	logging.CloseLog(logger)
}

func stopServer(immediate bool) {
	if !immediate {
		time.Sleep(time.Millisecond * 500)
	}
	err := server.Shutdown(context.TODO())
	if err != nil {
		panic(err)
	}
}
