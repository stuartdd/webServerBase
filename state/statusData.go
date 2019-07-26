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
GetStatusInstance create only ONE!
*/
func GetStatusInstance() *statusData {
	statusOnce.Do(func() {
		t := time.Now()
		serverName, _ := os.Executable()
		parts := strings.Split(serverName, "/")
		lastPart := parts[len(parts)-1]
		statusDataInstance = &statusData{
			startTime:  t.Format("2006-01-02 15:04:05"),
			executable: lastPart,
			state:      "RUNNING",
		}
	})
	return statusDataInstance
}

/*
SetStatusState set the operational state of the server
*/
func SetStatusState(newState string) {
	GetStatusInstance().state = newState
}

/*
GetStatusExecutableName returns the name os the exe file
*/
func GetStatusExecutableName() string {
	return GetStatusInstance().executable
}

/*
GetStatusJSON return the server status (including config data) as a JSON String
*/
func GetStatusJSON() string {
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\", \"config\":%s}", GetStatusInstance().state, GetStatusInstance().startTime, GetStatusInstance().executable, GetConfigJSON())
}

/*
GetStatusJSONWithoutConfig return the server status (excluding config data) as a JSON String
*/
func GetStatusJSONWithoutConfig() string {
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\"}", GetStatusInstance().state, GetStatusInstance().startTime, GetStatusInstance().executable)
}
