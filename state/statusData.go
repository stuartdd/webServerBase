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
	config     *ConfigData
	startTime  string
	executable string
	state      string
}

var statusInstance *statusData
var statusOnce sync.Once

/*
GetStatusInstance create only ONE!
*/
func GetStatusInstance() *statusData {
	statusOnce.Do(func() {
		t := time.Now()
		serverName, _ := os.Executable()
		parts := strings.Split(serverName, "/")
		lastPart := parts[len(parts)-1]
		statusInstance = &statusData{
			config:     nil,
			startTime:  t.Format("2006-01-02 15:04:05"),
			executable: lastPart,
			state:      "RUNNING",
		}
	})
	return statusInstance
}

/*
SetState set the operational state of the server
*/
func SetStatusState(newState string) {
	statusInstance.state = newState
}

/*
SetConfigData returns the configData data object
*/
func SetStatusConfigData(newConfig *ConfigData) {
	statusInstance.config = newConfig
}

/*
GetConfigData returns the configData data object
*/
func GetStatusConfigData() *ConfigData {
	return statusInstance.config
}

/*
GetExecutable returns the name os the exe file
*/
func GetStatusExecutableName() string {
	return statusInstance.executable
}

/*
GetStatusJSON return the server status (including config data) as a JSON String
*/
func GetStatusJSON() string {
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\", \"config\":%s}", statusInstance.state, statusInstance.startTime, statusInstance.executable, GetConfigJSON())
}

/*
GetStatusJSONWithoutConfig return the server status (excluding config data) as a JSON String
*/
func GetStatusJSONWithoutConfig() string {
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\"}", statusInstance.state, statusInstance.startTime, statusInstance.executable)
}
