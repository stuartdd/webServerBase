package substitution

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stuartdd/webServerBase/test"
)

func TestReplaceFound(t *testing.T) {
	ma := make(map[string]string)
	ma["a"] = "1"
	test.AssertStringEquals(t, "", "x1x", DoSubstitution("x${a}x", ma, '$'))
	test.AssertStringEquals(t, "", "x1x ${b}", DoSubstitution("x${a}x ${b}", ma, '$'))
	ma["b"] = "2"
	test.AssertStringEquals(t, "", "x1x 2", DoSubstitution("x${a}x ${b}", ma, '$'))
	ma["b"] = "2"
	test.AssertStringEquals(t, "", "x1x 2 "+os.Getenv("GOPATH"), DoSubstitution("x${a}x ${b} ${GOPATH}", ma, '$'))

	test.AssertStringEquals(t, "", "x1x 2 ${yyyy}", DoSubstitution("x${a}x ${b} ${yyyy}", ma, '$'))
	yyyy, _, _ := time.Now().Date()
	test.AssertStringDoesNotContain(t, "", DoSubstitution("x${a}x ${b} "+strconv.Itoa(yyyy), ma, '$'), []string{"${YYYY}"})

}

func TestReplaceNotFound(t *testing.T) {
	m := make(map[string]string)
	test.AssertStringEquals(t, "", "x${}x", DoSubstitution("x${}x", m, '$'))
	test.AssertStringEquals(t, "", "x${}", DoSubstitution("x${}", m, '$'))
	test.AssertStringEquals(t, "", "x${a}x", DoSubstitution("x${a}x", m, '$'))
	test.AssertStringEquals(t, "", "x${a}", DoSubstitution("x${a}", m, '$'))
	test.AssertStringEquals(t, "", "x${a{", DoSubstitution("x${a{", m, '$'))
	test.AssertStringEquals(t, "", "x${a", DoSubstitution("x${a", m, '$'))
	test.AssertStringEquals(t, "", "x${", DoSubstitution("x${", m, '$'))
	test.AssertStringEquals(t, "", "x$a", DoSubstitution("x$a", m, '$'))
	test.AssertStringEquals(t, "", "x$", DoSubstitution("x$", m, '$'))
	test.AssertStringEquals(t, "", "x{}", DoSubstitution("x{}", m, '$'))
	test.AssertStringEquals(t, "", "x{", DoSubstitution("x{", m, '$'))
	test.AssertStringEquals(t, "", "x", DoSubstitution("x", m, '$'))
	test.AssertStringEquals(t, "", "", DoSubstitution("", m, '$'))
	yyyy, _, _ := time.Now().Date()
	test.AssertStringEquals(t, "", "x${}x "+strconv.Itoa(yyyy), DoSubstitution("x${}x ${YYYY}", nil, '$'))

}
