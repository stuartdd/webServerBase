package test

import (
	"io/ioutil"
	"strings"
	"testing"
)

/*
AssertEqualInt assert ints are equal
*/
func AssertEqualInt(t *testing.T, message string, expected int, actual int) {
	if expected != actual {
		t.Fatalf("Failed: Expected %d actual %d - %s", expected, actual, message)
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
AssertEndsWithString assert strings are equal
*/
func AssertEndsWithString(t *testing.T, message string, value string, endsWithThis string) {
	if !strings.HasSuffix(value, endsWithThis) {
		t.Fatalf("Failed: '%s' does not end with '%s' - %s", value, endsWithThis, message)
	}
}

/*
AssertFileContains read a file ans see if any if the strings are contained in it
*/
func AssertFileContains(t *testing.T, message string, fileName string, contains []string) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Failed: file %s could not be read - %s", fileName, message)
	}
	str := string(b)
	for _, val := range contains {
		if !strings.Contains(str, val) {
			t.Fatalf("Failed: file %s does not contain '%s' - %s", fileName, val, message)
		}
	}
}
