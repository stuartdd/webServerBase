package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
)

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
func AssertErrorString(t *testing.T, info string, expected string, actual error) {
	if actual == nil {
		logStackTraceAndFail(t, "Expected error value should NOT be nil", info, debug.Stack())
	}
	if actual.Error() != expected {
		logStackTraceAndFail(t, fmt.Sprintf("Expected Error message '%s' actual Error message '%s'", expected, actual.Error()), info, debug.Stack())
	}
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
AssertIntEqual assert ints are equal
*/
func AssertIntEqual(t *testing.T, info string, expected int, actual int) {
	if expected != actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected %d actual %d", expected, actual), info, debug.Stack())
	}
}

/*
AssertIntNotEqual assert ints are NOT equal
*/
func AssertIntNotEqual(t *testing.T, info string, expected int, actual int) {
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
func AssertStringEquals(t *testing.T, info string, expected string, actual string) {
	if expected != actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected '%s' actual '%s'", expected, actual), info, debug.Stack())
	}
}

/*
AssertStringContains see if all the strings are contained in the string
*/
func AssertStringContains(t *testing.T, info string, content string, contains []string) {
	for _, val := range contains {
		if !strings.Contains(content, val) {
			logStackTraceAndFail(t, fmt.Sprintf("String '%s' does not contain '%s'", content, val), info, debug.Stack())
		}
	}
}

/*
AssertStringNotContain see if all the strings are contained in the string
*/
func AssertStringNotContain(t *testing.T, info string, content string, contains []string) {
	for _, val := range contains {
		if strings.Contains(content, val) {
			logStackTraceAndFail(t, fmt.Sprintf("String '%s' contains '%s'", content, val), info, debug.Stack())
		}
	}
}

/*
AssertTypeEquals assert strings are equal
*/
func AssertTypeEquals(t *testing.T, info string, expectedTypeName string, actual interface{}) {
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
func AssertFileContains(t *testing.T, info string, fileName string, contains []string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("File '%s' could not read. Error:%s", fileName, err.Error()), info, debug.Stack())
	}
	for _, val := range contains {
		if !strings.Contains(string(b), val) {
			logStackTraceAndFail(t, fmt.Sprintf("File '%s' does not contain the string '%s'", fileName, val), info, debug.Stack())
		}
	}
	AssertStringContains(t, info, string(b), contains)
}

/*
AssertFileDoesNotContain read a file ans see if any if the strings are contained in it
*/
func AssertFileDoesNotContain(t *testing.T, info string, fileName string, contains []string) {
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
