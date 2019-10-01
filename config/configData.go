package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	TemplateData       map[string]map[string]string
	ScriptData         map[string]*ScriptData
}

/*
ScriptData - For a goven OS define the scriptpath and the script data
*/
type ScriptData struct {
	Path string
	Data map[string][]string
}

/*
There should only ever be ONE of these
*/
var configDataInstance *Data

/*
GetConfigDataInstance get the confg data singleton
*/
func GetConfigDataInstance() *Data {
	return configDataInstance
}

/*
LoadConfigData method loads the config data from a file
*/
func LoadConfigData(configFileName string) error {

	if configFileName == "" {
		configFileName = GetApplicationModuleName() + ".json"
	}

	configDataInstance = &Data{
		Port:               8080,
		ContentTypeCharset: "utf-8",
		ContentTypes:       make(map[string]string),
		StaticPaths:        make(map[string]map[string]string),
		Redirections:       make(map[string]string),
		LoggerLevels:       make(map[string]string),
		TemplatePaths:      make(map[string]string),
		TemplateData:       make(map[string]map[string]string),
		ScriptData:         make(map[string]*ScriptData),
	}

	/*
		load the config object
	*/
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config data file:%s. Error:%s", configFileName, err.Error()))
	}

	err = json.Unmarshal(content, &configDataInstance)
	if err != nil {
		panic(fmt.Sprintf("Failed to understand the config data in the file:%s. Error:%s", configFileName, err.Error()))
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
	path := p.StaticPaths[GetOS()]
	if path == nil {
		log.Fatalf("Unable to find staticPath in configuration file '%s' for operating system '%s'", GetConfigDataInstance().ConfigName, runtime.GOOS)
	}
	return path
}

func (p *Data) GetScriptDataForOS() *ScriptData {
	sd := p.ScriptData[GetOS()]
	if sd == nil {
		log.Fatalf("Unable to find ScriptData in configuration file '%s' for operating system '%s'", GetConfigDataInstance().ConfigName, runtime.GOOS)
	}
	return sd
}

/*
GetConfigDataTemplateFilePathForOS Get the static path for the OS. If not found return the first one!
*/
func (p *Data) GetConfigDataTemplateFilePathForOS() string {
	path := p.TemplatePaths[GetOS()]
	if path == "" {
		log.Fatalf("Unable to find templatePath in configuration file '%s' for operating system '%s'", GetConfigDataInstance().ConfigName, runtime.GOOS)
	}
	return path
}

/*
GetApplicationModuleName returns the name of the application. Testing and debugging changes this name so the code
removes debug, test and .exe from the executable name.
*/
func GetApplicationModuleName() string {
	exec, err := os.Executable()
	if err != nil {
		exec = "UnknownModule"
	} else {
		parts := strings.Split(exec, fmt.Sprintf("%c", os.PathSeparator))
		exec = parts[len(parts)-1]
		if strings.Contains(strings.ToLower(exec), "debug.test") {
			exec = parts[len(parts)-2]
		}
	}
	if strings.HasSuffix(strings.ToLower(exec), ".exe") {
		exec = exec[0 : len(exec)-4]
	}
	if strings.HasSuffix(strings.ToLower(exec), ".test") {
		exec = exec[0 : len(exec)-5]
	}
	return exec
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
