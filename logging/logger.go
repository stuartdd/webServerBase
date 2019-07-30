package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type loggerLevelType int

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
A true in the slot means that log level is active
*/
var loggerLevelFlags = [...]bool{false, false, false, false, true, true}

/*
These values (not case sensitive) must map to the values passed to CreateLogWithFilenameAndAppID.
If these values are in the list then that log level will be active.
An empty list will mean that only ERROR and FATAL will be logged
*/
var loggerLevelMapNames = map[string]loggerLevelType{"INFO": InfoLevel, "DEBUG": DebugLevel, "WARN": WarnLevel, "ACCESS": AccessLevel, "ERROR": ErrorLevel, "FATAL": FatalLevel}

var longestModuleName int = 0

type loggerData struct {
	fileName string
	logFile  *os.File
	logger   *log.Logger
}

/*
LoggerDataReference contains a ref to th esingle logger instance and the module name (id).

Created via NewLogger
*/
type LoggerDataReference struct {
	loggerModuleName string
	loggerPrefix     string
	loggerDataRef    *loggerData
}

var logDataInstance *loggerData
var logDataModules map[string]*LoggerDataReference
var logDataFlags int
var logApplicationID string

/*
CreateLogWithFilenameAndAppID should configure the logger to output somthing like this!
2019-07-16 14:47:43.993 applicationID module  [-]  INFO Starti
2019-07-16 14:47:43.993 applicationID module  [-] DEBUG Runnin
*/
func CreateLogWithFilenameAndAppID(logFileName string, applicationID string, loggerLevelStrings []string) {
	processAndValidateLogLevels(loggerLevelStrings)

	logApplicationID = applicationID
	logDataFlags = log.LstdFlags | log.Lmicroseconds
	logDataModules = make(map[string]*LoggerDataReference)

	var logInstance *log.Logger
	var fileInstance *os.File
	if logFileName != "" {
		f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			LogPanicToStdErrAndExit("applicationID " + applicationID + ". Log file " + logFileName + " could NOT be opened\nError:" + err.Error())
		} else {
			logInstance = log.New(f, "", logDataFlags)
			fileInstance = f
		}
	} else {
		logInstance = log.New(os.Stdout, "", logDataFlags)
		fileInstance = nil
	}
	logDataInstance = &loggerData{
		fileName: logFileName,
		logFile:  fileInstance,
		logger:   logInstance,
	}
}

