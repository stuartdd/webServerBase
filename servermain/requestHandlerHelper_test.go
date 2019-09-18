package servermain

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
	"webServerBase/test"
)

type TestStruct struct {
	Name    string
	Types   []string
	ID      int `json:"ref"`
	Created time.Time
}

var jsonTestData = []byte(`
{
    "Name": "Fruit",
    "Types": [
        "Apple",
        "Banana",
        "Orange"
    ],
    "ref": 999,
    "Created": "2018-04-09T23:00:00Z"
}`)

func TestWithBodyJsonObject(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", bytes.NewReader(jsonTestData))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))

	testStruct := &TestStruct{}
	d.GetJSONBodyAsObject(testStruct)
	test.AssertStringEquals(t, "", "Fruit", testStruct.Name)
	test.AssertIntEqual(t, "", 999, testStruct.ID)
	test.AssertIntEqual(t, "", 3, len(testStruct.Types))
	test.AssertStringEquals(t, "", "Apple", testStruct.Types[0])
	test.AssertStringEquals(t, "", "Banana", testStruct.Types[1])
	test.AssertStringEquals(t, "", "Orange", testStruct.Types[2])
	test.AssertStringEquals(t, "", "2018-04-09T23:00:00", testStruct.Created.Format("2006-01-02T15:04:05"))
}

func TestWithBodyJsonList(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("[\"TEST\",\"VALUE\"]"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	aList := d.GetJSONBodyAsList()
	test.AssertTypeEquals(t, "", "[]interface {}", aList)
	test.AssertTypeEquals(t, "", "string", aList[0])
	test.AssertStringEquals(t, "", "TEST", aList[0].(string))
	test.AssertStringEquals(t, "", "VALUE", aList[1].(string))
}

func TestWithBodyJsonListPanic(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("[TEST\",\"VALUE\"]"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	defer test.AssertPanicAndRecover(t, "E|400|"+strconv.Itoa(SCInvalidJSONRequest)+"|Invalid JSON")
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	d.GetJSONBodyAsList()
}

func TestWithBodyJsonMap(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("{\"TEST\":\"VALUE\"}"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	aMap := d.GetJSONBodyAsMap()
	val := aMap["TEST"]
	test.AssertTypeEquals(t, "", "string", val)
	test.AssertStringEquals(t, "", "VALUE", val.(string))
}

func TestWithBodyJsonMapPanic(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("{\"TEST\"\"VALUE\"}"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	defer test.AssertPanicAndRecover(t, "E|400|"+strconv.Itoa(SCInvalidJSONRequest)+"|Invalid JSON")
	d.GetJSONBodyAsMap()
}

func TestWithBodyText(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("TEST"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	text := d.GetBodyString()
	test.AssertStringEquals(t, "", text, "TEST")
}

func TestGetURLPartPanics(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", nil)
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	defer test.AssertPanicAndRecover(t, "URL parameter at position '4' returned an empty value")
	d.GetURLPart(4, "")
}

func TestGetNamedPartPanics(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", nil)
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	defer test.AssertPanicAndRecover(t, "URL parameter 'XXX' returned an empty value")
	d.GetNamedURLPart("XXX", "")
}

func TestWithUrl(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", nil)
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	test.AssertStringEquals(t, "", "data1/1/data2/2", d.GetURL())
	test.AssertStringEquals(t, "", "data1", d.GetURLPart(0, ""))
	test.AssertStringEquals(t, "", "1", d.GetURLPart(1, ""))
	test.AssertStringEquals(t, "", "data2", d.GetURLPart(2, ""))
	test.AssertStringEquals(t, "", "2", d.GetURLPart(3, ""))
	test.AssertStringEquals(t, "", "Z", d.GetURLPart(4, "Z"))
	test.AssertStringEquals(t, "", "X", d.GetURLPart(-4, "X"))

	test.AssertStringEquals(t, "", "1", d.GetNamedURLPart("data1", ""))
	test.AssertStringEquals(t, "", "2", d.GetNamedURLPart("data2", ""))

	test.AssertStringEquals(t, "", "ZZ", d.GetNamedURLPart("", "ZZ"))
	test.AssertStringEquals(t, "", "ZZ", d.GetNamedURLPart("123", "ZZ"))

	test.AssertStringEquals(t, "", "5", d.GetNamedQuery("A"))

	test.AssertIntEqual(t, "", 4, d.GetPartsCount())

	d2 := NewRequestHandlerHelper(req, NewResponse(nil, nil))
	test.AssertIntEqual(t, "", 4, d2.GetPartsCount())

}
