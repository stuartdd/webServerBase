package logging

import (
	"fmt"
	"log"
	"os"
)

type loggerData struct {
	fileName string
	logFile  *os.File
	logger   *log.Logger
}

var logDataInstance *loggerData

/*
CreateLogWithFilename should output somthing like this!
2019-07-16 14:47:43.993 Instance module  [-]  INFO Starti
2019-07-16 14:47:43.993 Instance module  [-] DEBUG Runnin
*/
func CreateLogWithFilename(logFileName string) {
	var logInstance *log.Logger
	var fileInstance *os.File
	flags := log.LstdFlags | log.Lmicroseconds
	if logFileName != "" {
		f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Log file '%s' could NOT be opened\nError:%s", logFileName, err.Error())
			return
		}
		logInstance = log.New(f, "", flags)
		fileInstance = f
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
func Fatal(v ...interface{}) {
	logDataInstance.logger.Fatal(v)
}

/*
Logf delegates to log.Printf
*/
func Logf(format string, v ...interface{}) {
	go logDataInstance.logger.Printf("appid: module: [-] DEBUG: "+format, v)
}

/*
CloseLog close the log file
*/
func CloseLog() {
	if logDataInstance.logFile != nil {
		logDataInstance.logFile.Close()
	}
}
