package test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
)

/*
AppendToFile - Append text to a file. Create if not there
*/
func AppendToFile(t *testing.T, fileName string, text string) {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("Failed to open file [%s]: %s", fileName, err.Error()), "AppendToFile", debug.Stack())
		return
	}
	defer f.Close()
	_, err = f.WriteString(text)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("Failed to write to file [%s]: %s", fileName, err.Error()), "AppendToFile", debug.Stack())
		return
	}
}

/*
DeleteFile delete a file. Use with defer
*/
func DeleteFile(t *testing.T, fileName string, mustExist bool) {
	_, err := os.Stat(fileName)
	if err != nil {
		if mustExist {
			logStackTraceAndFail(t, fmt.Sprintf("Failed to remove file [%s] Could not be found!: %s", fileName, err.Error()), "DeleteFile", debug.Stack())
		}
		return
	}

	err = os.Remove(fileName)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("Failed to remove file [%s]: %s", fileName, err.Error()), "DeleteFile", debug.Stack())
		return
	}
}

/*
Fail - Fail with a message and a StackTrace
*/
func Fail(t *testing.T, info string, message string) {
	logStackTraceAndFail(t, fmt.Sprintf("Error[%s]", message), info, debug.Stack())
}

/*
AssertErrorIsNil - Fail if error is not null. Logs error and the stack trace.
*/
func AssertErrorIsNil(t *testing.T, info string, err error) {
	if err == nil {
		return
	}
	logStackTraceAndFail(t, fmt.Sprintf("Error must be nil. Error[%s]", err.Error()), info, debug.Stack())
}

/*
AssertError - Fail if error is null. Logs error and the stack trace.
*/
func AssertError(t *testing.T, info string, err error) string {
	if err != nil {
		return err.Error()
	}
	logStackTraceAndFail(t, "Error was nil:", info, debug.Stack())
	return ""
}

/*
AssertErrorString assert strings are equal
*/
func AssertErrorString(t *testing.T, info string, actual error, expected string) {
	if actual == nil {
		logStackTraceAndFail(t, "Expected error value should NOT be nil", info, debug.Stack())
	}
	if actual.Error() != expected {
		logStackTraceAndFail(t, fmt.Sprintf("Expected Error message '%s' actual Error message '%s'", expected, actual.Error()), info, debug.Stack())
	}
}

/*
AssertErrorIsNotExist assert that the error is a Not Found error
*/
func AssertErrorIsNotExist(t *testing.T, info string, err error) {
	if os.IsNotExist(err) {
		return
	}
	logStackTraceAndFail(t, fmt.Sprintf("Error [%s] is NOT a 'Not Exist' error ", err.Error()), info, debug.Stack())
}

/*
AssertErrorTextContains - Fail if error is null. Logs error and the stack trace.
*/
func AssertErrorTextContains(t *testing.T, info string, err error, contains string) string {
	if err != nil {
		text := err.Error()
		if contains != "" {
			if strings.Contains(text, contains) {
				return text
			}
			logStackTraceAndFail(t, fmt.Sprintf("Error text [%s] does NOT contain [%s]", text, contains), info, debug.Stack())
			return ""
		}
		return text
	}
	logStackTraceAndFail(t, "Error was nil", info, debug.Stack())
	return ""
}

/*
AssertInt64Equal assert ints are equal
*/
func AssertInt64Equal(t *testing.T, info string, actual int64, expected int64) {
	if expected != actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected %d actual %d", expected, actual), info, debug.Stack())
	}
}

/*
AssertIntEqual assert ints are equal
*/
func AssertIntEqual(t *testing.T, info string, actual int, expected int) {
	if expected != actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected %d actual %d", expected, actual), info, debug.Stack())
	}
}

/*
AssertIntNotEqual assert ints are NOT equal
*/
func AssertIntNotEqual(t *testing.T, info string, actual int, expected int) {
	if expected == actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected %d NOT actual %d", expected, actual), info, debug.Stack())
	}
}

/*
AssertNil assert object is (null) nil
*/
func AssertNil(t *testing.T, info string, expected interface{}) {
	if expected != nil {
		if reflect.ValueOf(expected).IsNil() {
			return
		}
	}
	logStackTraceAndFail(t, fmt.Sprintf("Expected (%T) should be nil", expected), info, debug.Stack())
}

/*
AssertNilNot assert object is NOT nil
*/
func AssertNilNot(t *testing.T, info string, expected interface{}) {
	if expected == nil {
		logStackTraceAndFail(t, "Expected value should NOT be nil", info, debug.Stack())
	}
	if reflect.ValueOf(expected).IsNil() {
		logStackTraceAndFail(t, "Expected value should NOT be nil", info, debug.Stack())
	}

}

/*
AssertStringEmpty assert string has a value ""
*/
func AssertStringEmpty(t *testing.T, info string, expected string) {
	if strings.TrimSpace(expected) != "" {
		logStackTraceAndFail(t, fmt.Sprintf("Expected should be empty. Not: %s", expected), info, debug.Stack())
	}
}

/*
AssertStringNotEmpty assert string has a value, NOT ""
*/
func AssertStringNotEmpty(t *testing.T, info string, expected string) {
	if strings.TrimSpace(expected) != "" {
		logStackTraceAndFail(t, "Expected string should NOT be empty", info, debug.Stack())
	}
}

/*
AssertBoolTrue assert value is true
*/
func AssertBoolTrue(t *testing.T, info string, actual bool) {
	if !actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected true actual %t", actual), info, debug.Stack())
	}
}

