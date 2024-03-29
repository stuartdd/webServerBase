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
	"sync"
	"time"

	"github.com/stuartdd/webServerBase/substitution"
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
This map is updated by CreateLogWithFilenameAndAppID.
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
	file         *logLevelFileData    // If the is a file associated with the log
}

/*
Map - Key is the Log level name, value is the definition (configuration) of that log
This is populated from logLevelDataMapKnownState map in
*/
var logLevelDataMap map[string]*logLevelData
var logLevelDataIndexList []*logLevelData

type logLevelFileData struct {
	fileName string   // The file name from the config data (used in map logLevelFileMap).
	logFile  *os.File // The actual file reference
}

/*
For each logger level there MAY be a file. Indexed by file name. This is so we can re-use the file with the same name for different levels
*/
var logLevelFileMap map[string]*logLevelFileData

/*
LoggerDataReference contains a ref to the single (named) logger instance and the module name (id).

Created via NewLogger
*/
type LoggerDataReference struct {
	loggerModuleName string // The name of the model (used in map logDataModules)
	loggerPrefix     string // Cached prefix for log lines throught this model
}

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
var fatalRC int

var mutex = &sync.Mutex{}

/*
false when Initialising true when complete!
*/
var loggerEnabled = false

/*
CreateTestLogger - creates a logger for testing
*/
func CreateTestLogger(id string) *LoggerDataReference {
	levels := make(map[string]string)
	for name := range logLevelDataMap {
		levels[name] = systemOutName
	}
	CreateLogWithFilenameAndAppID("", "TestLogger", -1, levels)
	return NewLogger(id)
}

/*
CreateLogWithFilenameAndAppID should configure the logger to output somthing like this!
2019-07-16 14:47:43.993 applicationID module  [-]  INFO Starti
2019-07-16 14:47:43.993 applicationID module  [-] DEBUG Runnin
*/
func CreateLogWithFilenameAndAppID(defaultLogFileNameIn string, applicationID string, fatalRCIn int, logLevelActivationData map[string]string) error {
	mutex.Lock()
	loggerEnabled = false
	fallBack = true
	defer mutex.Unlock()
	defer clearFlags()

	logDataFlags = log.LstdFlags | log.Lmicroseconds
	defaultLogFileName = defaultLogFileNameIn
	logApplicationID = applicationID
	fatalRC = fatalRCIn
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
		fmt.Printf("FALLBACK:Logging is in Fallback mode) Create Failed with error: %s\n" + err.Error())
		return err
	}
	fallBack = false
	return nil
}

/*
IsFallback - Return the fallback state
*/
func IsFallback() bool {
	return fallBack
}

/*
GetLogLevelTypeIndexForLevelName get the index for the level name
*/
func GetLogLevelTypeIndexForLevelName(name string) LoggerLevelTypeIndex {
	value := logLevelDataMap[strings.ToUpper(strings.TrimSpace(name))]
	if value == nil {
		return NotFound
	}
	return value.index
}

