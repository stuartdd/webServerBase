package test

import (
	"testing"
	"webServerBase/logging"
	"webServerBase/state"
)

func TestCreateLog(t *testing.T) {
	l1 := state.LoggerLevelData{
		Level: "DEBUG",
		File:  "d.log",
	}
	l2 := state.LoggerLevelData{
		Level: "INFO",
		File:  "i.log",
	}
	l3 := state.LoggerLevelData{
		Level: "ACCESS",
		File:  "i.log",
	}

	ll := []state.LoggerLevelData{l1, l2, l3}
	logging.CreateLogWithFilenameAndAppID("ef.log", "AppDI", ll)

	AssertEqualString(t, "", "DEBUG:Active:Out=:d.log:Open", logging.LoggerLevelDataString("DEBUG"))
	AssertEqualString(t, "", "INFO:Active:Out=:i.log:Open", logging.LoggerLevelDataString("INFO"))
	AssertEqualString(t, "", "WARN:In-Active", logging.LoggerLevelDataString("WARN"))
	AssertEqualString(t, "", "ACCESS:Active:Out=:i.log:Open", logging.LoggerLevelDataString("ACCESS"))
	AssertEqualString(t, "", "ERROR:Active:Out=:ef.log:Open", logging.LoggerLevelDataString("ERROR"))
	AssertEqualString(t, "", "FATAL:Active:Out=:ef.log:Open", logging.LoggerLevelDataString("FATAL"))
	logging.CloseLog()
	AssertEqualString(t, "", "DEBUG:In-Active", logging.LoggerLevelDataString("DEBUG"))
	AssertEqualString(t, "", "INFO:In-Active", logging.LoggerLevelDataString("INFO"))
	AssertEqualString(t, "", "WARN:In-Active", logging.LoggerLevelDataString("WARN"))
	AssertEqualString(t, "", "ACCESS:In-Active", logging.LoggerLevelDataString("ACCESS"))
	AssertEqualString(t, "", "ERROR:In-Active", logging.LoggerLevelDataString("ERROR"))
	AssertEqualString(t, "", "FATAL:In-Active", logging.LoggerLevelDataString("FATAL"))
	AssertEqualString(t, "", "UNKNOWN:Not Found", logging.LoggerLevelDataString("UNKNOWN"))

	AssertEndsWithString(t, "", logging.GetLogLevelFileName("DEBUG"), "d.log")
	AssertEndsWithString(t, "", logging.GetLogLevelFileName("INFO"), "i.log")
	AssertEqualString(t, "", "", logging.GetLogLevelFileName("WARN"))
	AssertEndsWithString(t, "", logging.GetLogLevelFileName("ACCESS"), "i.log")
	AssertEndsWithString(t, "", logging.GetLogLevelFileName("ERROR"), "ef.log")
	AssertEndsWithString(t, "", logging.GetLogLevelFileName("FATAL"), "ef.log")

	//	AssertFileContains(t, "", logging.GetLogLevelFileName("DEBUG"), []string{"X"})

}
func TestNameToIndex(t *testing.T) {
	AssertEqualInt(t, "GetLogLevelTypeForName:DEBUG", int(logging.DebugLevel), int(logging.GetLogLevelTypeForName("DEBUG")))
	AssertEqualInt(t, "GetLogLevelTypeForName:INFO", int(logging.InfoLevel), int(logging.GetLogLevelTypeForName("INFO")))
	AssertEqualInt(t, "GetLogLevelTypeForName:WARN", int(logging.WarnLevel), int(logging.GetLogLevelTypeForName("WARN")))
	AssertEqualInt(t, "GetLogLevelTypeForName:ACCESS", int(logging.AccessLevel), int(logging.GetLogLevelTypeForName("ACCESS")))
	AssertEqualInt(t, "GetLogLevelTypeForName:ERROR", int(logging.ErrorLevel), int(logging.GetLogLevelTypeForName("ERROR")))
	AssertEqualInt(t, "GetLogLevelTypeForName:FATAL", int(logging.FatalLevel), int(logging.GetLogLevelTypeForName("FATAL")))
	AssertEqualInt(t, "GetLogLevelTypeForName:UNKNOWN", int(logging.NotFound), int(logging.GetLogLevelTypeForName("UNKNOWN")))
}
