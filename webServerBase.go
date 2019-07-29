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
		logging.Fatal(err)
		os.Exit(1)
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
	logging.CreateLogWithFilename(configData.LogFileName)
	defer logging.CloseLog()
	/*
	   Say hello.
	*/
	logging.Logf("Server will start on port %d\n", configData.Port)
	logging.Logf("Server will start on port %d\n", configData.Port)
	logging.Logf("To stop the server http://localhost:%d/stop\n", configData.Port)
	logging.Logf("Action:stop - Stop the server\n")
	logging.Logf("Action:status - Return server status\n")
	logging.Logf("Server:Configured\n")
	if configData.Debug {
		logging.Logf("State:%s\n", state.GetStatusDataJSON())
	}
	/*
	   Start the server.
	*/
	handlerData := handlers.NewHandlerData()
	handlerData.AddBeforeHandler(filterBefore)
	handlerData.AddMappedHandler("/stop", http.MethodGet, stopHandler)
	handlerData.AddMappedHandler("/status", http.MethodGet, statusHandler)
	handlerData.AddAfterHandler(filterAfter)
	handlerData.SetErrorHandler(errorHandler)
	logging.Fatal(http.ListenAndServe(":"+strconv.Itoa(configData.Port), handlerData))
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
	logging.Logf("IN Filter Before 1")
	return nil
}
func filterAfter(r *http.Request) *dto.Response {
	logging.Logf("IN Filter After 1")
	return nil
}
func errorHandler(w http.ResponseWriter, r *http.Request, e *dto.Response) {
	logging.Logf("IN errorHandler")
	http.Error(w, "bollocks", 400)
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
