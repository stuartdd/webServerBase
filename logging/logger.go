package logging

import (
	"log"
	"os"
)

type loggerData struct {
	fileName string
	logFile  *os.File
	logger   *log.Logger
}

type LoggerDataReference struct {
	id string
}

var logDataInstance *loggerData
var logApplicationId string

/*
CreateLogWithFilename should output somthing like this!
2019-07-16 14:47:43.993 Instance module  [-]  INFO Starti
2019-07-16 14:47:43.993 Instance module  [-] DEBUG Runnin
*/
func CreateLogWithFilename(logFileName string, applicationID string) {
	logApplicationId = applicationID
	var logInstance *log.Logger
	var fileInstance *os.File
	flags := log.LstdFlags | log.Lmicroseconds
	if logFileName != "" {
		f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logInstance = log.New(os.Stdout, "", flags)
			fileInstance = nil
			logInstance.Printf("Log file '%s' could NOT be opened\nError:%s", logFileName, err.Error())
		} else {
			logInstance = log.New(f, "", flags)
			fileInstance = f
		}
	} else {
		logInstance = log.New(os.Stdout, "", flags)
		fileInstance = nil
	}
	logDataInstance = &loggerData{
		fileName: logFileName,
		logFile:  fileInstance,
		logger:   logInstance,
	}
}

/*
Fatal delegates to log.Fatal
*/
func Fatal(err error) {
	logDataInstance.logger.Printf(logApplicationId+": module: [-] FATAL: %s", err)
	os.Exit(1)
}

/*
Logf delegates to log.Printf
*/
func Logf(format string, v ...interface{}) {
	logDataInstance.logger.Printf(logApplicationId+": module: [-] DEBUG: "+format, v)
}
/*
Log delegates to log.Print
*/
func Log(message string) {
	logDataInstance.logger.Print(logApplicationId+": module: [-] DEBUG: " + message)
}

/*
CloseLog close the log file
*/
func CloseLog() {
	if logDataInstance.logFile != nil {
		logDataInstance.logFile.Close()
	}
}
