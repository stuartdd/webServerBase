package exec

import (
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stuartdd/webServerBase/test"
)

var callbackNotDone = true
var waitcount = 5
var callBackResult *CmdStatus
var testOS = runtime.GOOS

func TestRunErr(t *testing.T) {
	x := RunAndWait("", "sh", nil, "-c", "ls fred")
	test.AssertError(t, "", x.Err)
	test.AssertIntEqual(t, "", x.RetCode, 2)
	test.AssertStringContains(t, "", x.Stderr, "cannot access 'fred'")
	test.AssertStringEmpty(t, "", x.Stdout)
	test.AssertErrorTextContains(t, "", x.Err, "exit status 2")
}

func TestRunErrWithPath(t *testing.T) {
	x := RunAndWait("fred", "sh", nil, "-c", "ls")
	test.AssertError(t, "", x.Err)
	test.AssertIntEqual(t, "", x.RetCode, 1)
	test.AssertStringEmpty(t, "", x.Stderr)
	test.AssertStringEmpty(t, "", x.Stdout)
	test.AssertErrorTextContains(t, "", x.Err, "Path [fred] does not exist")
}

func TestRunOk(t *testing.T) {
	x := RunAndWait("", "sh", nil, "-c", "echo stdout; echo 1>&2 stderr")
	test.AssertErrorIsNil(t, "", x.Err)
	test.AssertIntEqual(t, "", x.RetCode, 0)
	test.AssertStringEquals(t, "", "stderr", x.Stderr)
	test.AssertStringEquals(t, "", "stdout", x.Stdout)
}

func TestRunDIR(t *testing.T) {
	if testOS == "windows" {
		x := RunAndWait("", "cmd", nil, "/C", "dir", "c:\\Program Files")
		test.AssertErrorIsNil(t, "", x.Err)
		test.AssertIntEqual(t, "", x.RetCode, 0)
		test.AssertStringEmpty(t, "", x.Stderr)
		test.AssertStringContains(t, "", x.Stdout, "<DIR>", "Directory of c:\\Program Files")
	}
}

func TestRunWithPath(t *testing.T) {
	if testOS == "windows" {
		x := RunAndWait("c:\\Program Files", "cmd", nil, "/C", "dir")
		test.AssertErrorIsNil(t, "", x.Err)
		test.AssertIntEqual(t, "", x.RetCode, 0)
		test.AssertStringEmpty(t, "", x.Stderr)
		test.AssertStringContains(t, "", x.Stdout, "<DIR>", "Directory of c:\\Program Files")
	}
}

func TestRunDIRRunBackground(t *testing.T) {
	if testOS == "windows" {
		callbackNotDone = true
		waitcount = 0
		RunAndCallback(callbackFunction, "", "cmd", nil, "/C", "dir", "c:\\Program Files")
		for callbackNotDone {
			time.Sleep(100 * time.Millisecond)
			waitcount++
			if waitcount > 15 {
				test.Fail(t, "", "Failed: (Took too long!) Waitcount is "+strconv.Itoa(waitcount))
				return
			}
		}
		if waitcount < 9 {
			test.Fail(t, "", "Failed: (came back too soon!) Waitcount is "+strconv.Itoa(waitcount))
			return
		}
		test.AssertErrorIsNil(t, "", callBackResult.Err)
		test.AssertIntEqual(t, "", callBackResult.RetCode, 0)
		test.AssertStringEmpty(t, "", callBackResult.Stderr)
		test.AssertStringContains(t, "", callBackResult.Stdout, "<DIR>")
	}
}

func callbackFunction(cs *CmdStatus) {
	for waitcount < 10 {
		time.Sleep(100 * time.Millisecond)
	}
	callbackNotDone = false
	callBackResult = cs
}