/*
AssertBoolFalse assert value is true
*/
func AssertBoolFalse(t *testing.T, info string, actual bool) {
	if actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected false actual %t", actual), info, debug.Stack())
	}
}

/*
AssertStringEquals assert strings are equal
*/
func AssertStringEquals(t *testing.T, info string, actual string, expected string) {
	if expected != actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected '%s' actual '%s'", expected, actual), info, debug.Stack())
	}
}

/*
AssertStringEqualsUnix assert strings are equal ignoring cr-lf inconsistencies in the OS
*/
func AssertStringEqualsUnix(t *testing.T, info string, actual string, expected string) {
	if cleanStr(expected) != cleanStr(actual) {
		logStackTraceAndFail(t, fmt.Sprintf("Expected '%s' actual '%s'", cleanStr(expected), cleanStr(actual)), info, debug.Stack())
	}
}

func cleanStr(str string) string {
	var o bytes.Buffer
	s := []byte(str)
	for i := 0; i < len(s); i++ {
		if s[i] != 13 {
			if s[i] < 32 {
				o.WriteString(fmt.Sprintf("[%d]", s[i]))
			} else {
				o.Write(s[i : i+1])
			}
		}
	}
	return o.String()
}

/*
AssertStringContains see if all the strings are contained in the string
*/
func AssertStringContains(t *testing.T, info string, content string, contains ...string) {
	for _, val := range contains {
		if !strings.Contains(content, val) {
			logStackTraceAndFail(t, fmt.Sprintf("String '%s' does not contain '%s'", content, val), info, debug.Stack())
		}
	}
}

/*
AssertStringDoesNotContain see if all the strings are contained in the string
*/
func AssertStringDoesNotContain(t *testing.T, info string, content string, contains ...string) {
	for _, val := range contains {
		if strings.Contains(content, val) {
			logStackTraceAndFail(t, fmt.Sprintf("String '%s' contains '%s'", content, val), info, debug.Stack())
		}
	}
}

/*
AssertTypeEquals assert strings are equal
*/
func AssertTypeEquals(t *testing.T, info string, actual interface{}, expectedTypeName string) {
	actualType := fmt.Sprintf("%T", actual)
	if expectedTypeName != actualType {
		logStackTraceAndFail(t, fmt.Sprintf("Expected type '%s' actual type '%s'", expectedTypeName, actualType), info, debug.Stack())
	}
}

/*
AssertStringEndsWith assert strings are equal
*/
func AssertStringEndsWith(t *testing.T, info string, value string, endsWithThis string) {
	if !strings.HasSuffix(value, endsWithThis) {
		logStackTraceAndFail(t, fmt.Sprintf("String '%s' does not end with '%s'", value, endsWithThis), info, debug.Stack())
	}
}

/*
AssertFileExists assert strings are equal
*/
func AssertFileExists(t *testing.T, info string, path string) {
	_, err := os.Stat(path)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("File '%s' does not exist. Error:%s", path, err.Error()), info, debug.Stack())
	}
}

/*
AssertFileNotExists assert strings are equal
*/
func AssertFileNotExists(t *testing.T, info string, path string) {
	_, err := os.Stat(path)
	if err == nil {
		logStackTraceAndFail(t, fmt.Sprintf("File '%s' should not exist. Error:%s", path, err.Error()), info, debug.Stack())
	}
}

/*
AssertFileRemoved assert strings are equal
*/
func AssertFileRemoved(t *testing.T, info string, path string) {
	var err = os.Remove(path)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("File '%s' could not be deleted. Error:%s", path, err.Error()), info, debug.Stack())
	}
	t.Logf("Remove File:%s '%s'", info, path)
}

/*
AssertFileContains see if all the strings are contained in the file
*/
func AssertFileContains(t *testing.T, info string, fileName string, contains ...string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("File '%s' could not read. Error:%s", fileName, err.Error()), info, debug.Stack())
	}
	for _, val := range contains {
		if !strings.Contains(string(b), val) {
			logStackTraceAndFail(t, fmt.Sprintf("File '%s' does not contain the string '%s'", fileName, val), info, debug.Stack())
		}
	}
	AssertStringContains(t, info, string(b), contains...)
}

/*
AssertFileDoesNotContain read a file ans see if any if the strings are contained in it
*/
func AssertFileDoesNotContain(t *testing.T, info string, fileName string, contains ...string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("File %s could not read. Error:%s", fileName, err.Error()), info, debug.Stack())
	}
	str := string(b)
	for _, val := range contains {
		if strings.Contains(str, val) {
			logStackTraceAndFail(t, fmt.Sprintf("File %s contains '%s'", fileName, val), info, debug.Stack())
		}
	}
}

/*
AssertPanicAndRecover - Called via a defer in tests that require a panic containing specific text to be thrown
*/
func AssertPanicAndRecover(t *testing.T, contains string) {
	rec := recover()
	if rec != nil {
		recText := fmt.Sprintf("%s", rec)
		if strings.Contains(recText, contains) {
			return
		}
		Fail(t, "", fmt.Sprintf("AssertPanicAndRecover: Panic message '%s' does not contain '%s'", recText, contains))
	}
	Fail(t, "", fmt.Sprintf("AssertPanicAndRecover: A 'panic' containing the text '%s' was NOT thrown!", contains))
}

func logStackTraceAndFail(t *testing.T, desc string, info string, bytes []byte) {
	t.Logf("TEST-FAILED:%s :%s", info, desc)
	for count, line := range strings.Split(strings.TrimSuffix(string(bytes), "\n"), "\n") {
		if count > 2 && count <= 10 {
			t.Logf("TEST-FAILED:%s :%s", info, line)
		}
	}
	t.Fail()
}
