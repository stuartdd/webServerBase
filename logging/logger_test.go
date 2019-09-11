package logging

import (
	"fmt"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"webServerBase/test"
)

var globalErr interface{}

/*
Create an error type so we can test the logError functionality
*/
type TrialError struct {
	Desc string
	Line int
	Col  int
}

/*
Overrite the Error method to generate our own message
*/
func (e *TrialError) Error() string {
	return fmt.Sprintf("%d:%d: Trial error: %s", e.Line, e.Col, e.Desc)
}

/*
Create a new
*/
func NewTrialError(desc string) *TrialError {
	return &TrialError{
		Desc: desc,
		Line: 10,
		Col:  20,
	}
}

func TestInvalidLevel(t *testing.T) {
	defer checkPanicIsThrown(t, "is not a valid log level")
	levels := make(map[string]string)
	levels["DEBUGd"] = "SYSERR"
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
}

func TestInvalidOption(t *testing.T) {
	defer checkPanicIsThrown(t, "File name requires a '.'")
	levels := make(map[string]string)
	levels["DEBUG"] = "SYSSSS"
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
}

func TestCreateLogDefaults(t *testing.T) {
	CreateLogWithFilenameAndAppID("", "AppID", -1, make(map[string]string))
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active note[OFF] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
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
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:Active note[SYSERR] error[NO]:Out=Console:", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
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
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active note[OFF] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:In-Active note[OFF] error[YES]", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertFalse(t, "IsAccess", t1.IsAccess())
	test.AssertFalse(t, "IsDebug", t1.IsDebug())
	test.AssertFalse(t, "IsWarn", t1.IsWarn())
	test.AssertFalse(t, "IsInfo", t1.IsInfo())
	test.AssertFalse(t, "IsError", t1.IsError())
	test.AssertTrue(t, "IsFatal", t1.IsFatal())
	t1.LogError(NewTrialError("SwitchOffError ended OK"))
}

func TestSwitchOffFatal(t *testing.T) {
	levels := make(map[string]string)
	levels["FATAL"] = "OFF"
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	defer CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active note[OFF] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:In-Active note[OFF] error[YES]", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertFalse(t, "IsAccess", t1.IsAccess())
	test.AssertFalse(t, "IsDebug", t1.IsDebug())
	test.AssertFalse(t, "IsWarn", t1.IsWarn())
	test.AssertFalse(t, "IsInfo", t1.IsInfo())
	test.AssertTrue(t, "IsError", t1.IsError())
	test.AssertFalse(t, "IsFatal", t1.IsFatal())
	t1.Fatal(NewTrialError("SwitchOffFatal ended OK"))
}

func TestCreateLogDefaultsWithFile(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUG"] = "DEFAULT"
	CreateLogWithFilenameAndAppID("ef.log", "AppID", -1, levels)
	/*
		Note last defer runs first. It is a stack!
		CloseLog must run before test.RemoveFile
	*/
	defer test.RemoveFile(t, "", GetLogLevelFileNameForLevelName("ERROR"))
	defer CloseLog()

	test.AssertEqualString(t, "", "DEBUG:Active note[DEFAULT] error[NO]:Out=:ef.log:Open", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
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
	t1.LogError(NewTrialError(t1Data))
	t1.LogWarn(t1Data)
	t1.LogAccess(t1Data)
	t1.LogInfo(t1Data)
	t2Data := strconv.Itoa(rand.Int())
	t2.LogDebug(t2Data)
	t2.LogError(NewTrialError(t2Data))
	t2.LogWarn(t2Data)
	t2.LogAccess(t2Data)
	t2.LogInfo(t2Data)
	fileName := GetLogLevelFileNameForLevelName("DEBUG")
	test.AssertFileContains(t, "", fileName, []string{
		"AppID T1 [-]  DEBUG " + t1Data,
		"AppID T2 [-]  DEBUG " + t2Data,
		"AppID T1 [-]  ERROR 10:20: Trial error: " + t1Data,
		"AppID T2 [-]  ERROR 10:20: Trial error: " + t2Data,
	})
	test.AssertFileDoesNotContain(t, "", fileName, []string{
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

	CreateLogWithFilenameAndAppID("ef.log", "AppID", -1, levels)
	defer test.RemoveFile(t, "", GetLogLevelFileNameForLevelName("DEBUG"))
	defer test.RemoveFile(t, "", GetLogLevelFileNameForLevelName("INFO"))
	defer test.RemoveFile(t, "", GetLogLevelFileNameForLevelName("ERROR"))

	test.AssertEqualString(t, "", "DEBUG:Active note[FILE] error[NO]:Out=:d.log:Open", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:Active note[FILE] error[NO]:Out=:i.log:Open", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:Active note[FILE] error[NO]:Out=:i.log:Open", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("FATAL"))

	CloseLog()
	test.AssertEqualString(t, "", "DEBUG:In-Active note[FILE] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertEqualString(t, "", "INFO:In-Active note[FILE] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertEqualString(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertEqualString(t, "", "ACCESS:In-Active note[FILE] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertEqualString(t, "", "ERROR:In-Active note[DEFAULT] error[YES]", LoggerLevelDataString("ERROR"))
	test.AssertEqualString(t, "", "FATAL:In-Active note[DEFAULT] error[YES]", LoggerLevelDataString("FATAL"))
	test.AssertEqualString(t, "", "UNKNOWN:Not Found", LoggerLevelDataString("UNKNOWN"))

	test.AssertEndsWithString(t, "", GetLogLevelFileNameForLevelName("DEBUG"), "d.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileNameForLevelName("INFO"), "i.log")
	test.AssertEqualString(t, "", "", GetLogLevelFileNameForLevelName("WARN"))
	test.AssertEndsWithString(t, "", GetLogLevelFileNameForLevelName("ACCESS"), "i.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileNameForLevelName("ERROR"), "ef.log")
	test.AssertEndsWithString(t, "", GetLogLevelFileNameForLevelName("FATAL"), "ef.log")

}

func TestNameToIndex(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUG"] = "DEFAULT"
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	test.AssertEqualInt(t, "GetLogLevelTypeForName:DEBUG", int(DebugLevel), int(GetLogLevelTypeIndexForLevelName("DEBUG")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:INFO", int(InfoLevel), int(GetLogLevelTypeIndexForLevelName("INFO")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:WARN", int(WarnLevel), int(GetLogLevelTypeIndexForLevelName("WARN")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:ACCESS", int(AccessLevel), int(GetLogLevelTypeIndexForLevelName("ACCESS")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:ERROR", int(ErrorLevel), int(GetLogLevelTypeIndexForLevelName("ERROR")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:FATAL", int(FatalLevel), int(GetLogLevelTypeIndexForLevelName("FATAL")))
	test.AssertEqualInt(t, "GetLogLevelTypeForName:UNKNOWN", int(NotFound), int(GetLogLevelTypeIndexForLevelName("UNKNOWN")))
}

func checkPanicIsThrown(t *testing.T, desc string) {
	if r := recover(); r != nil {
		s := fmt.Sprintf("%s", r)
		if strings.Contains(s, desc) {
			return
		}
		t.Log(string(debug.Stack()))
		t.Fatalf("\nERROR:\n  %s\nDOES NOT CONTAIN:\n  %s\n", s, desc)
	}
	t.Fatalf("\nTEST did not panic:\nThe test MUST panic with error message containing:'%s'", desc)
}
