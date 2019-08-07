package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
	"webServerBase/state"
)

type loggerLevelType int

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
InfoLevel is the finest. Nothing stops ErrorLevel of FatalLevel
*/
const (
	InfoLevel loggerLevelType = iota
	DebugLevel
	WarnLevel
	AccessLevel
	ErrorLevel
	FatalLevel
)

/*
These names should be ALL the same length and should have a ' ' before AND after the name
*/
var loggerLevelTypeNames = [...]string{"  INFO ", " DEBUG ", "  WARN ", "ACCESS ", " ERROR ", " FATAL "}

/*
LoggerLevelDataList data structure for each level
*/
var LoggerLevelDataList = [...]*loggerLevelData{newLoggerLevelTypeData(false), newLoggerLevelTypeData(false), newLoggerLevelTypeData(false), newLoggerLevelTypeData(false), newLoggerLevelTypeData(true), newLoggerLevelTypeData(true)}

/*
These values (not case sensitive) must map to the values passed to CreateLogWithFilenameAndAppID.
If these values are in the list then that log level will be active.
An empty list will mean that only ERROR and FATAL will be logged
*/
var loggerLevelMapNames = map[string]loggerLevelType{"INFO": InfoLevel, "DEBUG": DebugLevel, "WARN": WarnLevel, "ACCESS": AccessLevel, "ERROR": ErrorLevel, "FATAL": FatalLevel}

/*
For each logger level there MAY be a file. Indexed by file name. This is so we can re-use the file with the same name for different levels
*/
var loggerLevelFiles = make(map[string]*loggerFileData)

var longestModuleName int = 0

/*
LoggerDataReference contains a ref to th esingle logger instance and the module name (id).

Created via NewLogger
*/
type LoggerDataReference struct {
	loggerModuleName string
	loggerPrefix     string
}

var logDataModules map[string]*LoggerDataReference
var logDataFlags int
var logApplicationID string
var logFileNameGlobal string

/*
CreateLogWithFilenameAndAppID should configure the logger to output somthing like this!
2019-07-16 14:47:43.993 applicationID module  [-]  INFO Starti
2019-07-16 14:47:43.993 applicationID module  [-] DEBUG Runnin
*/
func CreateLogWithFilenameAndAppID(logFileName string, applicationID string, config []state.LoggerLevelData) {
	logFileNameGlobal = logFileName
	logApplicationID = applicationID
	logDataFlags = log.LstdFlags | log.Lmicroseconds
	logDataModules = make(map[string]*LoggerDataReference)
	/*
		Validate and Activate each log level.
	*/
	validateAndActivateLogLevels(config)
}

