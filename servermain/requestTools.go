package servermain

import (
	"fmt"
	"net/url"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strings"
)

/*
RequestTools contains details of url parameters

Dont fetch anything until asked (lazy load)
*/
type RequestTools struct {
	request  *http.Request
	url      string
	urlParts []string
	urlPartsCount int
	queries url.Values
}

/*
NewRequestTools create a new url details with a url
*/
func NewRequestTools(r *http.Request) *RequestTools {
	return &RequestTools{
		request:  r,
		url:      "",
		urlParts: nil,
		urlPartsCount: 0,
		queries: nil,
	}
}

/*
GetJSONBodyAsObject return an object, populated from a known JSON structure. This can only be done ONCE!
Example: (see RequestTools_test.go)
	testStruct := &TestStruct{}
	err = d.GetJSONBodyAsObject(testStruct)
*/
func (p *RequestTools) GetJSONBodyAsObject(configObject interface{}) {
	jsonBytes := p.GetBody()
	err := json.Unmarshal(jsonBytes, configObject)
	if err != nil {
		ThrowPanic("E", 400, SCInvalidJSONRequest, "Invalid JSON in request body", err.Error())
	}
}
/*
GetJSONBodyAsMap read the body from the request. This can only be done ONCE!
Use this method if the expected Json starts with {
Example: (see RequestTools_test.go)
	aMap, err := d.GetJSONBodyAsMap()
*/
func (p *RequestTools) GetJSONBodyAsMap() (map[string]interface{}) {
	jsonBytes := p.GetBody()
	var v interface{}
	err := json.Unmarshal(jsonBytes, &v)
	if err != nil {
		ThrowPanic("E", 400, SCInvalidJSONRequest, "Invalid JSON in request body", err.Error())
	}
	return v.(map[string]interface{})
}
/*
GetJSONBodyAsList read the body from the request. This can only be done ONCE!
Use this method if the expected Json starts with [
	aList, err := d.GetJSONBodyAsList()
*/
func (p *RequestTools) GetJSONBodyAsList() ([]interface{}) {
	jsonBytes := p.GetBody()
	var v interface{}
	err := json.Unmarshal(jsonBytes, &v)
	if err != nil {
		ThrowPanic("E", 400, SCInvalidJSONRequest, "Invalid JSON in request body", err.Error())
	}
	return v.([]interface{})
}

/*
GetBodyString read the body from the request. This can only be done ONCE!
*/
func (p *RequestTools) GetBodyString() (string) {
	return string(p.GetBody())
}

/*
GetBody read the body from the request. This can only be done ONCE!
*/
func (p *RequestTools) GetBody() ([]byte) {
	bodyBytes, err := ioutil.ReadAll(p.request.Body)
	defer p.request.Body.Close()
	if err != nil {
		ThrowPanic("E", 400, SCReadJSONRequest, "Error reading request body", err.Error())
	}
	return bodyBytes
}

/*
GetURL returns the URL
*/
func (p *RequestTools) GetURL() string {
	if (p.url=="") {
		p.url = p.request.URL.Path
		if (strings.HasPrefix(p.url, "/")) {
			p.url = p.request.URL.Path[1:]
		}
	}
	return p.url
}


/*
GetURLPart returns part by index
*/
func (p *RequestTools) GetURLPart(n int, defaultValue string) string {
	list := p.readParts()
	if ((n>=0 ) && (n<p.urlPartsCount)) {
		return list[n]
	}
	if (defaultValue == "") {
		ThrowPanic("E", 400, SCMissingURLParam, fmt.Sprintf("URL parameter '%d' missing", n), fmt.Sprintf("URL parameter at position '%d' returned an empty value.",n))
	}
	return defaultValue
}

/*
GetPartsCount returns the number of parts in the URL
*/
func (p *RequestTools) GetPartsCount() int {
	p.readParts()
	return p.urlPartsCount
}

/*
GetNamedPart returns part by name
*/
func (p *RequestTools) GetNamedURLPart(name string, defaultValue string) string {
	list := p.readParts()
	for index, val := range list {
		if (val == name) {
			return p.GetURLPart(index+1, defaultValue)
		}
	}
	if (defaultValue == "") {
		ThrowPanic("E", 400, SCMissingURLParam, fmt.Sprintf("URL parameter '%s' missing", name), fmt.Sprintf("URL parameter '%s' returned an empty value.",name))
	}
	return defaultValue
}
/*
GetNamedQuery returns part by name
*/
func (p *RequestTools) GetNamedQuery(name string) string {
	return p.readQueries().Get(name)
}

func (p *RequestTools) readParts() []string {
	if (p.urlParts==nil) {
		p.urlParts = strings.Split(strings.TrimSpace(p.GetURL()), "/")
		p.urlPartsCount = len(p.urlParts)
	}
	return p.urlParts
}

func (p *RequestTools) readQueries() url.Values {
	if (p.queries==nil) {
		p.queries = p.request.URL.Query()
	}
	return p.queries
}
