package servermain

import (
	"testing"

	"github.com/stuartdd/webServerBase/test"
)

func TestRunErr(t *testing.T) {
	x := Run("sh", "-c", "ls fred")
	test.AssertError(t, "", x.err)
	test.AssertIntEqual(t, "", 2, x.retCode)
	test.AssertStringContains(t, "", x.stderr, []string{"cannot access 'fred'"})
	test.AssertStringEmpty(t, "", x.stdout)
	test.AssertErrorTextContains(t, "", x.err, "exit status 2")
}

func TestRunOk(t *testing.T) {
	x := Run("sh", "-c", "echo stdout; echo 1>&2 stderr")
	test.AssertErrorIsNil(t, "", x.err)
	test.AssertIntEqual(t, "", 0, x.retCode)
	test.AssertStringEquals(t, "", x.stderr, "stderr")
	test.AssertStringEquals(t, "", x.stdout, "stdout")
}
