package servermain

import (
	"webServerBase/test"
	"testing"
	"net/http"
)

func TestWithUrl(t *testing.T) {
	req, err := http.NewRequest("GET","http://abc:8080/data1/1/data2/2?A=5", nil)
	if (err != nil) {
		test.Fail(t, "", err.Error())
	}
	d := NewURLDetails(req)
	test.AssertEqualString(t, "", "data1/1/data2/2", d.GetURL())
	test.AssertEqualString(t, "", "data1", d.GetURLPart(0))
	test.AssertEqualString(t, "", "1", d.GetURLPart(1))
	test.AssertEqualString(t, "", "data2", d.GetURLPart(2))
	test.AssertEqualString(t, "", "2", d.GetURLPart(3))
	test.AssertEmptyString(t, "", d.GetURLPart(4))
	test.AssertEmptyString(t, "", d.GetURLPart(-1))

	test.AssertEqualString(t, "", "1", d.GetNamedPart("data1"))
	test.AssertEqualString(t, "", "2", d.GetNamedPart("data2"))
	test.AssertEmptyString(t, "", d.GetNamedPart("data3"))
	test.AssertEmptyString(t, "", d.GetNamedPart(""))

	test.AssertEqualString(t, "", "5", d.GetNamedQuery("A"))
	test.AssertEmptyString(t, "", d.GetNamedQuery("X"))
}
