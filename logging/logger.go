package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"webServerBase/state"
)

/*
LoggerLevelType ENUM for log levels
*/
type LoggerLevelType int

/*
InfoLevel is the finest. Nothing stops ErrorLevel of FatalLevel
*/
const (
	InfoLevel LoggerLevelType = iota
	DebugLevel
	WarnLevel
	AccessLevel
	ErrorLevel
	FatalLevel
	NotFound
)

/*
loggerLevelData One instance per Log Level
*/
type loggerLevelData struct {
	active bool
	logger *log.Logger
	file   *loggerFileData
}

type loggerFileData struct {
	fileName string
	logFile  *os.File
}

/*
LoggerDataReference contains a ref to th esingle logger instance and the module name (id).

Created via NewLogger
*/
type LoggerDataReference struct {
	loggerModuleName string
	loggerPrefix     string
}

/*
These names should be ALL the same length and should have a ' ' before AND after the name
*/
var loggerLevelTypeNames = [...]string{"  INFO ", " DEBUG ", "  WARN ", "ACCESS ", " ERROR ", " FATAL "}

/*
These values (not case sensitive) must map to the values passed to CreateLogWithFilenameAndAppID.
If these values are in the list then that log level will be active.
An empty list will mean that only ERROR and FATAL will be logged
*/
var loggerLevelMapNames = map[string]LoggerLevelType{"INFO": InfoLevel, "DEBUG": DebugLevel, "WARN": WarnLevel, "ACCESS": AccessLevel, "ERROR": ErrorLevel, "FATAL": FatalLevel}

/*
For each logger level there MAY be a file. Indexed by file name. This is so we can re-use the file with the same name for different levels
*/
var loggerLevelFiles map[string]*loggerFileData
var logDataModules map[string]*LoggerDataReference
var loggerLevelDataList = initLoggerLevelDataList()

var longestModuleName int

var logDataFlags int
var logApplicationID string
var logFileNameGlobal string
var fallBack = true

/*
CreateLogWithFilenameAndAppID should configure the logger to output somthing like this!
2019-07-16 14:47:43.993 applicationID module  [-]  INFO Starti
2019-07-16 14:47:43.993 applicationID module  [-] DEBUG Runnin
*/
func CreateLogWithFilenameAndAppID(logFileName string, applicationID string, config []state.LoggerLevelData) {
	CloseLog()
	logFileNameGlobal = logFileName
	logApplicationID = applicationID
	logDataFlags = log.LstdFlags | log.Lmicroseconds
	logDataModules = make(map[string]*LoggerDataReference)
	loggerLevelFiles = make(map[string]*loggerFileData)
	loggerLevelDataList = initLoggerLevelDataList()
	longestModuleName = 0
	/*
		Validate and Activate each log level.
	*/
	validateAndActivateLogLevels(config)
	fallBack = false
}

func initLoggerLevelDataList() []*loggerLevelData {
	return []*loggerLevelData{
		newLoggerLevelTypeData(false),
		newLoggerLevelTypeData(false),
		newLoggerLevelTypeData(false),
		newLoggerLevelTypeData(false),
		newLoggerLevelTypeData(true),
		newLoggerLevelTypeData(true)}
}

/*
GetLogLevelTypeForName get the index for the level name
*/
func GetLogLevelTypeForName(name string) LoggerLevelType {
	if index, ok := loggerLevelMapNames[strings.ToUpper(strings.TrimSpace(name))]; ok {
		return index
	}
	return NotFound
}

/*
GetLogLevelFileName get the index for the level name
*/
func GetLogLevelFileName(name string) string {
	loggerLevelType := GetLogLevelTypeForName(name)
	if loggerLevelType != NotFound {
		typeInstance := loggerLevelDataList[loggerLevelType]
		if typeInstance.file != nil {
			return typeInstance.file.fileName
		}
	}
	return ""
}

/*
LogPanicToStdErrAndExit - Last resort!
This creates a logger for System Error channel and use it to log.Fatal.
It then exits the application with a return code of 1
*/
func LogPanicToStdErrAndExit(message string) {
	log.Panic(message)
	os.Exit(1)
}

