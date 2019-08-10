package test

import (
	"math/rand"
	"strconv"
	"testing"
	"webServerBase/logging"
	"webServerBase/state"
)

func TestCreateLogDefaults(t *testing.T) {
	logging.CreateLogWithFilenameAndAppID("", "AppDI", []state.LoggerLevelData{})
	defer logging.CloseLog()
	AssertEqualString(t, "", "DEBUG:In-Active", logging.LoggerLevelDataString("DEBUG"))
	AssertEqualString(t, "", "INFO:In-Active", logging.LoggerLevelDataString("INFO"))
	AssertEqualString(t, "", "WARN:In-Active", logging.LoggerLevelDataString("WARN"))
	AssertEqualString(t, "", "ACCESS:In-Active", logging.LoggerLevelDataString("ACCESS"))
	AssertEqualString(t, "", "ERROR:Active:Out=SysOut:", logging.LoggerLevelDataString("ERROR"))
	AssertEqualString(t, "", "FATAL:Active:Out=SysOut:", logging.LoggerLevelDataString("FATAL"))

	t1 := logging.NewLogger("T1")
	AssertFalse(t, "", t1.IsAccess())
	AssertFalse(t, "", t1.IsDebug())
	AssertFalse(t, "", t1.IsWarn())
	AssertFalse(t, "", t1.IsInfo())
}

func TestCreateLogDefaultsWithFile(t *testing.T) {
	l1 := state.LoggerLevelData{
		Level: "DEBUG",
	}
	logging.CreateLogWithFilenameAndAppID("ef.log", "AppDI", []state.LoggerLevelData{l1})
	defer logging.CloseLog()
	defer RemoveFile(t, "", logging.GetLogLevelFileName("ERROR"))

	AssertEqualString(t, "", "DEBUG:Active:Out=:ef.log:Open", logging.LoggerLevelDataString("DEBUG"))
	AssertEqualString(t, "", "INFO:In-Active", logging.LoggerLevelDataString("INFO"))
	AssertEqualString(t, "", "WARN:In-Active", logging.LoggerLevelDataString("WARN"))
	AssertEqualString(t, "", "ACCESS:In-Active", logging.LoggerLevelDataString("ACCESS"))
	AssertEqualString(t, "", "ERROR:Active:Out=:ef.log:Open", logging.LoggerLevelDataString("ERROR"))
	AssertEqualString(t, "", "FATAL:Active:Out=:ef.log:Open", logging.LoggerLevelDataString("FATAL"))
	t1 := logging.NewLogger("T1")
	t2 := logging.NewLogger("T2")

	AssertFalse(t, "", t1.IsAccess())
	AssertTrue(t, "", t1.IsDebug())
	AssertFalse(t, "", t1.IsWarn())
	AssertFalse(t, "", t1.IsInfo())
	t1Data := strconv.Itoa(rand.Int())
	t1.LogDebug(t1Data)
	t1.LogError(t1Data)
	t1.LogWarn(t1Data)
	t1.LogAccess(t1Data)
	t1.LogInfo(t1Data)
	t2Data := strconv.Itoa(rand.Int())
	t2.LogDebug(t2Data)
	t2.LogError(t2Data)
	t2.LogWarn(t2Data)
	t2.LogAccess(t2Data)
	t2.LogInfo(t2Data)
	AssertFileContains(t, "", logging.GetLogLevelFileName("DEBUG"), []string{
		"AppDI T1 [-]  DEBUG " + t1Data,
		"AppDI T2 [-]  DEBUG " + t2Data,
		"AppDI T1 [-]  ERROR " + t1Data,
		"AppDI T2 [-]  ERROR " + t2Data,
	})
	AssertFileDoesNotContain(t, "", logging.GetLogLevelFileName("DEBUG"), []string{
		"INFO",
		"ACCESS",
		"WARN",
		"INFO",
		"FATAL",
	})

}

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
	defer RemoveFile(t, "", logging.GetLogLevelFileName("DEBUG"))
	defer RemoveFile(t, "", logging.GetLogLevelFileName("INFO"))
	defer RemoveFile(t, "", logging.GetLogLevelFileName("ERROR"))

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
