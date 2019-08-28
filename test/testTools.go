package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"testing"
)

/*
Fail - Fail with a message and a StackTrace
*/
func Fail(t *testing.T, info string, message string) {
	logStackTraceAndFail(t, fmt.Sprintf("TEST FAILED: Error[%s]", message), info, debug.Stack())
}

/*
AssertNotError - Fail if error is not null. Logs error and the stack trace.
*/
func AssertNotError(t *testing.T, info string, err error) {
	if err == nil {
		return
	}
	logStackTraceAndFail(t, fmt.Sprintf("TEST FAILED: Error must be nil. Error[%s]", err.Error()), info, debug.Stack())
}

/*
AssertIsError - Fail if error is null. Logs error and the stack trace.
*/
func AssertIsError(t *testing.T, info string, err error) string {
	if err != nil {
		return err.Error()
	}
	logStackTraceAndFail(t, "TEST FAILED: Error was nil:", info, debug.Stack())
	return ""
}

/*
AssertNotNilErrorAndContains - Fail if error is null. Logs error and the stack trace.
*/
func AssertNotNilErrorAndContains(t *testing.T, info string, contains string, err error) string {
	if err != nil {
		text := err.Error()
		if contains != "" {
			if strings.Contains(text, contains) {
				return text
			}
			logStackTraceAndFail(t, fmt.Sprintf("TEST FAILED: Error text [%s] does NOT contain [%s]", text, contains), info, debug.Stack())
			return ""
		}
		return text
	}
	logStackTraceAndFail(t, "TEST FAILED: Error was nil", info, debug.Stack())
	return ""
}

/*
AssertEqualInt assert ints are equal
*/
func AssertEqualInt(t *testing.T, info string, expected int, actual int) {
	if expected != actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected %d actual %d", expected, actual), info, debug.Stack())
	}
}

/*
AssertNotEqualInt assert ints are NOT equal
*/
func AssertNotEqualInt(t *testing.T, info string, expected int, actual int) {
	if expected == actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected %d NOT actual %d", expected, actual), info, debug.Stack())
	}
}

/*
AssertNil assert object is (null) nil
*/
func AssertNil(t *testing.T, info string, expected interface{}) {
	if expected != nil {
		logStackTraceAndFail(t, fmt.Sprintf("Failed: Expected (%T) should be nil", expected), info, debug.Stack())
	}
}

/*
AssertNotNil assert object is NOT nil
*/
func AssertNotNil(t *testing.T, info string, expected interface{}) {
	if expected == nil {
		logStackTraceAndFail(t, "Expected value should NOT be nil", info, debug.Stack())
	}
}

/*
AssertEmptyString assert string has a value ""
*/
func AssertEmptyString(t *testing.T, info string, expected string) {
	if strings.TrimSpace(expected) != "" {
		logStackTraceAndFail(t, fmt.Sprintf("Expected should be empty. Not: %s", expected), info, debug.Stack())
	}
}

/*
AssertNotEmptyString assert string has a value, NOT ""
*/
func AssertNotEmptyString(t *testing.T, info string, expected string) {
	if strings.TrimSpace(expected) != "" {
		logStackTraceAndFail(t, "Expected string should NOT be empty", info, debug.Stack())
	}
}

/*
AssertTrue assert value is true
*/
func AssertTrue(t *testing.T, info string, actual bool) {
	if !actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected true actual %t", actual), info, debug.Stack())
	}
}

/*
AssertFalse assert value is true
*/
func AssertFalse(t *testing.T, info string, actual bool) {
	if actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected false actual %t", actual), info, debug.Stack())
	}
}

/*
AssertEqualString assert strings are equal
*/
func AssertEqualString(t *testing.T, info string, expected string, actual string) {
	if expected != actual {
		logStackTraceAndFail(t, fmt.Sprintf("Expected '%s' actual '%s'", expected, actual), info, debug.Stack())
	}
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
AssertEndsWithString assert strings are equal
*/
func AssertEndsWithString(t *testing.T, info string, value string, endsWithThis string) {
	if !strings.HasSuffix(value, endsWithThis) {
		logStackTraceAndFail(t, fmt.Sprintf("String '%s' does not end with '%s'", value, endsWithThis), info, debug.Stack())
	}
}

/*
RemoveFile assert strings are equal
*/
func RemoveFile(t *testing.T, info string, path string) {
	var err = os.Remove(path)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("File %s could not be deleted. Error:%s", path, err.Error()), info, debug.Stack())
	}
}

/*
AssertFileContains read a file ans see if any if the strings are contained in it
*/
func AssertFileContains(t *testing.T, info string, fileName string, contains []string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		logStackTraceAndFail(t, fmt.Sprintf("File %s could not read. Error:%s", fileName, err.Error()), info, debug.Stack())
	}
	str := string(b)
	for _, val := range contains {
		if !strings.Contains(str, val) {
			logStackTraceAndFail(t, fmt.Sprintf("File %s does not contain '%s'", fileName, val), info, debug.Stack())
		}
	}
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

func logStackTraceAndFail(t *testing.T, desc string, info string, bytes []byte) {
	t.Logf("FAILED:%s :%s", info, desc)
	for count, line := range strings.Split(strings.TrimSuffix(string(bytes), "\n"), "\n") {
		if count > 2 && count <= 10 {
			t.Logf("FAILED:%s :%s", info, line)
		}
	}
	t.Fail()
}
