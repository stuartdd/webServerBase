package servermain

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
	x := RunAndWait("", "sh", "-c", "ls fred")
	test.AssertError(t, "", x.err)
	test.AssertIntEqual(t, "", 2, x.retCode)
	test.AssertStringContains(t, "", x.stderr, []string{"cannot access 'fred'"})
	test.AssertStringEmpty(t, "", x.stdout)
	test.AssertErrorTextContains(t, "", x.err, "exit status 2")
}

func TestRunErrWithPath(t *testing.T) {
	x := RunAndWait("fred", "sh", "-c", "ls")
	test.AssertError(t, "", x.err)
	test.AssertIntEqual(t, "", 1, x.retCode)
	test.AssertStringEmpty(t, "", x.stderr)
	test.AssertStringEmpty(t, "", x.stdout)
	test.AssertErrorTextContains(t, "", x.err, "Path [fred] does not exist")
}

func TestRunOk(t *testing.T) {
	x := RunAndWait("", "sh", "-c", "echo stdout; echo 1>&2 stderr")
	test.AssertErrorIsNil(t, "", x.err)
	test.AssertIntEqual(t, "", 0, x.retCode)
	test.AssertStringEquals(t, "", x.stderr, "stderr")
	test.AssertStringEquals(t, "", x.stdout, "stdout")
}

func TestRunDIR(t *testing.T) {
	if testOS == "windows" {
		x := RunAndWait("", "cmd", "/C", "dir", "c:\\Program Files")
		test.AssertErrorIsNil(t, "", x.err)
		test.AssertIntEqual(t, "", 0, x.retCode)
		test.AssertStringEmpty(t, "", x.stderr)
		test.AssertStringContains(t, "", x.stdout, []string{"<DIR>", "Directory of c:\\Program Files"})
	}
}

func TestRunWithPath(t *testing.T) {
	if testOS == "windows" {
		x := RunAndWait("c:\\Program Files", "cmd", "/C", "dir")
		test.AssertErrorIsNil(t, "", x.err)
		test.AssertIntEqual(t, "", 0, x.retCode)
		test.AssertStringEmpty(t, "", x.stderr)
		test.AssertStringContains(t, "", x.stdout, []string{"<DIR>", "Directory of c:\\Program Files"})
	}
}

func TestRunDIRRunBackground(t *testing.T) {
	if testOS == "windows" {
		callbackNotDone = true
		waitcount = 0
		RunAndCallback(callbackFunction, "", "cmd", "/C", "dir", "c:\\Program Files")
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
		test.AssertErrorIsNil(t, "", callBackResult.err)
		test.AssertIntEqual(t, "", 0, callBackResult.retCode)
		test.AssertStringEmpty(t, "", callBackResult.stderr)
		test.AssertStringContains(t, "", callBackResult.stdout, []string{"<DIR>"})
	}
}

func callbackFunction(cs *CmdStatus) {
	for waitcount < 10 {
		time.Sleep(100 * time.Millisecond)
	}
	callbackNotDone = false
	callBackResult = cs
}
