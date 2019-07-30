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
	LoggerLevel []string
	Port        int
	LogFileName string
	ConfigName  string
}

var configDataInstance *ConfigData

/*
GetConfigDataInstance get the confg data singleton
*/
func GetConfigDataInstance() *ConfigData {
	return configDataInstance
}

/*
GetConfigDataJSON string the configuration data as JSON. Used to record it in the logs
*/
func GetConfigDataJSON() string {
	return fmt.Sprintf("{\"configName\":\"%s\",\"port\":%d,\"logFileName\":\"%s\"}", configDataInstance.ConfigName, configDataInstance.Port, configDataInstance.LogFileName)
}

/*
LoadConfigData method loads the config data
*/
func LoadConfigData(configFileName string) error {

	if configFileName == "" {
		configFileName = "webServerBase.json"
	}

	configDataInstance = &ConfigData{
		Port: 8080,
	}
	/*
		load the config object
	*/
	err := jsonconfig.LoadJson(configFileName, &configDataInstance)
	if err != nil {
		return err
	}

	configDataInstance.ConfigName = configFileName
	return nil
}
