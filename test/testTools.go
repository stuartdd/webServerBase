package test

import (
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"testing"
)

/*
FailIfError - Fail if error is not null. Logs error and the stack trace.
*/
func FailIfError(t *testing.T, info string, err error) {
	if err == nil {
		return
	}
	t.Logf("TEST FAILED: Error[%s] %s", err.Error(), info)
	st := string(debug.Stack())
	for count, line := range strings.Split(strings.TrimSuffix(st, "\n"), "\n") {
		if count > 2 && count <= 10 {
			t.Logf("TEST FAILED:%s", line)
		}
	}
	t.Fail()
}

/*
FailIfNilError - Fail if error is null. Logs error and the stack trace.
*/
func FailIfNilErrorAndContains(t *testing.T, info string, contains string, err error) string {
	if err != nil {
		text := err.Error()
		if contains != "" {
			if strings.Contains(text, contains) {
				return text
			}
			t.Logf("TEST FAILED: Error text [%s] did NOT contain [%s]: %s", text, contains, info)
			t.Fail()
			return text		
		}
		return text
	}
	t.Logf("TEST FAILED: Error was nil: %s", info)
	st := string(debug.Stack())
	for count, line := range strings.Split(strings.TrimSuffix(st, "\n"), "\n") {
		if count > 2 && count <= 10 {
			t.Logf("TEST FAILED:%s :%s", info, line)
		}
	}
	t.Fail()
	return ""
}

/*
FailIfNilError - Fail if error is null. Logs error and the stack trace.
*/
func FailIfNilError(t *testing.T, info string, err error) string {
	if err != nil {
		return err.Error()
	}
	t.Logf("TEST FAILED: Error was nil: %s", info)
	st := string(debug.Stack())
	for count, line := range strings.Split(strings.TrimSuffix(st, "\n"), "\n") {
		if count > 2 && count <= 10 {
			t.Logf("TEST FAILED:%s :%s", info, line)
		}
	}
	t.Fail()
	return ""
}

/*
AssertEqualInt assert ints are equal
*/
func AssertEqualInt(t *testing.T, message string, expected int, actual int) {
	if expected != actual {
		t.Fatalf("Failed: Expected %d actual %d - %s", expected, actual, message)
	}
}

/*
AssertEqualInt assert ints are equal
*/
func AssertNotEqualInt(t *testing.T, message string, expected int, actual int) {
	if expected == actual {
		t.Fatalf("Failed: Actual value %d should NOT equal %d - %s", actual, expected, message)
	}
}

/*
AssertNil assert object is (null) nil
*/
func AssertNil(t *testing.T, message string, expected interface{}) {
	if expected != nil {
		t.Fatalf("Failed: Expected (%T) should be nil %s", expected, message)
	}
}

/*
AssertNotNil assert object is NOT (null) nil
*/
func AssertNotNil(t *testing.T, message string, expected interface{}) {
	if expected == nil {
		t.Fatalf("Failed: Expected value should NOT be nil %s", message)
	}
}

/*
AssertEmptyString assert string has a value ""
*/
func AssertEmptyString(t *testing.T, message string, expected string) {
	if expected != "" {
		t.Fatalf("Failed: Expected should be \"\" %s", message)
	}
}

/*
AssertNotEmptyString assert string has a value, NOT ""
*/
func AssertNotEmptyString(t *testing.T, message string, expected string) {
	if expected != "" {
		t.Fatalf("Failed: Expected should be \"\" %s", message)
	}
}

/*
AssertTrue assert value is true
*/
func AssertTrue(t *testing.T, message string, actual bool) {
	if !actual {
		t.Fatalf("Failed: Expected true actual %t - %s", actual, message)
	}
}

/*
AssertFalse assert value is true
*/
func AssertFalse(t *testing.T, message string, actual bool) {
	if actual {
		t.Fatalf("Failed: Expected false actual %t - %s", actual, message)
	}
}

/*
AssertEqualString assert strings are equal
*/
func AssertEqualString(t *testing.T, message string, expected string, actual string) {
	if expected != actual {
		t.Fatalf("Failed: Expected '%s' actual '%s' - %s", expected, actual, message)
	}
}

/*
AssertErrorString assert strings are equal
*/
func AssertErrorString(t *testing.T, message string, expected error, actual string) {
	if expected == nil {
		t.Fatalf("Failed: Expected error value should NOT be nil %s", message)
	}
	if expected.Error() != actual {
		t.Fatalf("Failed: Error message. Expected '%s' actual '%s' - %s", expected.Error(), actual, message)
	}
}

/*
AssertEndsWithString assert strings are equal
*/
func AssertEndsWithString(t *testing.T, message string, value string, endsWithThis string) {
	if !strings.HasSuffix(value, endsWithThis) {
		t.Fatalf("Failed: '%s' does not end with '%s' - %s", value, endsWithThis, message)
	}
}

/*
RemoveFile assert strings are equal
*/
func RemoveFile(t *testing.T, message string, path string) {
	var err = os.Remove(path)
	if err != nil {
		t.Fatalf("Failed: file %s could not be deleted - %s Error:%s", path, message, err.Error())
	}
}

/*
AssertFileContains read a file ans see if any if the strings are contained in it
*/
func AssertFileContains(t *testing.T, message string, fileName string, contains []string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Failed: file %s could not be read - %s Error:%s", fileName, message, err.Error())
	}
	str := string(b)
	for _, val := range contains {
		if !strings.Contains(str, val) {
			t.Fatalf("Failed: file %s does not contain '%s' - %s", fileName, val, message)
		}
	}
}

/*
AssertFileDoesNotContain read a file ans see if any if the strings are contained in it
*/
func AssertFileDoesNotContain(t *testing.T, message string, fileName string, contains []string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Failed: file %s could not be read - %s Error:%s", fileName, message, err.Error())
	}
	str := string(b)
	for _, val := range contains {
		if strings.Contains(str, val) {
			t.Fatalf("Failed: file %s contains '%s' - %s", fileName, val, message)
		}
	}
}