/*
LoggerLevelDataString return the state of a log level as a string
*/
func LoggerLevelDataString(name string) string {
	loggerLevelType := GetLogLevelTypeForName(name)
	if loggerLevelType != NotFound {
		lld := loggerLevelDataList[loggerLevelType]
		if lld.active {
			active := name + ":Active:"
			if lld.file == nil {
				return active + "Out=SysOut:"
			}
			active = active + "Out=:" + filepath.Base(lld.file.fileName)
			if lld.file.logFile == nil {
				return active + ":Closed"
			}
			return active + ":Open"

		}
		return name + ":In-Active"
	}
	return name + ":Not Found"
}

/*
CloseLog close the log file
*/
func CloseLog() {
	for _, value := range loggerLevelDataList {
		if value.file != nil {
			value.file.logFile.Close()
			value.active = false
		}
	}
}

/*
NewLogger created a new logger instance for a specific module
All log lines printed via the returned ref will contain the specific module name.
*/
func NewLogger(moduleName string) *LoggerDataReference {
	if val, ok := logDataModules[moduleName]; ok {
		return val
	}
	ldRef := &LoggerDataReference{
		loggerModuleName: moduleName,
		loggerPrefix:     logApplicationID,
	}
	logDataModules[moduleName] = ldRef
	updateLoggerPrefixesForAllModules()
	return ldRef
}

/*
IsDebug return true is the debug log function is enabled
*/
func (p *LoggerDataReference) IsDebug() bool {
	return loggerLevelDataList[DebugLevel].active
}

/*
IsAccess return true is the access log function is enabled
*/
func (p *LoggerDataReference) IsAccess() bool {
	return loggerLevelDataList[AccessLevel].active
}

/*
IsInfo return true is the info log function is enabled
*/
func (p *LoggerDataReference) IsInfo() bool {
	return loggerLevelDataList[InfoLevel].active
}

/*
IsWarn return true is the info log function is enabled
*/
func (p *LoggerDataReference) IsWarn() bool {
	return loggerLevelDataList[WarnLevel].active
}

/*
Fatal does the same as log.Fatal
*/
func (p *LoggerDataReference) Fatal(err error) {
	if fallBack {
		fmt.Printf("FATAL: [%T] %s", err, err.Error())
	} else {
		loggerLevelDataList[FatalLevel].logger.Printf(p.loggerPrefix+"[%s] %T %s", loggerLevelTypeNames[FatalLevel], err, err.Error())
	}
	os.Exit(1)
}

/*
LogErrorf delegates to log.Printf
*/
func (p *LoggerDataReference) LogErrorf(format string, v ...interface{}) {
	loggerLevelDataList[ErrorLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[ErrorLevel]+format, v...)
}

/*
LogError delegates to log.Print
*/
func (p *LoggerDataReference) LogError(message string) {
	loggerLevelDataList[ErrorLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[ErrorLevel] + message)
}

/*
LogInfof delegates to log.Printf
*/
func (p *LoggerDataReference) LogInfof(format string, v ...interface{}) {
	if loggerLevelDataList[InfoLevel].active {
		loggerLevelDataList[InfoLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[InfoLevel]+format, v...)
	}
}

/*
LogInfo delegates to log.Print
*/
func (p *LoggerDataReference) LogInfo(message string) {
	if loggerLevelDataList[InfoLevel].active {
		loggerLevelDataList[InfoLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[InfoLevel] + message)
	}
}

/*
LogAccessf delegates to log.Printf
*/
func (p *LoggerDataReference) LogAccessf(format string, v ...interface{}) {
	if loggerLevelDataList[AccessLevel].active {
		loggerLevelDataList[AccessLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[AccessLevel]+format, v...)
	}
}

/*
LogAccess delegates to log.Print
*/
func (p *LoggerDataReference) LogAccess(message string) {
	if loggerLevelDataList[AccessLevel].active {
		loggerLevelDataList[AccessLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[AccessLevel] + message)
	}
}

