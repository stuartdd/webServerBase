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
func GetStatusDataInstance() *statusData {
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
func SetStatusDataState(newState string) {
	GetStatusDataInstance().state = newState
}

/*
GetStatusExecutableName returns the name os the exe file
*/
func GetStatusDataExecutableName() string {
	return GetStatusDataInstance().executable
}

/*
GetStatusJSON return the server status (including config data) as a JSON String
*/
func GetStatusDataJSON() string {
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\", \"config\":%s}", GetStatusDataInstance().state, GetStatusDataInstance().startTime, GetStatusDataInstance().executable, GetConfigDataJSON())
}

/*
GetStatusJSONWithoutConfig return the server status (excluding config data) as a JSON String
*/
func GetStatusDataJSONWithoutConfig() string {
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\"}", GetStatusDataInstance().state, GetStatusDataInstance().startTime, GetStatusDataInstance().executable)
}
