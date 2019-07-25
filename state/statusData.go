package state

import (
	"fmt"
	"os"
	"strings"
	"time"
)

/*
StatusData is the state of the server
*/
type StatusData struct {
	config     *ConfigData
	startTime  string
	executable string
	state      string
}

/*
SetState set the operational state of the server
*/
func (p *StatusData) SetState(newState string) {
	p.state = newState
}

/*
GetConfigData returns the configData data object
*/
func (p *StatusData) GetConfigData() *ConfigData {
	return p.config
}

/*
GetExecutable returns the name os the exe file
*/
func (p *StatusData) GetExecutable() string {
	return p.executable
}

/*
ToString return the server status (including config data) as a JSON String
*/
func (p *StatusData) ToString() string {
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\", \"config\":%s}", p.state, p.startTime, p.executable, p.config.ToString())
}

/*
ToStringNoConfig return the server status (excluding config data) as a JSON String
*/
func (p *StatusData) ToStringWithoutConfig() string {
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\"}", p.state, p.startTime, p.executable)
}

/*
NewStatusData creation
*/
func NewStatusData(configData *ConfigData) *StatusData {
	t := time.Now()
	serverName, _ := os.Executable()
	parts := strings.Split(serverName, "/")
	lastPart := parts[len(parts)-1]
	return &StatusData{
		config:     configData,
		startTime:  t.Format("2006-01-02 15:04:05"),
		executable: lastPart,
		state:      "RUNNING",
	}
}