/*
LogWarnf delegates to log.Printf
*/
func (p *LoggerDataReference) LogWarnf(format string, v ...interface{}) {
	if loggerLevelDataList[WarnLevel].active {
		loggerLevelDataList[WarnLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[WarnLevel]+format, v...)
	}
}

/*
LogWarn delegates to log.Print
*/
func (p *LoggerDataReference) LogWarn(message string) {
	if loggerLevelDataList[WarnLevel].active {
		loggerLevelDataList[WarnLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[WarnLevel] + message)
	}
}

/*
LogDebugf delegates to log.Printf
*/
func (p *LoggerDataReference) LogDebugf(format string, v ...interface{}) {
	if loggerLevelDataList[DebugLevel].active {
		loggerLevelDataList[DebugLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[DebugLevel]+format, v...)
	}
}

/*
LogDebug delegates to log.Print
*/
func (p *LoggerDataReference) LogDebug(message string) {
	if loggerLevelDataList[DebugLevel].active {
		loggerLevelDataList[DebugLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[DebugLevel] + message)
	}
}

/*
	Validate and Activate each log level.
*/
func validateAndActivateLogLevels(values []state.LoggerLevelData) {
	/*
		For each log level definition
	*/
	for _, loggerLevelData := range values {
		/*
			check the name is valid
		*/
		loggerLevelType := GetLogLevelTypeForName(loggerLevelData.Level)
		if loggerLevelType != NotFound {
			filedata := getLoggerWithFilename(loggerLevelData.File)
			loggerLevelDataList[loggerLevelType].file = filedata
			if filedata == nil {
				loggerLevelDataList[loggerLevelType].logger = log.New(os.Stdout, "", logDataFlags)
			} else {
				loggerLevelDataList[loggerLevelType].logger = log.New(filedata.logFile, "", logDataFlags)
			}
			loggerLevelDataList[loggerLevelType].active = true
		} else {
			list := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(loggerLevelTypeNames)), ", "), "[]")
			LogPanicToStdErrAndExit("The Log level name '" + loggerLevelData.Level + "' is not a valid log level. Valid values are:" + list)
		}
	}
	if logFileNameGlobal != "" {
		for _, lld := range loggerLevelDataList {
			if lld.active && (lld.file == nil) {
				filedata := getLoggerWithFilename(logFileNameGlobal)
				lld.file = filedata
				if filedata == nil {
					lld.logger = log.New(os.Stdout, "", logDataFlags)
				} else {
					lld.logger = log.New(filedata.logFile, "", logDataFlags)
				}

			}
		}
	}
}

func updateLoggerPrefixesForAllModules() {
	longestName := 0
	for _, value := range logDataModules {
		length := len(value.loggerModuleName)
		if longestName < length {
			longestName = length
		}
	}
	for _, value := range logDataModules {
		value.loggerPrefix = logApplicationID + " " + (value.loggerModuleName + strings.Repeat(" ", longestName-len(value.loggerModuleName))) + " [-] "
	}
}

func newLoggerLevelTypeData(active bool) *loggerLevelData {
	return &loggerLevelData{
		active: active,
		logger: nil,
		file:   nil,
	}
}

func getLoggerWithFilename(logFileName string) *loggerFileData {
	nameUcTrim := strings.TrimSpace(strings.ToUpper(logFileName))
	if nameUcTrim == "" {
		return nil
	}
	if val, ok := loggerLevelFiles[nameUcTrim]; ok {
		return val
	}
	absFileName, err := filepath.Abs(logFileName)
	if err != nil {
		LogPanicToStdErrAndExit("applicationID " + logApplicationID + ". Log file " + logFileName + " is not a valid path:" + err.Error())
	}
	f, err := os.OpenFile(absFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		LogPanicToStdErrAndExit("applicationID " + logApplicationID + ". Log file " + logFileName + " could NOT be Created or Opened\nError:" + err.Error())
	}
	lfd := &loggerFileData{
		fileName: absFileName,
		logFile:  f,
	}
	loggerLevelFiles[nameUcTrim] = lfd
	return lfd
}
