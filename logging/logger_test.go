package logging

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"github.com/stuartdd/webServerBase/test"
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

func TestFallback(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUG"] = "ZZZZZZ"
	/*
		Test that fallback did not leave the constructor locked!
	*/
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	test.AssertBoolTrue(t, "", IsFallback())
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tstlog := NewLogger("NEW-LOG")
	tstlog.LogAccess("LINE 1")
	tstlog.LogAccessf("LINE %d", 2)
	tstlog.LogDebug("LINE 3")
	tstlog.LogDebugf("LINE %d", 4)
	tstlog.LogWarn("LINE 5")
	tstlog.LogWarnf("LINE %d", 6)
	tstlog.LogInfo("LINE 7")
	tstlog.LogInfof("LINE %d", 8)
	tstlog.LogError(fmt.Errorf("LINE %d", 9))
	tstlog.LogErrorf("LINE %d", 10)
	tstlog.LogErrorWithStackTrace("TRACE 11", "LINE 11")
	tstlog.Fatal(fmt.Errorf("FATAL %d", 12))

	test.AssertBoolFalse(t, "", tstlog.IsAccess())
	test.AssertBoolFalse(t, "", tstlog.IsWarn())
	test.AssertBoolFalse(t, "", tstlog.IsDebug())
	test.AssertBoolFalse(t, "", tstlog.IsInfo())
	test.AssertBoolTrue(t, "", tstlog.IsError())
	test.AssertBoolTrue(t, "", tstlog.IsFatal())

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	captured := string(out)

	t.Log("AA *******************************\n" + captured + "BB *********************************")
	test.AssertStringContains(t, "", captured, []string{
		"FALLBACK:ACCESS: LINE 1",
		"FALLBACK:ACCESS: LINE 2",
		"FALLBACK:DEBUG: LINE 3",
		"FALLBACK:DEBUG: LINE 4",
		"FALLBACK:WARN: LINE 5",
		"FALLBACK:WARN: LINE 6",
		"FALLBACK:INFO: LINE 7",
		"FALLBACK:INFO: LINE 8",
		"FALLBACK:ERROR: LINE 9",
		"FALLBACK:ERROR: LINE 10",
		"FALLBACK:ERROR: TRACE 11 LINE 11",
		"logging.(*LoggerDataReference).LogErrorWithStackTrace",
		"FALLBACK:FATAL: type[*errors.errorString] FATAL 12",
	})

}

func TestInvalidLevel(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUGd"] = "SYSERR"
	test.AssertError(t, "Maust have an error", CreateLogWithFilenameAndAppID("", "AppID", -1, levels))
}

func TestInvalidOption(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUG"] = "SYSSSS"
	test.AssertError(t, "Maust have an error", CreateLogWithFilenameAndAppID("", "AppID", -1, levels))
}

