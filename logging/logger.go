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
	ErrorLevel
	FatalLevel
)

/*
These names should be ALL the same length and should have a ' ' before AND after the name
*/
var loggerLevelTypeNames = [...]string{"  INFO ", " DEBUG ", " ERROR ", " FATAL "}

/*
A true in the slot means that log level is active
*/
var loggerLevelFlags = [...]bool{false, false, true, true}

/*
These values (not case sensitive) must map to the values passed to CreateLogWithFilenameAndAppID.
If these values are in the list then that log level will be active.
An empty list will mean that only ERROR and FATAL will be logged
*/
var loggerLevelMapNames = map[string]loggerLevelType{"INFO": InfoLevel, "DEBUG": DebugLevel, "ERROR": ErrorLevel, "FATAL": FatalLevel}

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
	loggerPrefix  string
	loggerDataRef *loggerData
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
	if longestModuleName < len(moduleName) {
		longestModuleName = len(moduleName)
	}
	ldRef := &LoggerDataReference{
		loggerPrefix:  logApplicationID + " " + moduleName + " [-] ",
		loggerDataRef: logDataInstance,
	}
	logDataModules[moduleName] = ldRef
	return ldRef

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
	p.loggerDataRef.logger.Printf(p.loggerPrefix+loggerLevelTypeNames[ErrorLevel]+format, v)
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
		p.loggerDataRef.logger.Printf(p.loggerPrefix+loggerLevelTypeNames[InfoLevel]+format, v)
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
LogDebugf delegates to log.Printf
*/
func (p *LoggerDataReference) LogDebugf(format string, v ...interface{}) {
	if loggerLevelFlags[DebugLevel] {
		p.loggerDataRef.logger.Printf(p.loggerPrefix+loggerLevelTypeNames[DebugLevel]+format, v)
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
func CloseLog() {
	if logDataInstance.logFile != nil {
		logDataInstance.logFile.Close()
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