/*
CloseLog close the log file
*/
func CloseLog(logger *LoggerDataReference) {
	// if logDataInstance.logFile != nil {
	// 	logger.LogInfof("logging.CloseLog: Log file %s is closing", logDataInstance.fileName)
	// 	logDataInstance.logFile.Close()
	// 	logDataInstance.logFile = nil
	// } else {
	// 	logger.LogWarn("logging.CloseLog: Was called but there is NO log file open")
	// }
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

/*
IsDebug return true is the debug log function is enabled
*/
func (p *LoggerDataReference) IsDebug() bool {
	return LoggerLevelDataList[DebugLevel].active
}

/*
IsAccess return true is the access log function is enabled
*/
func (p *LoggerDataReference) IsAccess() bool {
	return LoggerLevelDataList[AccessLevel].active
}

/*
IsInfo return true is the info log function is enabled
*/
func (p *LoggerDataReference) IsInfo() bool {
	return LoggerLevelDataList[InfoLevel].active
}

/*
IsWarn return true is the info log function is enabled
*/
func (p *LoggerDataReference) IsWarn() bool {
	return LoggerLevelDataList[WarnLevel].active
}

/*
Fatal does the same as log.Fatal
*/
func (p *LoggerDataReference) Fatal(err error) {
	LoggerLevelDataList[FatalLevel].logger.Printf(p.loggerPrefix+"%s%s", loggerLevelTypeNames[FatalLevel], err)
	os.Exit(1)
}

/*
LogErrorf delegates to log.Printf
*/
func (p *LoggerDataReference) LogErrorf(format string, v ...interface{}) {
	LoggerLevelDataList[ErrorLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[ErrorLevel]+format, v...)
}

/*
LogError delegates to log.Print
*/
func (p *LoggerDataReference) LogError(message string) {
	LoggerLevelDataList[ErrorLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[ErrorLevel] + message)
}

/*
LogInfof delegates to log.Printf
*/
func (p *LoggerDataReference) LogInfof(format string, v ...interface{}) {
	if LoggerLevelDataList[InfoLevel].active {
		LoggerLevelDataList[InfoLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[InfoLevel]+format, v...)
	}
}

/*
LogInfo delegates to log.Print
*/
func (p *LoggerDataReference) LogInfo(message string) {
	if LoggerLevelDataList[InfoLevel].active {
		LoggerLevelDataList[InfoLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[InfoLevel] + message)
	}
}

/*
LogAccessf delegates to log.Printf
*/
func (p *LoggerDataReference) LogAccessf(format string, v ...interface{}) {
	if LoggerLevelDataList[AccessLevel].active {
		LoggerLevelDataList[AccessLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[AccessLevel]+format, v...)
	}
}

/*
LogAccess delegates to log.Print
*/
func (p *LoggerDataReference) LogAccess(message string) {
	if LoggerLevelDataList[AccessLevel].active {
		LoggerLevelDataList[AccessLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[AccessLevel] + message)
	}
}

/*
LogWarnf delegates to log.Printf
*/
func (p *LoggerDataReference) LogWarnf(format string, v ...interface{}) {
	if LoggerLevelDataList[WarnLevel].active {
		LoggerLevelDataList[WarnLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[WarnLevel]+format, v...)
	}
}

/*
LogWarn delegates to log.Print
*/
func (p *LoggerDataReference) LogWarn(message string) {
	if LoggerLevelDataList[WarnLevel].active {
		LoggerLevelDataList[WarnLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[WarnLevel] + message)
	}
}

/*
LogDebugf delegates to log.Printf
*/
func (p *LoggerDataReference) LogDebugf(format string, v ...interface{}) {
	if LoggerLevelDataList[DebugLevel].active {
		LoggerLevelDataList[DebugLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[DebugLevel]+format, v...)
	}
}

/*
LogDebug delegates to log.Print
*/
func (p *LoggerDataReference) LogDebug(message string) {
	if LoggerLevelDataList[DebugLevel].active {
		LoggerLevelDataList[DebugLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[DebugLevel] + message)
	}
}

/*
	Validate and Activate each log level.
*/
func validateAndActivateLogLevels(values []state.LoggerLevelData) {
	/*
		For each log level definition
	*/
	for index, value := range values {
		/*
			check the name is valid
		*/
		name := strings.ToUpper(strings.TrimSpace(value.Level))
		if _, ok := loggerLevelMapNames[name]; ok {
			filedata := getLoggerWithFilename(value.File)
			LoggerLevelDataList[index].file = filedata
			if filedata == nil {
				LoggerLevelDataList[index].logger = log.New(os.Stdout, "", logDataFlags)
			} else {
				LoggerLevelDataList[index].logger = log.New(filedata.logFile, "", logDataFlags)
			}
		} else {
			list := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(loggerLevelTypeNames)), ", "), "[]")
			LogPanicToStdErrAndExit("The Log level name '" + value.Level + "' is not a valid log level. Valid values are:" + list)
		}
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
	name := strings.TrimSpace(strings.ToUpper(logFileName))
	if name == "" {
		return nil
	}
	if val, ok := loggerLevelFiles[name]; ok {
		return val
	}
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		LogPanicToStdErrAndExit("applicationID " + logApplicationID + ". Log file " + logFileName + " could NOT be Created or Opened\nError:" + err.Error())
	}
	lfd := &loggerFileData{
		fileName: logFileName,
		logFile:  f,
	}
	loggerLevelFiles[name] = lfd
	return lfd
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