func TestCreateLogDefaults(t *testing.T) {
	CreateLogWithFilenameAndAppID("", "AppID", -1, make(map[string]string))
	defer CloseLog()
	test.AssertStringEquals(t, "", "DEBUG:In-Active note[OFF] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertStringEquals(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertStringEquals(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertStringEquals(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertStringEquals(t, "", "ERROR:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("ERROR"))
	test.AssertStringEquals(t, "", "FATAL:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertBoolFalse(t, "IsAccess", t1.IsAccess())
	test.AssertBoolFalse(t, "IsDebug", t1.IsDebug())
	test.AssertBoolFalse(t, "IsWarn", t1.IsWarn())
	test.AssertBoolFalse(t, "IsInfo", t1.IsInfo())
	test.AssertBoolTrue(t, "IsError", t1.IsError())
	test.AssertBoolTrue(t, "IsFatal", t1.IsFatal())
}

func TestDebugToSysErr(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUG"] = "SYSERR"
	levels["ERROR"] = "SYSOUT"
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	defer CloseLog()
	test.AssertStringEquals(t, "", "DEBUG:Active note[SYSERR] error[NO]:Out=Console:", LoggerLevelDataString("DEBUG"))
	test.AssertStringEquals(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertStringEquals(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertStringEquals(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertStringEquals(t, "", "ERROR:Active note[SYSOUT] error[YES]:Out=Console:", LoggerLevelDataString("ERROR"))
	test.AssertStringEquals(t, "", "FATAL:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertBoolFalse(t, "IsAccess", t1.IsAccess())
	test.AssertBoolTrue(t, "IsDebug", t1.IsDebug())
	test.AssertBoolFalse(t, "IsWarn", t1.IsWarn())
	test.AssertBoolFalse(t, "IsInfo", t1.IsInfo())
	test.AssertBoolTrue(t, "IsError", t1.IsError())
	test.AssertBoolTrue(t, "IsFatal", t1.IsFatal())
}

func TestSwitchOffError(t *testing.T) {
	levels := make(map[string]string)
	levels["ERROR"] = "OFF"
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	defer CloseLog()
	test.AssertStringEquals(t, "", "DEBUG:In-Active note[OFF] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertStringEquals(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertStringEquals(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertStringEquals(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertStringEquals(t, "", "ERROR:In-Active note[OFF] error[YES]", LoggerLevelDataString("ERROR"))
	test.AssertStringEquals(t, "", "FATAL:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertBoolFalse(t, "IsAccess", t1.IsAccess())
	test.AssertBoolFalse(t, "IsDebug", t1.IsDebug())
	test.AssertBoolFalse(t, "IsWarn", t1.IsWarn())
	test.AssertBoolFalse(t, "IsInfo", t1.IsInfo())
	test.AssertBoolFalse(t, "IsError", t1.IsError())
	test.AssertBoolTrue(t, "IsFatal", t1.IsFatal())
	t1.LogError(NewTrialError("SwitchOffError ended OK"))
}

func TestSwitchOffFatal(t *testing.T) {
	levels := make(map[string]string)
	levels["FATAL"] = "OFF"
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	defer CloseLog()
	test.AssertStringEquals(t, "", "DEBUG:In-Active note[OFF] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertStringEquals(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertStringEquals(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertStringEquals(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertStringEquals(t, "", "ERROR:Active note[SYSERR] error[YES]:Out=Console:", LoggerLevelDataString("ERROR"))
	test.AssertStringEquals(t, "", "FATAL:In-Active note[OFF] error[YES]", LoggerLevelDataString("FATAL"))

	t1 := NewLogger("T1")
	test.AssertBoolFalse(t, "IsAccess", t1.IsAccess())
	test.AssertBoolFalse(t, "IsDebug", t1.IsDebug())
	test.AssertBoolFalse(t, "IsWarn", t1.IsWarn())
	test.AssertBoolFalse(t, "IsInfo", t1.IsInfo())
	test.AssertBoolTrue(t, "IsError", t1.IsError())
	test.AssertBoolFalse(t, "IsFatal", t1.IsFatal())
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
	defer test.AssertFileRemoved(t, "", GetLogLevelFileNameForLevelName("ERROR"))
	defer CloseLog()

	test.AssertStringEquals(t, "", "DEBUG:Active note[DEFAULT] error[NO]:Out=:ef.log:Open", LoggerLevelDataString("DEBUG"))
	test.AssertStringEquals(t, "", "INFO:In-Active note[OFF] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertStringEquals(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertStringEquals(t, "", "ACCESS:In-Active note[OFF] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertStringEquals(t, "", "ERROR:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("ERROR"))
	test.AssertStringEquals(t, "", "FATAL:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("FATAL"))
	t1 := NewLogger("T1")
	t2 := NewLogger("T2")

	test.AssertBoolFalse(t, "IsAccess", t1.IsAccess())
	test.AssertBoolTrue(t, "IsDebug", t1.IsDebug())
	test.AssertBoolFalse(t, "IsWarn", t1.IsWarn())
	test.AssertBoolFalse(t, "IsInfo", t1.IsInfo())
	test.AssertBoolTrue(t, "IsError", t1.IsError())
	test.AssertBoolTrue(t, "IsFatal", t1.IsFatal())
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
	defer test.AssertFileRemoved(t, "", GetLogLevelFileNameForLevelName("DEBUG"))
	defer test.AssertFileRemoved(t, "", GetLogLevelFileNameForLevelName("INFO"))
	defer test.AssertFileRemoved(t, "", GetLogLevelFileNameForLevelName("ERROR"))

	test.AssertStringEquals(t, "", "DEBUG:Active note[FILE] error[NO]:Out=:d.log:Open", LoggerLevelDataString("DEBUG"))
	test.AssertStringEquals(t, "", "INFO:Active note[FILE] error[NO]:Out=:i.log:Open", LoggerLevelDataString("INFO"))
	test.AssertStringEquals(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertStringEquals(t, "", "ACCESS:Active note[FILE] error[NO]:Out=:i.log:Open", LoggerLevelDataString("ACCESS"))
	test.AssertStringEquals(t, "", "ERROR:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("ERROR"))
	test.AssertStringEquals(t, "", "FATAL:Active note[DEFAULT] error[YES]:Out=:ef.log:Open", LoggerLevelDataString("FATAL"))

	CloseLog()
	test.AssertStringEquals(t, "", "DEBUG:In-Active note[FILE] error[NO]", LoggerLevelDataString("DEBUG"))
	test.AssertStringEquals(t, "", "INFO:In-Active note[FILE] error[NO]", LoggerLevelDataString("INFO"))
	test.AssertStringEquals(t, "", "WARN:In-Active note[OFF] error[NO]", LoggerLevelDataString("WARN"))
	test.AssertStringEquals(t, "", "ACCESS:In-Active note[FILE] error[NO]", LoggerLevelDataString("ACCESS"))
	test.AssertStringEquals(t, "", "ERROR:In-Active note[DEFAULT] error[YES]", LoggerLevelDataString("ERROR"))
	test.AssertStringEquals(t, "", "FATAL:In-Active note[DEFAULT] error[YES]", LoggerLevelDataString("FATAL"))
	test.AssertStringEquals(t, "", "UNKNOWN:Not Found", LoggerLevelDataString("UNKNOWN"))

	test.AssertStringEndsWith(t, "", GetLogLevelFileNameForLevelName("DEBUG"), "d.log")
	test.AssertStringEndsWith(t, "", GetLogLevelFileNameForLevelName("INFO"), "i.log")
	test.AssertStringEquals(t, "", "", GetLogLevelFileNameForLevelName("WARN"))
	test.AssertStringEndsWith(t, "", GetLogLevelFileNameForLevelName("ACCESS"), "i.log")
	test.AssertStringEndsWith(t, "", GetLogLevelFileNameForLevelName("ERROR"), "ef.log")
	test.AssertStringEndsWith(t, "", GetLogLevelFileNameForLevelName("FATAL"), "ef.log")

}

func TestNameToIndex(t *testing.T) {
	levels := make(map[string]string)
	levels["DEBUG"] = "DEFAULT"
	CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
	test.AssertIntEqual(t, "GetLogLevelTypeForName:DEBUG", int(DebugLevel), int(GetLogLevelTypeIndexForLevelName("DEBUG")))
	test.AssertIntEqual(t, "GetLogLevelTypeForName:INFO", int(InfoLevel), int(GetLogLevelTypeIndexForLevelName("INFO")))
	test.AssertIntEqual(t, "GetLogLevelTypeForName:WARN", int(WarnLevel), int(GetLogLevelTypeIndexForLevelName("WARN")))
	test.AssertIntEqual(t, "GetLogLevelTypeForName:ACCESS", int(AccessLevel), int(GetLogLevelTypeIndexForLevelName("ACCESS")))
	test.AssertIntEqual(t, "GetLogLevelTypeForName:ERROR", int(ErrorLevel), int(GetLogLevelTypeIndexForLevelName("ERROR")))
	test.AssertIntEqual(t, "GetLogLevelTypeForName:FATAL", int(FatalLevel), int(GetLogLevelTypeIndexForLevelName("FATAL")))
	test.AssertIntEqual(t, "GetLogLevelTypeForName:UNKNOWN", int(NotFound), int(GetLogLevelTypeIndexForLevelName("UNKNOWN")))
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