/*
GetLogLevelFileNameForLevelName get the file name for the level name
*/
func GetLogLevelFileNameForLevelName(name string) string {
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
CloseLog close ALL the log files
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
NewLogger creats a new logger instance for a specific module name

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
func IsDebug() bool {
	return logLevelDataIndexList[DebugLevel].active
}

/*
IsAccess return true is the access log function is enabled
*/
func IsAccess() bool {
	return logLevelDataIndexList[AccessLevel].active
}

/*
IsInfo return true is the info log function is enabled
*/
func IsInfo() bool {
	return logLevelDataIndexList[InfoLevel].active
}

/*
IsError return true is the error log function is enabled
*/
func IsError() bool {
	return logLevelDataIndexList[ErrorLevel].active
}

/*
IsFatal return true is the fatal log function is enabled
*/
func IsFatal() bool {
	return logLevelDataIndexList[FatalLevel].active
}

/*
IsWarn return true is the warn log function is enabled
*/
func IsWarn() bool {
	return logLevelDataIndexList[WarnLevel].active
}

/*
Fatal does the same as log.Fatal
Prints to console if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console

If fatalRC is not 0 the application will exit with that return code
*/
func (p *LoggerDataReference) Fatal(err error) {
	if fallBack {
		fmt.Printf("FALLBACK:FATAL: type[%T] %s\n", err, err.Error())
	} else {
		if logLevelDataIndexList[FatalLevel].active && isEnabled() {
			logLevelDataIndexList[FatalLevel].logger.Printf(p.loggerPrefix+"[%s] %T %s", logLevelDataIndexList[FatalLevel].paddedName, err, err.Error())
		} else {
			fmt.Printf("FATAL: type[%T] %s\n", err, err.Error())
		}
	}
	if fatalRC >= 0 {
		os.Exit(fatalRC)
	}
}

/*
LogErrorf delegates to log.Printf
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogErrorf(format string, v ...interface{}) {
	if fallBack {
		fmt.Printf("FALLBACK:ERROR: "+format+"\n", v...)
		return
	}
	if logLevelDataIndexList[ErrorLevel].active && isEnabled() {
		logLevelDataIndexList[ErrorLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[ErrorLevel].paddedName+format, v...)
	}
}

/*
LogError delegates to log.Print
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogError(message error) {
	if fallBack {
		fmt.Println("FALLBACK:ERROR: " + message.Error())
		return
	}
	if logLevelDataIndexList[ErrorLevel].active && isEnabled() {
		logLevelDataIndexList[ErrorLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[ErrorLevel].paddedName + message.Error())
	}
}

/*
LogInfof delegates to log.Printf
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogInfof(format string, v ...interface{}) {
	if fallBack {
		fmt.Printf("FALLBACK:INFO: "+format+"\n", v...)
		return
	}
	if logLevelDataIndexList[InfoLevel].active && isEnabled() {
		logLevelDataIndexList[InfoLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[InfoLevel].paddedName+format, v...)
	}
}

/*
LogInfo delegates to log.Print
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogInfo(message string) {
	if fallBack {
		fmt.Println("FALLBACK:INFO: " + message)
		return
	}
	if logLevelDataIndexList[InfoLevel].active && isEnabled() {
		logLevelDataIndexList[InfoLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[InfoLevel].paddedName + message)
	}
}

/*
LogAccessf delegates to log.Printf
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogAccessf(format string, v ...interface{}) {
	if fallBack {
		fmt.Printf("FALLBACK:ACCESS: "+format+"\n", v...)
		return
	}
	if logLevelDataIndexList[AccessLevel].active && isEnabled() {
		logLevelDataIndexList[AccessLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[AccessLevel].paddedName+format, v...)
	}
}

/*
LogAccess delegates to log.Print
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogAccess(message string) {
	if fallBack {
		fmt.Println("FALLBACK:ACCESS: " + message)
		return
	}
	if logLevelDataIndexList[AccessLevel].active && isEnabled() {
		logLevelDataIndexList[AccessLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[AccessLevel].paddedName + message)
	}
}

/*
LogWarnf delegates to log.Printf
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogWarnf(format string, v ...interface{}) {
	if fallBack {
		fmt.Printf("FALLBACK:WARN: "+format+"\n", v...)
		return
	}
	if logLevelDataIndexList[WarnLevel].active && isEnabled() {
		logLevelDataIndexList[WarnLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[WarnLevel].paddedName+format, v...)
	}
}

/*
LogWarn delegates to log.Print
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogWarn(message string) {
	if fallBack {
		fmt.Println("FALLBACK:WARN: " + message)
		return
	}
	if logLevelDataIndexList[WarnLevel].active && isEnabled() {
		logLevelDataIndexList[WarnLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[WarnLevel].paddedName + message)
	}
}

/*
LogDebugf delegates to log.Printf
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogDebugf(format string, v ...interface{}) {
	if fallBack {
		fmt.Printf("FALLBACK:DEBUG: "+format+"\n", v...)
		return
	}
	if logLevelDataIndexList[DebugLevel].active && isEnabled() {
		logLevelDataIndexList[DebugLevel].logger.Printf(p.loggerPrefix+logLevelDataIndexList[DebugLevel].paddedName+format, v...)
	}
}

/*
LogDebug delegates to log.Print.
Does nothing if logging is disabled or level is inactive
Fallback mode is true when logger configuration failed. It logs to the console
*/
func (p *LoggerDataReference) LogDebug(message string) {
	if fallBack {
		fmt.Println("FALLBACK:DEBUG: " + message)
		return
	}
	if logLevelDataIndexList[DebugLevel].active && isEnabled() {
		logLevelDataIndexList[DebugLevel].logger.Print(p.loggerPrefix + logLevelDataIndexList[DebugLevel].paddedName + message)
	}
}

/*
LogErrorWithStackTrace - Log an error with a prefix and a stack trace. Each line of the stacktrace has the prefix.
*/
func (p *LoggerDataReference) LogErrorWithStackTrace(txid string, prefix string, message string) {
	if fallBack {
		fmt.Println("FALLBACK:ERROR: " + prefix + " " + message + "\n" + string(debug.Stack()))
		return
	}
	prefix = "ID: " + txid + " " + prefix
	if logLevelDataIndexList[ErrorLevel].active && isEnabled() {
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
LoggerLevelDataString return the state of a log level as a string. UIsefull for testing and debugging)
*/
func LoggerLevelDataString(name string) string {
	loggerLevelTypeIndex := GetLogLevelTypeIndexForLevelName(name)
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
LogDirectToSystemError - Log to the syserr channel. Optionally ad a stach trace!
*/
func LogDirectToSystemError(msg string, withTrace bool) {
	var b bytes.Buffer
	l := log.New(os.Stderr, "", logDataFlags)
	b.WriteString(msg)
	b.WriteString("\n")
	if withTrace {
		b.WriteString(string(debug.Stack()))
		b.WriteString("\n")
	}
	l.Println(b.String())
}

func startFromKnownState() {
	/*
		Make sure the lists and maps are empty first
	*/
	CloseLog()
	logDataModules = make(map[string]*LoggerDataReference)
	logLevelFileMap = make(map[string]*logLevelFileData)
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

func isEnabled() bool {
	if loggerEnabled {
		return true
	}
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		if loggerEnabled {
			return true
		}

	}
	return false
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
		loggerLevelTypeIndex := GetLogLevelTypeIndexForLevelName(key)
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
						If default file name is undefined then choose stderr or stdout according to isErrorLevel
					*/
					if logLevelDataIndexList[GetLogLevelTypeIndexForLevelName(key)].isErrorLevel {
						loggerLevelDataValue.logger = log.New(os.Stderr, "", logDataFlags)
						loggerLevelDataValue.note = systemErrName
					} else {
						if logLevelDataIndexList[GetLogLevelTypeIndexForLevelName(key)].isErrorLevel {
							loggerLevelDataValue.logger = log.New(os.Stdout, "", logDataFlags)
							loggerLevelDataValue.note = systemOutName
						}
					}
				} else {
					logFileData, err := getLogLevelFileDataForFilename(defaultLogFileName)
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
				logFileData, err := getLogLevelFileDataForFilename(value)
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
			for name := range logLevelDataMap {
				b.WriteString(name)
				b.WriteString(", ")
			}
			return newError("The Log level name '" + key + "' is not a valid log level. Valid values are:" + b.String()[0:b.Len()-2])
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

func getLogLevelFileDataForFilename(logFileNameUnresolved string) (*logLevelFileData, error) {
	m := make(map[string]string)
	m["ID"] = logApplicationID

	logFileName := substitution.DoSubstitution(logFileNameUnresolved, m, '$')
	nameUcTrim := strings.TrimSpace(strings.ToUpper(logFileName))
	if val, ok := logLevelFileMap[nameUcTrim]; ok {
		return val, nil
	}
	if !strings.ContainsRune(logFileName, '.') {
		return nil, newError("applicationID " + logApplicationID + ". Log file " + logFileName + " is invalid. File name requires a '.' extension:")
	}
	absFileName, err := filepath.Abs(logFileName)
	if err != nil {
		return nil, newError("applicationID " + logApplicationID + ". Log file " + logFileName + " is not a valid file path: " + err.Error())
	}
	f, err := os.OpenFile(absFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, newError("applicationID " + logApplicationID + ". Log file " + logFileName + " could NOT be Created or Opened: " + err.Error())
	}
	lfd := &logLevelFileData{
		fileName: absFileName,
		logFile:  f,
	}
	logLevelFileMap[nameUcTrim] = lfd
	return lfd, nil
}

func newError(message string) error {
	return errors.New("Logging:" + message)
}

func clearFlags() {
	loggerEnabled = true
}
