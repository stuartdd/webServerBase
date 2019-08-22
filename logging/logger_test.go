package logging

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"webServerBase/test"
)

var globalErr interface{}

type SyntaxError struct {
	Desc string
	Line int
	Col  int
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("%d:%d: syntax error: %s", e.Line, e.Col, e.Desc)
}

func NewSyntaxError(desc string) *SyntaxError {
	return &SyntaxError{
		Desc: desc,
		Line: 10,
		Col:  20,
	}
}

func TestInvalidLevel(t *testing.T) {
	defer checkPanicIsThrown(t, "is not a valid log level")
	levels := make(map[string]string)
	levels["DEBUGd"] = "SYSERR"
	CreateLogWithFilenameAndAppID("", "AppDI", levels)
}

func TestInvalidOption(t *testing.T) {
	defer checkPanicIsThrown(t, "File name requires a '.'")
	levels := make(map[string]string)
	levels["DEBUG"] = "SYSSSS"
	CreateLogWithFilenameAndAppID("", "AppDI", levels)
}

func TestCreateLogDefaults(t *testing.T) {
	CreateLogWithFilenameAndAppID("", "AppDI", make(map[string]string))
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertFalse(t, "IsAccess", t1.IsAccess())
	test.AssertFalse(t, "IsDebug", t1.IsDebug())
	test.AssertFalse(t, "IsWarn", t1.IsWarn())
	test.AssertFalse(t, "IsInfo", t1.IsInfo())
	test.AssertTrue(t, "IsError", t1.IsError())
	test.AssertTrue(t, "IsFatal", t1.IsFatal())
}

func TestDebugToSysErr(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUG"] = "SYSERR"
	levels["ERROR"] = "SYSOUT"
	CreateLogWithFilenameAndAppID("", "AppDI", levels)
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:Active note[SYSERR] error[NO]:Out=Console:", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active note[SYSOUT] error[YES]:Out=Console:", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertFalse(t, "IsAccess", t1.IsAccess())
	test.AssertTrue(t, "IsDebug", t1.IsDebug())
	test.AssertFalse(t, "IsWarn", t1.IsWarn())
	test.AssertFalse(t, "IsInfo", t1.IsInfo())
	test.AssertTrue(t, "IsError", t1.IsError())
	test.AssertTrue(t, "IsFatal", t1.IsFatal())
}

func TestSwitchOffError(t *testing.T) {
	levels := make(map[string]string)
	levels["ERROR"] = "OFF"
	CreateLogWithFilenameAndAppID("", "AppDI", levels)
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:In-Active note[OFF] error[YES]", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertFalse(t, "IsAccess", t1.IsAccess())
	test.AssertFalse(t, "IsDebug", t1.IsDebug())
	test.AssertFalse(t, "IsWarn", t1.IsWarn())
	test.AssertFalse(t, "IsInfo", t1.IsInfo())
	test.AssertFalse(t, "IsError", t1.IsError())
	test.AssertTrue(t, "IsFatal", t1.IsFatal())
	t1.LogError(NewSyntaxError("Test Error OFF"))
}

func TestSwitchOffFatal(t *testing.T) {
	levels := make(map[string]string)
	levels["FATAL"] = "OFF"
	CreateLogWithFilenameAndAppID("", "AppDI", levels)
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:In-Active note[OFF] error[YES]", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertFalse(t, "IsAccess", t1.IsAccess())
	test.AssertFalse(t, "IsDebug", t1.IsDebug())
	test.AssertFalse(t, "IsWarn", t1.IsWarn())
	test.AssertFalse(t, "IsInfo", t1.IsInfo())
	test.AssertTrue(t, "IsError", t1.IsError())
	test.AssertFalse(t, "IsFatal", t1.IsFatal())
	t1.Fatal(NewSyntaxError("Test Fatal OFF"))
}

