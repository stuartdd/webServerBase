package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

const errorName = "ERROR"
const fatalName = "FATAL"
const systemErrName = "SYSERR"
const systemOutName = "SYSOUT"
const defaultName = "DEFAULT"
const offName = "OFF"

/*
LoggerLevelTypeIndex ENUM for log levels. Used to index in to lists
*/
type LoggerLevelTypeIndex int

/*
InfoLevel is the finest. Nothing stops ErrorLevel of FatalLevel
*/
const (
	InfoLevel LoggerLevelTypeIndex = iota
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
	note         string
	active       bool
	isErrorLevel bool
	logger       *log.Logger
	file         *loggerFileData
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
var loggerLevelMapNames = map[string]LoggerLevelTypeIndex{"INFO": InfoLevel, "DEBUG": DebugLevel, "WARN": WarnLevel, "ACCESS": AccessLevel, errorName: ErrorLevel, fatalName: FatalLevel}

/*
For each logger level there MAY be a file. Indexed by file name. This is so we can re-use the file with the same name for different levels
*/
var loggerLevelFiles map[string]*loggerFileData
var logDataModules map[string]*LoggerDataReference
var loggerLevelDataList = initLoggerLevelDataList()

var longestModuleName int

var logDataFlags int
var logApplicationID string
var defaultLogFileName string
var fallBack = true

/*
CreateTestLogger - creates a logger for testing
*/
func CreateTestLogger(id string) *LoggerDataReference {
	levels := make(map[string]string)
	for logName :=range loggerLevelMapNames {
		levels[logName] = systemOutName
	}
	CreateLogWithFilenameAndAppID("", "TestLogger", levels)
	return NewLogger(id)
}

/*
CreateLogWithFilenameAndAppID should configure the logger to output somthing like this!
2019-07-16 14:47:43.993 applicationID module  [-]  INFO Starti
2019-07-16 14:47:43.993 applicationID module  [-] DEBUG Runnin
*/
func CreateLogWithFilenameAndAppID(defaultLogFileNameIn string, applicationID string, logLeveldata map[string]string) error {
	CloseLog()
	defaultLogFileName = defaultLogFileNameIn
	logApplicationID = applicationID
	logDataFlags = log.LstdFlags | log.Lmicroseconds
	logDataModules = make(map[string]*LoggerDataReference)
	loggerLevelFiles = make(map[string]*loggerFileData)
	loggerLevelDataList = initLoggerLevelDataList()
	longestModuleName = 0
	/*
		Validate and Activate each log level.
		We MUST activate Error and Fatal so add them if not already defined
	*/
	if logLeveldata[errorName] == "" {
		logLeveldata[errorName] = defaultName
	}
	if logLeveldata[fatalName] == "" {
		logLeveldata[fatalName] = defaultName
	}
	err := validateAndActivateLogLevels(logLeveldata)
	if err != nil {
		return err
	}
	fallBack = false
	return nil
}

/*
GetLogLevelTypeIndexForName get the index for the level name
*/
func GetLogLevelTypeIndexForName(name string) LoggerLevelTypeIndex {
	if index, ok := loggerLevelMapNames[strings.ToUpper(strings.TrimSpace(name))]; ok {
		return index
	}
	return NotFound
}

/*
GetLogLevelFileName get the file name for the level name
*/
func GetLogLevelFileName(name string) string {
	loggerLevelTypeIndex := GetLogLevelTypeIndexForName(name)
	if loggerLevelTypeIndex != NotFound {
		typeInstance := loggerLevelDataList[loggerLevelTypeIndex]
		if typeInstance.file != nil {
			return typeInstance.file.fileName
		}
	}
	return ""
}

/*
LoggerLevelDataString return the state of a log level as a string
*/
func LoggerLevelDataString(name string) string {
	loggerLevelTypeIndex := GetLogLevelTypeIndexForName(name)
	if loggerLevelTypeIndex != NotFound {
		lld := loggerLevelDataList[loggerLevelTypeIndex]
		errorLevel := "NO"
		if lld.isErrorLevel {
			errorLevel = "YES"
		}
		if lld.active {
			active := name + ":Active note[" + lld.note + "] error[" + errorLevel + "]:"
			if lld.file == nil {
				return active + "Out=Console:"
			}
			active = active + "Out=:" + filepath.Base(lld.file.fileName)
			if lld.file.logFile == nil {
				return active + ":Closed"
			}
			return active + ":Open"

		}
		return name + ":In-Active note[" + lld.note + "] error[" + errorLevel + "]"
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
	if logDataModules == nil {
		logDataModules = make(map[string]*LoggerDataReference)
	}
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
IsError return true is the error log function is enabled
*/
func (p *LoggerDataReference) IsError() bool {
	return loggerLevelDataList[ErrorLevel].active
}

/*
IsFatal return true is the fatal log function is enabled
*/
func (p *LoggerDataReference) IsFatal() bool {
	return loggerLevelDataList[FatalLevel].active
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
		fmt.Printf("FATAL: type[%T] %s\n", err, err.Error())
		os.Exit(1)
	} else {
		if loggerLevelDataList[FatalLevel].active {
			loggerLevelDataList[FatalLevel].logger.Printf(p.loggerPrefix+"[%s] %T %s", loggerLevelTypeNames[FatalLevel], err, err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("FATAL: type[%T] %s\n", err, err.Error())
		}
	}
}

/*
LogErrorf delegates to log.Printf
*/
func (p *LoggerDataReference) LogErrorf(format string, v ...interface{}) {
	if loggerLevelDataList[ErrorLevel].active {
		loggerLevelDataList[ErrorLevel].logger.Printf(p.loggerPrefix+loggerLevelTypeNames[ErrorLevel]+format, v...)
	}
}

/*
LogError delegates to log.Print
*/
func (p *LoggerDataReference) LogError(message error) {
	if loggerLevelDataList[ErrorLevel].active {
		loggerLevelDataList[ErrorLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[ErrorLevel] + message.Error())
	}
}

/*
LogErrorWithStackTrace - Log an error and a stack trace
*/
func (p *LoggerDataReference) LogErrorWithStackTrace(prefix string, message string) {
	if p.IsError() {
		loggerLevelDataList[ErrorLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[ErrorLevel] + prefix + " " + message)
		st := string(debug.Stack())
		for count, line := range strings.Split(strings.TrimSuffix(st, "\n"), "\n") {
			if count > 6 && count <= 18 {
				loggerLevelDataList[ErrorLevel].logger.Print(p.loggerPrefix + loggerLevelTypeNames[ErrorLevel] + prefix + " " + line)
			}
		}
	}
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

func initLoggerLevelDataList() []*loggerLevelData {
	return []*loggerLevelData{
		newLoggerLevelTypeData(false, false),
		newLoggerLevelTypeData(false, false),
		newLoggerLevelTypeData(false, false),
		newLoggerLevelTypeData(false, false),
		newLoggerLevelTypeData(true, true),
		newLoggerLevelTypeData(true, true)}
}

func logError(message string) error {
	log.Panic("Logging:" + message)
	return errors.New("Logging:" + message)
}

/*
	Validate and Activate each log level.
*/
func validateAndActivateLogLevels(values map[string]string) error {
	/*
		For each log level definition
	*/
	for key, value := range values {
		/*
			check the name is valid
		*/
		loggerLevelTypeIndex := GetLogLevelTypeIndexForName(key)
		if loggerLevelTypeIndex != NotFound {
			/*
				Configure the level according to the value in the map (case insensitive).
				Note the value can be a file name!
			*/
			loggerLevelDataValue := loggerLevelDataList[loggerLevelTypeIndex]
			valueUC := strings.TrimSpace(strings.ToUpper(value))
			switch valueUC {
			case offName, "":
				loggerLevelDataValue.active = false
				loggerLevelDataValue.note = valueUC
				break
			case systemOutName:
				loggerLevelDataValue.logger = log.New(os.Stdout, "", logDataFlags)
				loggerLevelDataValue.active = true
				loggerLevelDataValue.note = valueUC
				break
			case systemErrName:
				loggerLevelDataValue.logger = log.New(os.Stderr, "", logDataFlags)
				loggerLevelDataValue.active = true
				loggerLevelDataValue.note = valueUC
				break
			case defaultName:
				/*
					For default we use the the default file name
				*/
				if defaultLogFileName == "" {
					/*
						If default file name is undefined thgen choose stderr or stdout according to isErrorLevel
					*/
					if loggerLevelDataList[GetLogLevelTypeIndexForName(key)].isErrorLevel {
						loggerLevelDataValue.logger = log.New(os.Stderr, "", logDataFlags)
						loggerLevelDataValue.note = systemErrName
					} else {
						if loggerLevelDataList[GetLogLevelTypeIndexForName(key)].isErrorLevel {
							loggerLevelDataValue.logger = log.New(os.Stdout, "", logDataFlags)
							loggerLevelDataValue.note = systemOutName
						}
					}
				} else {
					logFileData, err := getLoggerWithFilename(defaultLogFileName)
					if err != nil {
						return err
					}
					loggerLevelDataValue.file = logFileData
					loggerLevelDataValue.logger = log.New(logFileData.logFile, "", logDataFlags)
					loggerLevelDataValue.note = valueUC
				}
				loggerLevelDataValue.active = true
				break
			default:
				logFileData, err := getLoggerWithFilename(value)
				if err != nil {
					return err
				}
				loggerLevelDataValue.file = logFileData
				loggerLevelDataValue.logger = log.New(logFileData.logFile, "", logDataFlags)
				loggerLevelDataValue.active = true
				loggerLevelDataValue.note = "FILE"
			}
		} else {
			list := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(loggerLevelTypeNames)), ", "), "[]")
			return logError("The Log level name '" + key + "' is not a valid log level. Valid values are:" + list)
		}
	}
	return nil
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

func newLoggerLevelTypeData(active bool, isError bool) *loggerLevelData {
	return &loggerLevelData{
		note:         "UNDEFINED",
		active:       active,
		isErrorLevel: isError,
		logger:       nil,
		file:         nil,
	}
}

func getLoggerWithFilename(logFileName string) (*loggerFileData, error) {
	nameUcTrim := strings.TrimSpace(strings.ToUpper(logFileName))
	if val, ok := loggerLevelFiles[nameUcTrim]; ok {
		return val, nil
	}
	if !strings.ContainsRune(logFileName, '.') {
		return nil, logError("applicationID " + logApplicationID + ". Log file " + logFileName + " is invalid. File name requires a '.' extension:")
	}
	absFileName, err := filepath.Abs(logFileName)
	if err != nil {
		return nil, logError("applicationID " + logApplicationID + ". Log file " + logFileName + " is not a valid file path: " + err.Error())
	}
	f, err := os.OpenFile(absFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, logError("applicationID " + logApplicationID + ". Log file " + logFileName + " could NOT be Created or Opened: " + err.Error())
	}
	lfd := &loggerFileData{
		fileName: absFileName,
		logFile:  f,
	}
	loggerLevelFiles[nameUcTrim] = lfd
	return lfd, nil
}
