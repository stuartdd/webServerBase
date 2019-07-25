package state

import (
	"fmt"

	jsonconfig "github.com/stuartdd/tools_jsonconfig"
)

/*
ConfigData read cinfiguration data from the JSON configuration file.
Note any undefined values are defaulted to constants defined below
*/
type ConfigData struct {
	Debug       bool
	Port        int
	Timeout     int64
	LogFileName string
	ConfigName  string
}

/*
ToString string the configuration data. Used to record it in the logs
*/
func (p *ConfigData) ToString() string {
	return fmt.Sprintf("{\"configName\":\"%s\",\"port\":%d,\"timeout\":%d,\"logFileName\":\"%s\"}", p.ConfigName, p.Port, p.Timeout, p.LogFileName)
}

/*
LoadConfigData method loads the config data
*/
func LoadConfigData(configFileName string) (*ConfigData, error) {

	if configFileName == "" {
		configFileName = "webServerBase.json"
	}

	configData := ConfigData{
		Debug:   true,
		Port:    8080,
		Timeout: 1000,
	}
	/*
		load the config object
	*/
	err := jsonconfig.LoadJson(configFileName, &configData)
	if err != nil {
		return nil, err
	}

	configData.ConfigName = configFileName
	return &configData, nil
}
