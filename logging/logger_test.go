package logging

import (
	"math/rand"
	"strconv"
	"testing"
	"webServerBase/state"
	"webServerBase/test"
)

func TestCreateLogDefaults(t *testing.T) {
	CreateLogWithFilenameAndAppID("", "AppDI", []state.LoggerLevelData{})
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active:Out=SysOut:", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active:Out=SysOut:", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertFalse(t, "", t1.IsAccess())
	test.AssertFalse(t, "", t1.IsDebug())
	test.AssertFalse(t, "", t1.IsWarn())
	test.AssertFalse(t, "", t1.IsInfo())
}

func TestCreateLogDefaultsWithFile(t *testing.T) {
	l1 := state.LoggerLevelData{
		Level: "DEBUG",
	}
	CreateLogWithFilenameAndAppID("ef.log", "AppDI", []state.LoggerLevelData{l1})
	/*
		Note last defer runs first. It is a stack!
		CloseLog must run before test.RemoveFile
	*/
	defer test.RemoveFile(t, "", GetLogLevelFileName("ERROR"))
	defer CloseLog()

	test.AssertEqualString(t, "", "DEBUG:Active:Out=:ef.log:Open", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active:Out=:ef.log:Open", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active:Out=:ef.log:Open", LoggerLevelDataString("FATAL"))
	t1 := NewLogger("T1")
	t2 := NewLogger("T2")

	test.AssertFalse(t, "", t1.IsAccess())
	test.AssertTrue(t, "", t1.IsDebug())
	test.AssertFalse(t, "", t1.IsWarn())
	test.AssertFalse(t, "", t1.IsInfo())
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
	test.AssertFileContains(t, "", GetLogLevelFileName("DEBUG"), []string{
		"AppDI T1 [-]  DEBUG " + t1Data,
		"AppDI T2 [-]  DEBUG " + t2Data,
		"AppDI T1 [-]  ERROR " + t1Data,
		"AppDI T2 [-]  ERROR " + t2Data,
	})
	test.AssertFileDoesNotContain(t, "", GetLogLevelFileName("DEBUG"), []string{
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
	CreateLogWithFilenameAndAppID("ef.log", "AppDI", ll)
	defer test.RemoveFile(t, "", GetLogLevelFileName("DEBUG"))
	defer test.RemoveFile(t, "", GetLogLevelFileName("INFO"))
	defer test.RemoveFile(t, "", GetLogLevelFileName("ERROR"))

	test.AssertEqualString(t, "", "DEBUG:Active:Out=:d.log:Open", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:Active:Out=:i.log:Open", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:Active:Out=:i.log:Open", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active:Out=:ef.log:Open", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active:Out=:ef.log:Open", LoggerLevelDataString("FATAL"))

	CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:In-Active", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:In-Active", LoggerLevelDataString("FATAL"))
	test.AssertEqualString(t, "", "UNKNOWN:Not Found", LoggerLevelDataString("UNKNOWN"))

	test.AssertEndsWithString(t, "", GetLogLevelFileName("DEBUG"), "d.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileName("INFO"), "i.log")
	test.AssertEqualString(t, "", "", GetLogLevelFileName("WARN"))
	test.AssertEndsWithString(t, "", GetLogLevelFileName("ACCESS"), "i.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileName("ERROR"), "ef.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileName("FATAL"), "ef.log")

}

func TestNameToIndex(t *testing.T) {
	test.AssertEqualInt(t, "GetLogLevelTypeForName:DEBUG", int(DebugLevel), int(GetLogLevelTypeForName("DEBUG")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:INFO", int(InfoLevel), int(GetLogLevelTypeForName("INFO")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:WARN", int(WarnLevel), int(GetLogLevelTypeForName("WARN")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:ACCESS", int(AccessLevel), int(GetLogLevelTypeForName("ACCESS")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:ERROR", int(ErrorLevel), int(GetLogLevelTypeForName("ERROR")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:FATAL", int(FatalLevel), int(GetLogLevelTypeForName("FATAL")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:UNKNOWN", int(NotFound), int(GetLogLevelTypeForName("UNKNOWN")))
}
