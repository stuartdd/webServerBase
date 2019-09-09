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
	d := NewRequestTools(req)

	testStruct := &TestStruct{}
	d.GetJSONBodyAsObject(testStruct)
	test.AssertEqualString(t, "", "Fruit", testStruct.Name)
	test.AssertEqualInt(t, "", 999, testStruct.ID)
	test.AssertEqualInt(t, "", 3, len(testStruct.Types))
	test.AssertEqualString(t, "", "Apple", testStruct.Types[0])
	test.AssertEqualString(t, "", "Banana", testStruct.Types[1])
	test.AssertEqualString(t, "", "Orange", testStruct.Types[2])
	test.AssertEqualString(t, "", "2018-04-09T23:00:00", testStruct.Created.Format("2006-01-02T15:04:05"))
}

func TestWithBodyJsonList(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("[\"TEST\",\"VALUE\"]"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	aList := d.GetJSONBodyAsList()
	test.AssertInterfaceType(t, "", "[]interface {}", aList)
	test.AssertInterfaceType(t, "", "string", aList[0])
	test.AssertEqualString(t, "", "TEST", aList[0].(string))
	test.AssertEqualString(t, "", "VALUE", aList[1].(string))
}

func TestWithBodyJsonListPanic(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("[TEST\",\"VALUE\"]"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	defer test.AssertPanicThrown(t, "E|400|"+strconv.Itoa(SCInvalidJSONRequest)+"|Invalid JSON")
	d := NewRequestTools(req)
	d.GetJSONBodyAsList()
}

func TestWithBodyJsonMap(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("{\"TEST\":\"VALUE\"}"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	aMap := d.GetJSONBodyAsMap()
	val := aMap["TEST"]
	test.AssertInterfaceType(t, "", "string", val)
	test.AssertEqualString(t, "", "VALUE", val.(string))
}

func TestWithBodyJsonMapPanic(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("{\"TEST\"\"VALUE\"}"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	defer test.AssertPanicThrown(t, "E|400|"+strconv.Itoa(SCInvalidJSONRequest)+"|Invalid JSON")
	d.GetJSONBodyAsMap()
}

func TestWithBodyText(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", strings.NewReader("TEST"))
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	text := d.GetBodyString()
	test.AssertEqualString(t, "", text, "TEST")
}

func TestGetURLPartPanics(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", nil)
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	defer test.AssertPanicThrown(t, "URL parameter at position '4' returned an empty value")
	d.GetURLPart(4, "")
}

func TestGetNamedPartPanics(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", nil)
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	defer test.AssertPanicThrown(t, "URL parameter 'XXX' returned an empty value")
	d.GetNamedURLPart("XXX", "")
}

func TestWithUrl(t *testing.T) {
	req, err := http.NewRequest("GET", "http://abc:8080/data1/1/data2/2?A=5", nil)
	if err != nil {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	test.AssertEqualString(t, "", "data1/1/data2/2", d.GetURL())
	test.AssertEqualString(t, "", "data1", d.GetURLPart(0, ""))
	test.AssertEqualString(t, "", "1", d.GetURLPart(1, ""))
	test.AssertEqualString(t, "", "data2", d.GetURLPart(2, ""))
	test.AssertEqualString(t, "", "2", d.GetURLPart(3, ""))
	test.AssertEqualString(t, "", "Z", d.GetURLPart(4, "Z"))
	test.AssertEqualString(t, "", "X", d.GetURLPart(-4, "X"))

	test.AssertEqualString(t, "", "1", d.GetNamedURLPart("data1", ""))
	test.AssertEqualString(t, "", "2", d.GetNamedURLPart("data2", ""))

	test.AssertEqualString(t, "", "ZZ", d.GetNamedURLPart("", "ZZ"))
	test.AssertEqualString(t, "", "ZZ", d.GetNamedURLPart("123", "ZZ"))

	test.AssertEqualString(t, "", "5", d.GetNamedQuery("A"))

	test.AssertEqualInt(t, "", 4, d.GetPartsCount())

	d2 := NewRequestTools(req)
	test.AssertEqualInt(t, "", 4, d2.GetPartsCount())

}
