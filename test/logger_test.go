package test

import (
	"testing"
	"webServerBase/logging"
	"webServerBase/state"
)

func TestCreateLog(t *testing.T) {
	l1 := state.LoggerLevelData{
		Level: "DEBUG",
		File:  "debugFile.log",
	}
	l2 := state.LoggerLevelData{
		Level: "INFO",
		File:  "",
	}
	ll := []state.LoggerLevelData{l1, l2}
	logging.CreateLogWithFilenameAndAppID("tempFile.log", "AppDI", ll)
}
