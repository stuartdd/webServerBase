package state

import (
	"fmt"
	"strings"

	jsonconfig "github.com/stuartdd/tools_jsonconfig"
)

/*
LoggerLevelData defines whicl log levels are active and their output{
*/
type LoggerLevelData struct {
	Level string
	File  string
}

/*
ConfigData read configuration data from the JSON configuration file.
Note any undefined values are defaulted to constants defined below
*/
type ConfigData struct {
	LoggerLevels       []LoggerLevelData
	Port               int
	LogFileName        string
	ConfigName         string
	StaticPaths        map[string]string
	ContentTypes       map[string]string
	ContentTypeCharset string
}

var configDataInstance *ConfigData

func (p *LoggerLevelData) String() string {
	return fmt.Sprintf("{\"level\":\"%s\", \"file\":\"%s\"}", p.Level, p.File)
}

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
	return fmt.Sprintf("{\"configName\":\"%s\",\"port\":%d,\"logFileName\":\"%s\",\"LoggerLevel\":%s,\"staticPath\":%s}",
		configDataInstance.ConfigName,
		configDataInstance.Port,
		configDataInstance.LogFileName,
		toStringList(configDataInstance.LoggerLevels),
		toStringMap(configDataInstance.StaticPaths))
}

func toStringList(list []LoggerLevelData) string {
	out := "["
	ind := len(out)
	for _, element := range list {
		out = out + element.String()
		ind = len(out)
		out = out + ", "
	}
	return string(out[0:ind]) + "]"
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
	}
	/*
		load the config object
	*/
	err := jsonconfig.LoadJson(configFileName, &configDataInstance)
	if err != nil {
		return err
	}

	addContentTypes()
	configDataInstance.ConfigName = configFileName
	return nil
}

/*
from : https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Complete_list_of_MIME_types
*/
func addContentTypes() {
	doNotOverwriteContentTypes("aac", "audio/aac")
	doNotOverwriteContentTypes("abw", "application/x-abiword")
	doNotOverwriteContentTypes("arc", "application/x-freearc")
	doNotOverwriteContentTypes("avi", "video/x-msvideo")
	doNotOverwriteContentTypes("azw", "application/vnd.amazon.ebook")
	doNotOverwriteContentTypes("bin", "application/octet-stream")
	doNotOverwriteContentTypes("bmp", "image/bmp")
	doNotOverwriteContentTypes("bz", "application/x-bzip")
	doNotOverwriteContentTypes("bz2", "application/x-bzip2")
	doNotOverwriteContentTypes("csh", "application/x-csh")
	doNotOverwriteContentTypes("css", "text/css")
	doNotOverwriteContentTypes("csv", "text/csv")
	doNotOverwriteContentTypes("doc", "application/msword")
	doNotOverwriteContentTypes("docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	doNotOverwriteContentTypes("eot", "application/vnd.ms-fontobject")
	doNotOverwriteContentTypes("epub", "application/epub+zip")
	doNotOverwriteContentTypes("gif", "image/gif")
	doNotOverwriteContentTypes("htm", "text/html")
	doNotOverwriteContentTypes("html", "text/html")
	doNotOverwriteContentTypes("ico", "image/vnd.microsoft.icon")
	doNotOverwriteContentTypes("ics", "text/calendar")
	doNotOverwriteContentTypes("jar", "application/java-archive")
	doNotOverwriteContentTypes("jpeg", "image/jpeg")
	doNotOverwriteContentTypes("jpg", "image/jpeg")
	doNotOverwriteContentTypes("js", "text/javascript")
	doNotOverwriteContentTypes("json", "application/json")
	doNotOverwriteContentTypes("jsonld", "application/ld+json")
	doNotOverwriteContentTypes("mid", "audio/midi audio/x-midi")
	doNotOverwriteContentTypes("midi", "audio/midi audio/x-midi")
	doNotOverwriteContentTypes("mjs", "text/javascript")
	doNotOverwriteContentTypes("mp3", "audio/mpeg")
	doNotOverwriteContentTypes("mpeg", "video/mpeg")
	doNotOverwriteContentTypes("mpkg", "application/vnd.apple.installer+xml")
	doNotOverwriteContentTypes("odp", "application/vnd.oasis.opendocument.presentation")
	doNotOverwriteContentTypes("ods", "application/vnd.oasis.opendocument.spreadsheet")
	doNotOverwriteContentTypes("odt", "application/vnd.oasis.opendocument.text")
	doNotOverwriteContentTypes("oga", "audio/ogg")
	doNotOverwriteContentTypes("ogv", "video/ogg")
	doNotOverwriteContentTypes("ogx", "application/ogg")
	doNotOverwriteContentTypes("otf", "font/otf")
	doNotOverwriteContentTypes("png", "image/png")
	doNotOverwriteContentTypes("pdf", "application/pdf")
	doNotOverwriteContentTypes("ppt", "application/vnd.ms-powerpoint")
	doNotOverwriteContentTypes("pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	doNotOverwriteContentTypes("rar", "application/x-rar-compressed")
	doNotOverwriteContentTypes("rtf", "application/rtf")
	doNotOverwriteContentTypes("sh", "application/x-sh")
	doNotOverwriteContentTypes("svg", "image/svg+xml")
	doNotOverwriteContentTypes("swf", "application/x-shockwave-flash")
	doNotOverwriteContentTypes("tar", "application/x-tar")
	doNotOverwriteContentTypes("tif", "image/tiff")
	doNotOverwriteContentTypes("tiff", "image/tiff")
	doNotOverwriteContentTypes("ts", "video/mp2t")
	doNotOverwriteContentTypes("ttf", "font/ttf")
	doNotOverwriteContentTypes("txt", "text/plain")
	doNotOverwriteContentTypes("vsd", "application/vnd.visio")
	doNotOverwriteContentTypes("wav", "audio/wav")
	doNotOverwriteContentTypes("weba", "audio/webm")
	doNotOverwriteContentTypes("webm", "video/webm")
	doNotOverwriteContentTypes("webp", "image/webp")
	doNotOverwriteContentTypes("woff", "font/woff")
	doNotOverwriteContentTypes("woff2", "font/woff2")
	doNotOverwriteContentTypes("xhtml", "application/xhtml+xml")
	doNotOverwriteContentTypes("xls", "application/vnd.ms-excel")
	doNotOverwriteContentTypes("xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	doNotOverwriteContentTypes("xml", "application/xml")
	doNotOverwriteContentTypes("xul", "application/vnd.mozilla.xul+xml")
	doNotOverwriteContentTypes("zip", "application/zip")
	doNotOverwriteContentTypes("7z", "application/x-7z-compressed")
}

func doNotOverwriteContentTypes(mimeType string, mime string) {
	_, found := configDataInstance.ContentTypes[mimeType]
	if !found {
		configDataInstance.ContentTypes[mimeType] = mime
	}
}
