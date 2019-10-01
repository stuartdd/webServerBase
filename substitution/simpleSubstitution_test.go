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
	test.AssertStringEquals(t, "", "x1x", ReplaceDollar("x${a}x", ma, '$'))
	test.AssertStringEquals(t, "", "x1x ${b}", ReplaceDollar("x${a}x ${b}", ma, '$'))
	ma["b"] = "2"
	test.AssertStringEquals(t, "", "x1x 2", ReplaceDollar("x${a}x ${b}", ma, '$'))
	ma["b"] = "2"
	test.AssertStringEquals(t, "", "x1x 2 "+os.Getenv("GOPATH"), ReplaceDollar("x${a}x ${b} ${GOPATH}", ma, '$'))

	test.AssertStringEquals(t, "", "x1x 2 ${yyyy}", ReplaceDollar("x${a}x ${b} ${yyyy}", ma, '$'))
	yyyy, _, _ := time.Now().Date()
	test.AssertStringDoesNotContain(t, "", ReplaceDollar("x${a}x ${b} "+strconv.Itoa(yyyy), ma, '$'), []string{"${YYYY}"})

}

func TestReplaceNotFound(t *testing.T) {
	m := make(map[string]string)
	test.AssertStringEquals(t, "", "x${}x", ReplaceDollar("x${}x", m, '$'))
	test.AssertStringEquals(t, "", "x${}", ReplaceDollar("x${}", m, '$'))
	test.AssertStringEquals(t, "", "x${a}x", ReplaceDollar("x${a}x", m, '$'))
	test.AssertStringEquals(t, "", "x${a}", ReplaceDollar("x${a}", m, '$'))
	test.AssertStringEquals(t, "", "x${a{", ReplaceDollar("x${a{", m, '$'))
	test.AssertStringEquals(t, "", "x${a", ReplaceDollar("x${a", m, '$'))
	test.AssertStringEquals(t, "", "x${", ReplaceDollar("x${", m, '$'))
	test.AssertStringEquals(t, "", "x$a", ReplaceDollar("x$a", m, '$'))
	test.AssertStringEquals(t, "", "x$", ReplaceDollar("x$", m, '$'))
	test.AssertStringEquals(t, "", "x{}", ReplaceDollar("x{}", m, '$'))
	test.AssertStringEquals(t, "", "x{", ReplaceDollar("x{", m, '$'))
	test.AssertStringEquals(t, "", "x", ReplaceDollar("x", m, '$'))
	test.AssertStringEquals(t, "", "", ReplaceDollar("", m, '$'))
	yyyy, _, _ := time.Now().Date()
	test.AssertStringEquals(t, "", "x${}x "+strconv.Itoa(yyyy), ReplaceDollar("x${}x ${YYYY}", nil, '$'))

}
