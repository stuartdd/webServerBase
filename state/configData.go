package state

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	jsonconfig "github.com/stuartdd/tools_jsonconfig"
)

/*
ConfigData read configuration data from the JSON configuration file.
Note any undefined values are defaulted to constants defined below
*/
type ConfigData struct {
	Port                   int
	DefaultLogFileName     string
	ConfigName             string
	StaticPaths            map[string]map[string]string
	TemplatePaths          map[string]map[string]string
	Redirections           map[string]string
	ContentTypes           map[string]string
	ContentTypeCharset     string
	LoggerLevels           map[string]string
	PanicResponseCode      int
	NoResponseResponseCode int
}

var configDataInstance *ConfigData

/*
LoadConfigData method loads the config data
*/
func LoadConfigData(configFileName string) error {

	if configFileName == "" {
		configFileName = "webServerBase.json"
	}

	configDataInstance = &ConfigData{
		Port:               8080,
		ContentTypeCharset: "utf-8",
		ContentTypes:       make(map[string]string),
		StaticPaths:        make(map[string]map[string]string),
		TemplatePaths:      make(map[string]map[string]string),
		Redirections:       make(map[string]string),
		LoggerLevels:       make(map[string]string),
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

/*
GetConfigDataStaticPathForOS Get the static path for the OS. If not found return the first one!
*/
func GetConfigDataStaticPathForOS() map[string]string {
	path := GetConfigDataInstance().StaticPaths[runtime.GOOS]
	if path == nil {
		log.Fatalf("Unable to find staticPath in configuration file '%s' for operating system '%s'", GetConfigDataInstance().ConfigName, runtime.GOOS)
	}
	return path
}

/*
GetConfigDataTemplatePathForOS Get the static path for the OS. If not found return the first one!
*/
func GetConfigDataTemplatePathForOS() map[string]string {
	path := GetConfigDataInstance().TemplatePaths[runtime.GOOS]
	if path == nil {
		log.Fatalf("Unable to find templatePath in configuration file '%s' for operating system '%s'", GetConfigDataInstance().ConfigName, runtime.GOOS)
	}
	return path
}

/*
GetConfigDataInstance get the confg data singleton
*/
func GetConfigDataInstance() *ConfigData {
	return configDataInstance
}

/*
GetConfigDataJSON string the 'usefull' configuration data as JSON. Used to record it in the logs
*/
func GetConfigDataJSON() string {
	return fmt.Sprintf("{\"configName\":\"%s\",\"port\":%d,\"logFileName\":\"%s\",\"LoggerLevel\":%s}",
		configDataInstance.ConfigName,
		configDataInstance.Port,
		configDataInstance.DefaultLogFileName,
		toStringMap(configDataInstance.LoggerLevels))
}

func toStringMap(mapIn map[string]string) string {
	out := "{"
	ind := len(out)
	for key, value := range mapIn {
		value = strings.ReplaceAll(value, "\\", "\\\\")
		out = out + "\"" + key + "\":\"" + value + "\""
		ind = len(out)
		out = out + ", "
	}
	return string(out[0:ind]) + "}"
}
