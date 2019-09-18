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
RequestHandlerHelper contains details of url parameters

Dont fetch anything until asked (lazy load)
*/
type RequestHandlerHelper struct {
	request  *http.Request
	response  *Response
	url      string
	urlParts []string
	urlPartsCount int
	queries url.Values
}

/*
NewRequestHandlerHelper create a new url details with a url
*/
func NewRequestHandlerHelper(r *http.Request, response *Response) *RequestHandlerHelper {
	return &RequestHandlerHelper{
		request:  r,
		response: response,
		url:      "",
		urlParts: nil,
		urlPartsCount: 0,
		queries: nil,
	}
}

/*
GetServer returns the server instance
*/
func (p *RequestHandlerHelper) GetServer() (*ServerInstanceData) {
	return p.response.GetWrappedServer()
}

/*
GetStaticPathForName get the path for a static name
*/
func (p *RequestHandlerHelper) GetStaticPathForName(name string) string {
	return p.GetServer().fileServerData.GetStaticPathForName(name).FsPath
}

/*
GetStaticPathForURL get the path for a static url
*/
func (p *RequestHandlerHelper) GetStaticPathForURL(url string) string {
	return p.GetServer().fileServerData.GetStaticPathForURL(url).FsPath
}
/*
GetJSONBodyAsObject return an object, populated from a known JSON structure. This can only be done ONCE!
Example: (see RequestTools_test.go)
	testStruct := &TestStruct{}
	err = d.GetJSONBodyAsObject(testStruct)
*/
func (p *RequestHandlerHelper) GetJSONBodyAsObject(configObject interface{}) {
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
func (p *RequestHandlerHelper) GetJSONBodyAsMap() (map[string]interface{}) {
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
func (p *RequestHandlerHelper) GetJSONBodyAsList() ([]interface{}) {
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
func (p *RequestHandlerHelper) GetBodyString() (string) {
	return string(p.GetBody())
}

/*
GetBody read the body from the request. This can only be done ONCE!
*/
func (p *RequestHandlerHelper) GetBody() ([]byte) {
	bodyBytes, err := ioutil.ReadAll(p.request.Body)
	defer p.request.Body.Close()
	if err != nil {
		ThrowPanic("E", 400, SCReadJSONRequest, "Error reading request body", err.Error())
	}
	return bodyBytes
}

/*
GetURL returns the URL (Cached in the in thos tool's instance)
*/
func (p *RequestHandlerHelper) GetURL() string {
	if (p.url=="") {
		p.url = p.request.URL.Path
		if (strings.HasPrefix(p.url, "/")) {
			p.url = p.request.URL.Path[1:]
		}
	}
	return p.url
}

/*
GetURLPart returns part by index or panics if not found and default is empty
*/
func (p *RequestHandlerHelper) GetURLPart(n int, defaultValue string) string {
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
func (p *RequestHandlerHelper) GetPartsCount() int {
	p.readParts()
	return p.urlPartsCount
}

/*
GetNamedURLPart returns part by name or panics if not found and default is empty
*/
func (p *RequestHandlerHelper) GetNamedURLPart(name string, defaultValue string) string {
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
func (p *RequestHandlerHelper) GetNamedQuery(name string) string {
	return p.readQueries().Get(name)
}

func (p *RequestHandlerHelper) readParts() []string {
	if (p.urlParts==nil) {
		p.urlParts = strings.Split(strings.TrimSpace(p.GetURL()), "/")
		p.urlPartsCount = len(p.urlParts)
	}
	return p.urlParts
}

func (p *RequestHandlerHelper) readQueries() url.Values {
	if (p.queries==nil) {
		p.queries = p.request.URL.Query()
	}
	return p.queries
}