func TestCreateLogDefaultsWithFile(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUG"] = "DEFAULT"
	CreateLogWithFilenameAndAppID("ef.log", "AppDI", levels)
	/*
		Note last defer runs first. It is a stack!
		CloseLog must run before test.RemoveFile
	*/
	defer test.RemoveFile(t, "", GetLogLevelFileName("ERROR"))
	defer CloseLog()

	test.AssertEqualString(t, "", "DEBUG:Active note[DEFAULT] error[NO]:Out=:ef.log:Open", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("FATAL"))
	t1 := NewLogger("T1")
	t2 := NewLogger("T2")

	test.AssertFalse(t, "IsAccess", t1.IsAccess())
	test.AssertTrue(t, "IsDebug", t1.IsDebug())
	test.AssertFalse(t, "IsWarn", t1.IsWarn())
	test.AssertFalse(t, "IsInfo", t1.IsInfo())
	test.AssertTrue(t, "IsError", t1.IsError())
	test.AssertTrue(t, "IsFatal", t1.IsFatal())
	t1Data := strconv.Itoa(rand.Int())
	t1.LogDebug(t1Data)
	t1.LogError(NewSyntaxError(t1Data))
	t1.LogWarn(t1Data)
	t1.LogAccess(t1Data)
	t1.LogInfo(t1Data)
	t2Data := strconv.Itoa(rand.Int())
	t2.LogDebug(t2Data)
	t2.LogError(NewSyntaxError(t2Data))
	t2.LogWarn(t2Data)
	t2.LogAccess(t2Data)
	t2.LogInfo(t2Data)
	test.AssertFileContains(t, "", GetLogLevelFileName("DEBUG"), []string{
		"AppDI T1 [-]  DEBUG " + t1Data,
		"AppDI T2 [-]  DEBUG " + t2Data,
		"AppDI T1 [-]  ERROR 10:20: syntax error: " + t1Data,
		"AppDI T2 [-]  ERROR 10:20: syntax error: " + t2Data,
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
	levels := make(map[string]string)
	levels["DEBUG"] = "d.log"
	levels["INFO"] = "i.log"
	levels["ACCESS"] = "i.log"

	CreateLogWithFilenameAndAppID("ef.log", "AppDI", levels)
	defer test.RemoveFile(t, "", GetLogLevelFileName("DEBUG"))
	defer test.RemoveFile(t, "", GetLogLevelFileName("INFO"))
	defer test.RemoveFile(t, "", GetLogLevelFileName("ERROR"))

	test.AssertEqualString(t, "", "DEBUG:Active note[FILE] error[NO]:Out=:d.log:Open", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:Active note[FILE] error[NO]:Out=:i.log:Open", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:Active note[FILE] error[NO]:Out=:i.log:Open", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("FATAL"))

	CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active note[FILE] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[FILE] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[UNDEFINED] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[FILE] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:In-Active note[DEFAULT] error[YES]", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:In-Active note[DEFAULT] error[YES]", LoggerLevelDataString("FATAL"))
	test.AssertEqualString(t, "", "UNKNOWN:Not Found", LoggerLevelDataString("UNKNOWN"))

	test.AssertEndsWithString(t, "", GetLogLevelFileName("DEBUG"), "d.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileName("INFO"), "i.log")
	test.AssertEqualString(t, "", "", GetLogLevelFileName("WARN"))
	test.AssertEndsWithString(t, "", GetLogLevelFileName("ACCESS"), "i.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileName("ERROR"), "ef.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileName("FATAL"), "ef.log")

}

func TestNameToIndex(t *testing.T) {
	test.AssertEqualInt(t, "GetLogLevelTypeForName:DEBUG", int(DebugLevel), int(GetLogLevelTypeIndexForName("DEBUG")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:INFO", int(InfoLevel), int(GetLogLevelTypeIndexForName("INFO")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:WARN", int(WarnLevel), int(GetLogLevelTypeIndexForName("WARN")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:ACCESS", int(AccessLevel), int(GetLogLevelTypeIndexForName("ACCESS")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:ERROR", int(ErrorLevel), int(GetLogLevelTypeIndexForName("ERROR")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:FATAL", int(FatalLevel), int(GetLogLevelTypeIndexForName("FATAL")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:UNKNOWN", int(NotFound), int(GetLogLevelTypeIndexForName("UNKNOWN")))
}

func checkPanicIsThrown(t *testing.T, desc string) {
	if r := recover(); r != nil {
		s := fmt.Sprintf("%s", r)
		if strings.Contains(s, desc) {
			return
		}
		t.Fatalf("\nERROR:\n  %s\nDOES NOT CONTAIN:\n  %s\n", s, desc)
	}
	t.Fatalf("\nTEST did not panic:\nThe test MUST panic with error message containing:'%s'", desc)
}
