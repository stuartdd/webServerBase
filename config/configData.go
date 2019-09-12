package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strings"
)

/*
Data - Read configuration data from the JSON configuration file.
Note any undefined values are defaulted to constants defined below
*/
type Data struct {
	Port               int
	DefaultLogFileName string
	ConfigName         string
	Redirections       map[string]string
	ContentTypes       map[string]string
	ContentTypeCharset string
	LoggerLevels       map[string]string
	PanicResponseCode  int
	StaticPaths        map[string]map[string]string
	TemplatePaths      map[string]string
}

/*
There should only ever be ONE of these
*/
var configDataInstance *Data

/*
LoadConfigData method loads the config data
*/
func LoadConfigData(configFileName string) error {

	if configFileName == "" {
		configFileName = "webServerBase.json"
	}

	configDataInstance = &Data{
		Port:               8080,
		ContentTypeCharset: "utf-8",
		ContentTypes:       make(map[string]string),
		StaticPaths:        make(map[string]map[string]string),
		TemplatePaths:      make(map[string]string),
		Redirections:       make(map[string]string),
		LoggerLevels:       make(map[string]string),
	}
	/*
		load the config object
	*/
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &configDataInstance)
	if err != nil {
		return err
	}

	configDataInstance.ConfigName = configFileName
	return nil
}

/*
GetOS returns the name of the operating system. Used to look up os
specific paths in config data staticPaths and templatePaths.
Use this in error messages to indicate a path is not found for the OS
*/
func GetOS() string {
	return runtime.GOOS
}

/*
GetConfigDataStaticFilePathForOS Get the static path for the OS. If not found return the first one!
*/
func (p *Data) GetConfigDataStaticFilePathForOS() map[string]string {
	path := p.StaticPaths[runtime.GOOS]
	if path == nil {
		log.Fatalf("Unable to find staticPath in configuration file '%s' for operating system '%s'", GetConfigDataInstance().ConfigName, runtime.GOOS)
	}
	return path
}

/*
GetConfigDataTemplateFilePathForOS Get the static path for the OS. If not found return the first one!
*/
func (p *Data) GetConfigDataTemplateFilePathForOS() string {
	path := p.TemplatePaths[runtime.GOOS]
	if path == "" {
		log.Fatalf("Unable to find templatePath in configuration file '%s' for operating system '%s'", GetConfigDataInstance().ConfigName, runtime.GOOS)
	}
	return path
}

/*
GetConfigDataInstance get the confg data singleton
*/
func GetConfigDataInstance() *Data {
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