/*
NewLogger created a new logger instance for a specific module
All log lines printed via the returned ref will contain the specific module name.
*/
func NewLogger(moduleName string) *LoggerDataReference {
	if logDataInstance == nil {
		LogPanicToStdErrAndExit("Application or Module (" + moduleName + ") Must call CreateLogWithFilenameAndAppID before calling NewLogger")
	}

	if val, ok := logDataModules[moduleName]; ok {
		return val
	}
	ldRef := &LoggerDataReference{
		loggerModuleName: moduleName,
		loggerPrefix:     logApplicationID,
		loggerDataRef:    logDataInstance,
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
	return loggerLevelFlags[DebugLevel]
}

/*
IsAccess return true is the access log function is enabled
*/
func (p *LoggerDataReference) IsAccess() bool {
	return loggerLevelFlags[AccessLevel]
}

/*
IsInfo return true is the info log function is enabled
*/
func (p *LoggerDataReference) IsInfo() bool {
	return loggerLevelFlags[InfoLevel]
}

/*
IsWarn return true is the info log function is enabled
*/
func (p *LoggerDataReference) IsWarn() bool {
	return loggerLevelFlags[WarnLevel]
}

/*
Fatal does the same as log.Fatal
*/
func (p *LoggerDataReference) Fatal(err error) {
	p.loggerDataRef.logger.Printf(p.loggerPrefix+"%s%s", loggerLevelTypeNames[FatalLevel], err)
	os.Exit(1)
}

/*
LogErrorf delegates to log.Printf
*/
func (p *LoggerDataReference) LogErrorf(format string, v ...interface{}) {
	p.loggerDataRef.logger.Printf(p.loggerPrefix+loggerLevelTypeNames[ErrorLevel]+format, v...)
}

/*
LogError delegates to log.Print
*/
func (p *LoggerDataReference) LogError(message string) {
	p.loggerDataRef.logger.Print(p.loggerPrefix + loggerLevelTypeNames[ErrorLevel] + message)
}

/*
LogInfof delegates to log.Printf
*/
func (p *LoggerDataReference) LogInfof(format string, v ...interface{}) {
	if loggerLevelFlags[InfoLevel] {
		p.loggerDataRef.logger.Printf(p.loggerPrefix+loggerLevelTypeNames[InfoLevel]+format, v...)
	}
}

/*
LogInfo delegates to log.Print
*/
func (p *LoggerDataReference) LogInfo(message string) {
	if loggerLevelFlags[InfoLevel] {
		p.loggerDataRef.logger.Print(p.loggerPrefix + loggerLevelTypeNames[InfoLevel] + message)
	}
}

/*
LogAccessf delegates to log.Printf
*/
func (p *LoggerDataReference) LogAccessf(format string, v ...interface{}) {
	if loggerLevelFlags[AccessLevel] {
		p.loggerDataRef.logger.Printf(p.loggerPrefix+loggerLevelTypeNames[AccessLevel]+format, v...)
	}
}

/*
LogAccess delegates to log.Print
*/
func (p *LoggerDataReference) LogAccess(message string) {
	if loggerLevelFlags[AccessLevel] {
		p.loggerDataRef.logger.Print(p.loggerPrefix + loggerLevelTypeNames[AccessLevel] + message)
	}
}

/*
LogWarnf delegates to log.Printf
*/
func (p *LoggerDataReference) LogWarnf(format string, v ...interface{}) {
	if loggerLevelFlags[WarnLevel] {
		p.loggerDataRef.logger.Printf(p.loggerPrefix+loggerLevelTypeNames[WarnLevel]+format, v...)
	}
}

/*
LogWarn delegates to log.Print
*/
func (p *LoggerDataReference) LogWarn(message string) {
	if loggerLevelFlags[WarnLevel] {
		p.loggerDataRef.logger.Print(p.loggerPrefix + loggerLevelTypeNames[WarnLevel] + message)
	}
}

/*
LogDebugf delegates to log.Printf
*/
func (p *LoggerDataReference) LogDebugf(format string, v ...interface{}) {
	if loggerLevelFlags[DebugLevel] {
		p.loggerDataRef.logger.Printf(p.loggerPrefix+loggerLevelTypeNames[DebugLevel]+format, v...)
	}
}

/*
LogDebug delegates to log.Print
*/
func (p *LoggerDataReference) LogDebug(message string) {
	if loggerLevelFlags[DebugLevel] {
		p.loggerDataRef.logger.Print(p.loggerPrefix + loggerLevelTypeNames[DebugLevel] + message)
	}
}

/*
CloseLog close the log file
*/
func CloseLog(logger *LoggerDataReference) {
	if logDataInstance.logFile != nil {
		logger.LogInfof("logging.CloseLog: Log file %s is closing", logDataInstance.fileName)
		logDataInstance.logFile.Close()
	} else {
		logger.LogWarn("logging.CloseLog: Was called but there is NO log file open")
	}
}

func processAndValidateLogLevels(values []string) {
	for _, value := range values {
		name := strings.ToUpper(value)
		if val, ok := loggerLevelMapNames[name]; ok {
			loggerLevelFlags[val] = true
		} else {
			list := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(loggerLevelTypeNames)), ", "), "[]")
			LogPanicToStdErrAndExit("The Log level name '" + value + "' is not a valid log level. Valid values are:" + list)
		}
	}
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
