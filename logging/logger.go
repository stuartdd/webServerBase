package logging

import (
	"bytes"
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
Create default instances of each log level.
This map is updated by CreateLogWithFilenameAndAppID. This can cause issues when test is run as teh state cannot be reset!
*/
var logLevelDataMapKnownState = map[string]*logLevelDataKnownState{
	"INFO":   &logLevelDataKnownState{InfoLevel, offName, false, false},
	"DEBUG":  &logLevelDataKnownState{DebugLevel, offName, false, false},
	"WARN":   &logLevelDataKnownState{WarnLevel, offName, false, false},
	"ACCESS": &logLevelDataKnownState{AccessLevel, offName, false, false},
	"ERROR":  &logLevelDataKnownState{ErrorLevel, systemErrName, true, true},
	"FATAL":  &logLevelDataKnownState{FatalLevel, systemErrName, true, true},
}

type logLevelDataKnownState struct {
	index        LoggerLevelTypeIndex // The index. This identifies which log data is used. No duplicates allowed!
	note         string               // mode - OFF, SYSERR, SYSOUT etc
	active       bool                 // Is the log logging?
	isErrorLevel bool                 // Is it or error or fatal log as these are active by default
}

/*
logLevelData One instance per Log Level. Note the order of the fields must match the above definition.
*/
type logLevelData struct {
	paddedName   string
	index        LoggerLevelTypeIndex // The index. This identifies which log data is used. No duplicates allowed!
	note         string               // mode - OFF, SYSERR, SYSOUT etc
	active       bool                 // Is the log logging?
	isErrorLevel bool                 // Is it or error or fatal log as these are active by default
	logger       *log.Logger          // The actual (wrapped) logger imported via "log"
	file         *loggerFileData      // If the is a file associated with the log
}

/*
These values (not case sensitive) must map to the values passed to CreateLogWithFilenameAndAppID.
If these values are in the list then that log level will be active.
An empty list will mean that only ERROR and FATAL will be logged
*/

type loggerFileData struct {
	fileName string   // The file name from the config data (used in map loggerLevelFiles).
	logFile  *os.File // The actual file reference
}

/*
LoggerDataReference contains a ref to the single logger instance and the module name (id).

Created via NewLogger
*/
type LoggerDataReference struct {
	loggerModuleName string // The name of the model (used in map logDataModules)
	loggerPrefix     string // Cached prefix for log lines throught this model
}

/*
Map - Key is the Log level name, value is the definition (configuration) of that log
This is populated from logLevelDataMapKnownState map in
*/
var logLevelDataMap map[string]*logLevelData
var logLevelDataIndexList []*logLevelData

/*
For each logger level there MAY be a file. Indexed by file name. This is so we can re-use the file with the same name for different levels
*/
var loggerLevelFiles map[string]*loggerFileData

/*
For each module, a module name and a cached prefix
*/
var logDataModules map[string]*LoggerDataReference

/*
Used when creating a log (log.New...)
*/
var logDataFlags int

/*
The identity of teh application. This is in the prefix for each line
*/
var logApplicationID string
var defaultLogFileName string
var fallBack = true

/*
CreateTestLogger - creates a logger for testing
*/
func CreateTestLogger(id string) *LoggerDataReference {
	levels := make(map[string]string)
	for name := range logLevelDataMap {
		levels[name] = systemOutName
	}
	CreateLogWithFilenameAndAppID("", "TestLogger", levels)
	return NewLogger(id)
}

/*
CreateLogWithFilenameAndAppID should configure the logger to output somthing like this!
2019-07-16 14:47:43.993 applicationID module  [-]  INFO Starti
2019-07-16 14:47:43.993 applicationID module  [-] DEBUG Runnin
*/
func CreateLogWithFilenameAndAppID(defaultLogFileNameIn string, applicationID string, logLevelActivationData map[string]string) error {
	logDataFlags = log.LstdFlags | log.Lmicroseconds
	defaultLogFileName = defaultLogFileNameIn
	logApplicationID = applicationID
	startFromKnownState()
	/*
		Validate and Activate each log level.
		We MUST activate Error and Fatal so add them to the inINPUTput list if not already defined
	*/
	if logLevelActivationData[errorName] == "" {
		logLevelActivationData[errorName] = defaultName
	}
	if logLevelActivationData[fatalName] == "" {
		logLevelActivationData[fatalName] = defaultName
	}
	err := validateAndActivateLogLevels(logLevelActivationData)
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
	value := logLevelDataMap[strings.ToUpper(strings.TrimSpace(name))]
	if value == nil {
		return NotFound
	}
	return value.index
}

/*
GetLogLevelFileName get the file name for the level name
*/
func GetLogLevelFileName(name string) string {
	value, ok := logLevelDataMap[strings.ToUpper(strings.TrimSpace(name))]
	if !ok {
		return ""
	}
	if value.file == nil {
		return ""
	}
	return value.file.fileName
}

/*
CloseLog close the log file
*/
func CloseLog() {
	for _, value := range logLevelDataMap {
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
		loggerPrefix:     "",
	}
	logDataModules[moduleName] = ldRef
	updateLoggerPrefixesForAllModules()
	return ldRef
}

/*
IsDebug return true is the debug log function is enabled
*/
func (p *LoggerDataReference) IsDebug() bool {
	return logLevelDataIndexList[DebugLevel].active
}

/*
IsAccess return true is the access log function is enabled
*/
func (p *LoggerDataReference) IsAccess() bool {
	return logLevelDataIndexList[AccessLevel].active
}

/*
IsInfo return true is the info log function is enabled
*/
func (p *LoggerDataReference) IsInfo() bool {
	return logLevelDataIndexList[InfoLevel].active
}

/*
IsError return true is the error log function is enabled
*/
func (p *LoggerDataReference) IsError() bool {
	return logLevelDataIndexList[ErrorLevel].active
}

/*
IsFatal return true is the fatal log function is enabled
*/
func (p *LoggerDataReference) IsFatal() bool {
	return logLevelDataIndexList[FatalLevel].active
}

/*
IsWarn return true is the info log function is enabled
*/
func (p *LoggerDataReference) IsWarn() bool {
	return logLevelDataIndexList[WarnLevel].active
}

/*
Fatal does the same as log.Fatal
*/
func (p *LoggerDataReference) Fatal(err error) {
	if fallBack {
		fmt.Printf("FATAL: type[%T] %s\n", err, err.Error())
		os.Exit(1)
	} else {
		if logLevelDataIndexList[FatalLevel].active {
			logLevelDataIndexList[FatalLevel].logger.Printf(p.loggerPrefix+"[%s] %T %s", logLevelDataIndexList[FatalLevel].paddedName, err, err.Error())
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
	if logLevelDataIndexList[ErrorLevel].active {
		logLevelDataIndexList[ErrorLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[ErrorLevel].paddedName+format, v...)
	}
}

/*
LogError delegates to log.Print
*/
func (p *LoggerDataReference) LogError(message error) {
	if logLevelDataIndexList[ErrorLevel].active {
		logLevelDataIndexList[ErrorLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[ErrorLevel].paddedName + message.Error())
	}
}

/*
LogErrorWithStackTrace - Log an error and a stack trace
*/
func (p *LoggerDataReference) LogErrorWithStackTrace(prefix string, message string) {
	if p.IsError() {
		logLevelDataIndexList[ErrorLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[ErrorLevel].paddedName + prefix + " " + message)
		st := string(debug.Stack())
		for count, line := range strings.Split(strings.TrimSuffix(st, "\n"), "\n") {
			if count > 6 && count <= 18 {
				logLevelDataIndexList[ErrorLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[ErrorLevel].paddedName + prefix + " " + line)
			}
		}
	}
}

/*
LogInfof delegates to log.Printf
*/
func (p *LoggerDataReference) LogInfof(format string, v ...interface{}) {
	if logLevelDataIndexList[InfoLevel].active {
		logLevelDataIndexList[InfoLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[InfoLevel].paddedName+format, v...)
	}
}

/*
LogInfo delegates to log.Print
*/
func (p *LoggerDataReference) LogInfo(message string) {
	if logLevelDataIndexList[InfoLevel].active {
		logLevelDataIndexList[InfoLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[InfoLevel].paddedName + message)
	}
}

/*
LogAccessf delegates to log.Printf
*/
func (p *LoggerDataReference) LogAccessf(format string, v ...interface{}) {
	if logLevelDataIndexList[AccessLevel].active {
		logLevelDataIndexList[AccessLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[AccessLevel].paddedName+format, v...)
	}
}

/*
LogAccess delegates to log.Print
*/
func (p *LoggerDataReference) LogAccess(message string) {
	if logLevelDataIndexList[AccessLevel].active {
		logLevelDataIndexList[AccessLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[AccessLevel].paddedName + message)
	}
}

/*
LogWarnf delegates to log.Printf
*/
func (p *LoggerDataReference) LogWarnf(format string, v ...interface{}) {
	if logLevelDataIndexList[WarnLevel].active {
		logLevelDataIndexList[WarnLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[WarnLevel].paddedName+format, v...)
	}
}

/*
LogWarn delegates to log.Print
*/
func (p *LoggerDataReference) LogWarn(message string) {
	if logLevelDataIndexList[WarnLevel].active {
		logLevelDataIndexList[WarnLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[WarnLevel].paddedName + message)
	}
}

/*
LogDebugf delegates to log.Printf
*/
func (p *LoggerDataReference) LogDebugf(format string, v ...interface{}) {
	if logLevelDataIndexList[DebugLevel].active {
		logLevelDataIndexList[DebugLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[DebugLevel].paddedName+format, v...)
	}
}

/*
LoggerLevelDataString return the state of a log level as a string. UIsefull for testing and debugging)
*/
func LoggerLevelDataString(name string) string {
	loggerLevelTypeIndex := GetLogLevelTypeIndexForName(name)
	if loggerLevelTypeIndex != NotFound {
		lld := logLevelDataIndexList[loggerLevelTypeIndex]
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
LogDebug delegates to log.Print
*/
func (p *LoggerDataReference) LogDebug(message string) {
	if logLevelDataIndexList[DebugLevel].active {
		logLevelDataIndexList[DebugLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[DebugLevel].paddedName + message)
	}
}

func logError(message string) error {
	log.Panic("Logging:" + message)
	return errors.New("Logging:" + message)
}

func startFromKnownState() {
	/*
		Make sure the lists and maps are empty first
	*/
	CloseLog()
	logDataModules = make(map[string]*LoggerDataReference)
	loggerLevelFiles = make(map[string]*loggerFileData)
	logLevelDataMap = make(map[string]*logLevelData)
	logLevelDataIndexList = []*logLevelData{}
	/*
		Find the longest name
	*/
	longest := 0
	for name := range logLevelDataMapKnownState {
		if longest < len(name) {
			longest = len(name)
		}
	}
	/*
		Create each logLevelData from each logLevelDataKnownState
	*/
	for name, value := range logLevelDataMapKnownState {
		logLevelDataMap[name] = &logLevelData{
			paddedName:   padName(name, longest),
			index:        value.index,
			note:         value.note,
			active:       value.active,
			isErrorLevel: value.isErrorLevel,
		}
	}
	/*
		Create a list of empty (nil) values the right size.
	*/
	for i := 0; i < len(logLevelDataMap); i++ {
		logLevelDataIndexList = append(logLevelDataIndexList, nil)
	}
	/*
		Insert each logLevelData in the correct slot. Duplicate entries will cause a panic

		Note they may not be in the right sequence (it depends on values in logLevelDataMapKnownState)
			so we need to use the value.index from the logLevelData to make sure order is maintained!
	*/
	for name, value := range logLevelDataMap {
		if logLevelDataIndexList[value.index] == nil {
			logLevelDataIndexList[value.index] = value
		} else {
			panic("Duplicate index value in logLevelDataMapKnownState[" + name + "]")
		}
	}
}

func padName(name string, longest int) string {
	return strings.Repeat(" ", longest-len(name)) + name + " "
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
			loggerLevelDataValue := logLevelDataIndexList[loggerLevelTypeIndex]
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
					if logLevelDataIndexList[GetLogLevelTypeIndexForName(key)].isErrorLevel {
						loggerLevelDataValue.logger = log.New(os.Stderr, "", logDataFlags)
						loggerLevelDataValue.note = systemErrName
					} else {
						if logLevelDataIndexList[GetLogLevelTypeIndexForName(key)].isErrorLevel {
							loggerLevelDataValue.logger = log.New(os.Stdout, "", logDataFlags)
							loggerLevelDataValue.note = systemOutName
						}
					}
				} else {
					logFileData, err := getLoggerFileDataWithFilename(defaultLogFileName)
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
				logFileData, err := getLoggerFileDataWithFilename(value)
				if err != nil {
					return err
				}
				loggerLevelDataValue.file = logFileData
				loggerLevelDataValue.logger = log.New(logFileData.logFile, "", logDataFlags)
				loggerLevelDataValue.active = true
				loggerLevelDataValue.note = "FILE"
			}
		} else {
			var b bytes.Buffer
			for index := 0; index < len(logLevelDataIndexList); index++ {
				b.WriteString(logLevelDataIndexList[index].paddedName)
				b.WriteString(", ")
			}
			return logError("The Log level name '" + key + "' is not a valid log level. Valid values are:" + b.String()[0:b.Len()-2])
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

func getLoggerFileDataWithFilename(logFileName string) (*loggerFileData, error) {
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
