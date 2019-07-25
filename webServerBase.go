package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"webServerBase/handlers"
	"webServerBase/state"
)

var statusData *state.StatusData
var logFile *os.File

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
	configData, err := state.LoadConfigData(configFileName)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	RunWithConfig(configData)
}

/*

RunWithConfig runs with a sgeneric handler
Param - config a ref to the config object
*/
func RunWithConfig(configData *state.ConfigData) {

	statusData = state.NewStatusData(configData)
	/*
		Open the logs. Log name is in the congig data. If not defined default to sysout
	*/
	createLog(configData)
	defer closeLog()
	/*
	   Say hello.
	*/
	log.Printf("Server will start on port %d\n", configData.Port)
	log.Printf("To stop the server http://localhost:%d/stop\n", configData.Port)
	log.Printf("Action:stop - Stop the server\n")
	log.Printf("Action:status - Return server status\n")
	log.Printf("Server:Configured\n")
	if configData.Debug {
		log.Printf("State:%s\n", statusData.ToString())
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
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(configData.Port), handlerData))
}

/************************************************
Start of handlers section
*************************************************/

func stopHandler(w http.ResponseWriter, r *http.Request) *handlers.ErrorResponse {
	statusData.SetState("STOPPING")
	go stopServer(false)
	return logAndReturn200(w, r, statusData.ToStringWithoutConfig())
}

func statusHandler(w http.ResponseWriter, r *http.Request) *handlers.ErrorResponse {
	return logAndReturn200(w, r, statusData.ToString())
}

func filterBefore(w http.ResponseWriter, r *http.Request) *handlers.ErrorResponse {
	log.Print("IN Filter Before 1")
	return handlers.NewErrorResponse(500, "", nil)
}
func filterAfter(w http.ResponseWriter, r *http.Request) *handlers.ErrorResponse {
	log.Print("IN Filter After 1")
	return nil
}
func errorHandler(w http.ResponseWriter, r *http.Request, e *handlers.ErrorResponse) {
	log.Print("IN errorHandler")
	http.Error(w, "bollocks", 400)
}
func filterAfter2(w http.ResponseWriter, r *http.Request) *handlers.ErrorResponse {
	log.Print("IN Filter After 2")
	return nil
}

/************************************************
End of handlers section

Start of utility functions
*************************************************/
func logAndReturn200(w http.ResponseWriter, r *http.Request, resp string) *handlers.ErrorResponse {
	log.Print("REQ:" + r.URL.Path + " RESP:" + resp)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Server", statusData.GetExecutable())
	fmt.Fprintf(w, resp)
	return nil
}

func createLog(configData *state.ConfigData) {
	if configData.LogFileName != "" {
		f, err := os.OpenFile(configData.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Log file '%s' could NOT be opened\nError:%s", configData.LogFileName, err.Error())
			return
		}
		logFile = f
		log.SetOutput(logFile)
	}
}

func closeLog() {
	if logFile != nil {
		logFile.Close()
	}
}

func stopServer(immediate bool) {
	if !immediate {
		time.Sleep(time.Millisecond * 500)
	}
	closeLog()
	os.Exit(0)
}
