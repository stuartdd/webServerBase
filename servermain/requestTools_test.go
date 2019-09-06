package servermain

import (
	"webServerBase/test"
	"testing"
	"strings"
	"bytes"
	"time"
	"net/http"
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
	req, err := http.NewRequest("GET","http://abc:8080/data1/1/data2/2?A=5",bytes.NewReader(jsonTestData))
	if (err != nil) {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)

	testStruct := &TestStruct{}
	err = d.GetJSONBodyAsObject(testStruct)
	if (err != nil) {
		test.Fail(t, "", "Could not read body!")
	}

	test.AssertEqualString(t, "", "Fruit", testStruct.Name)
	test.AssertEqualInt(t, "", 999, testStruct.ID)
	test.AssertEqualInt(t, "", 3, len(testStruct.Types))
	test.AssertEqualString(t, "", "Apple", testStruct.Types[0])
	test.AssertEqualString(t, "", "Banana", testStruct.Types[1])
	test.AssertEqualString(t, "", "Orange", testStruct.Types[2])
	test.AssertEqualString(t, "", "2018-04-09T23:00:00", testStruct.Created.Format("2006-01-02T15:04:05"))
}


func TestWithBodyJsonList(t *testing.T) {
	req, err := http.NewRequest("GET","http://abc:8080/data1/1/data2/2?A=5",strings.NewReader("[\"TEST\",\"VALUE\"]"))
	if (err != nil) {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	aList, err := d.GetJSONBodyAsList()
	if (err != nil) {
		test.Fail(t, "", "Could not read body!")
	}
	test.AssertInterfaceType(t, "", "[]interface {}", aList)
	test.AssertInterfaceType(t, "", "string", aList[0])
	test.AssertEqualString(t, "", "TEST", aList[0].(string))
	test.AssertEqualString(t, "", "VALUE", aList[1].(string))
}

func TestWithBodyJsonMap(t *testing.T) {
	req, err := http.NewRequest("GET","http://abc:8080/data1/1/data2/2?A=5",strings.NewReader("{\"TEST\":\"VALUE\"}"))
	if (err != nil) {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	aMap, err := d.GetJSONBodyAsMap()
	if (err != nil) {
		test.Fail(t, "", "Could not read body!")
	}
	val := aMap["TEST"]
	test.AssertInterfaceType(t, "", "string", val)
	test.AssertEqualString(t, "", "VALUE", val.(string))
}

func TestWithBodyText(t *testing.T) {
	req, err := http.NewRequest("GET","http://abc:8080/data1/1/data2/2?A=5",strings.NewReader("TEST"))
	if (err != nil) {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	text, err := d.GetBodyString()
	if (err != nil) {
		test.Fail(t, "", "Could not read body!")
	}
	test.AssertEqualString(t, "", text, "TEST")
}

func TestWithUrl(t *testing.T) {
	req, err := http.NewRequest("GET","http://abc:8080/data1/1/data2/2?A=5", nil)
	if (err != nil) {
		test.Fail(t, "", err.Error())
	}
	d := NewRequestTools(req)
	test.AssertEqualString(t, "", "data1/1/data2/2", d.GetURL())
	test.AssertEqualString(t, "", "data1", d.GetURLPart(0, ""))
	test.AssertEqualString(t, "", "1", d.GetURLPart(1, ""))
	test.AssertEqualString(t, "", "data2", d.GetURLPart(2, ""))
	test.AssertEqualString(t, "", "2", d.GetURLPart(3, ""))
	test.AssertEmptyString(t, "", d.GetURLPart(4, ""))
	test.AssertEmptyString(t, "", d.GetURLPart(-1, ""))
	test.AssertEqualString(t, "", "Z", d.GetURLPart(4, "Z"))
	test.AssertEqualString(t, "", "X", d.GetURLPart(-4, "X"))

	test.AssertEqualString(t, "", "1", d.GetNamedPart("data1", ""))
	test.AssertEqualString(t, "", "2", d.GetNamedPart("data2", ""))
	test.AssertEmptyString(t, "", d.GetNamedPart("data3", ""))
	test.AssertEmptyString(t, "", d.GetNamedPart("", ""))

	test.AssertEqualString(t, "", "ZZ", d.GetNamedPart("", "ZZ"))
	test.AssertEqualString(t, "", "ZZ", d.GetNamedPart("123", "ZZ"))

	test.AssertEqualString(t, "", "5", d.GetNamedQuery("A"))
	test.AssertEmptyString(t, "", d.GetNamedQuery("X"))

	test.AssertEqualInt(t, "", 4, d.GetPartsCount())

	d2 := NewRequestTools(req)
	test.AssertEqualInt(t, "", 4, d2.GetPartsCount())

	
}
