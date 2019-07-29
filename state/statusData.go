package state

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

/*
StatusData is the state of the server
*/
type statusData struct {
	startTime  string
	executable string
	state      string
}

var statusDataInstance *statusData
var statusOnce sync.Once

/*
InitStatusDataInstance create only ONE!
*/
func InitStatusDataInstance() {
	statusOnce.Do(func() {
		var parts []string
		var lastPart string
		serverName, _ := os.Executable()
		if strings.Contains(serverName, "/") {
			parts = strings.Split(serverName, "/")
		} else {
			parts = strings.Split(serverName, "\\")
		}
		if len(parts) < 2 {
			lastPart = serverName
		} else {
			lastPart = parts[len(parts)-1]
		}
		statusDataInstance = &statusData{
			startTime:  time.Now().Format("2006-01-02 15:04:05"),
			executable: lastPart,
			state:      "RUNNING",
		}
	})
}

/*
SetStatusDataState set the operational state of the server
*/
func SetStatusDataState(newState string) {
	if statusDataInstance == nil {
		InitStatusDataInstance()
	}
	statusDataInstance.state = newState
}

/*
GetStatusDataExecutableName returns the name os the exe file
*/
func GetStatusDataExecutableName() string {
	if statusDataInstance == nil {
		InitStatusDataInstance()
	}
	return statusDataInstance.executable
}

/*
GetStatusDataJSON return the server status (including config data) as a JSON String
*/
func GetStatusDataJSON() string {
	if statusDataInstance == nil {
		InitStatusDataInstance()
	}
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\", \"config\":%s}", statusDataInstance.state, statusDataInstance.startTime, statusDataInstance.executable, GetConfigDataJSON())
}

/*
GetStatusDataJSONWithoutConfig return the server status (excluding config data) as a JSON String
*/
func GetStatusDataJSONWithoutConfig() string {
	if statusDataInstance == nil {
		InitStatusDataInstance()
	}
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\"}", statusDataInstance.state, statusDataInstance.startTime, statusDataInstance.executable)
}
