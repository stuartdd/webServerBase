package servermain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

/*
RequestHandlerHelper contains details of url parameters

Dont fetch anything until asked (lazy load)
*/
type RequestHandlerHelper struct {
	request       *http.Request
	response      *Response
	url           string
	urlParts      []string
	urlPartsCount int
	queries       url.Values
	uuid          string
}

type jsonWrapper struct {
	Name string
	UUID string
	Data string
}

/*
NewRequestHandlerHelper create a new url details with a url
*/
func NewRequestHandlerHelper(r *http.Request, response *Response) *RequestHandlerHelper {
	return &RequestHandlerHelper{
		request:       r,
		response:      response,
		url:           "",
		urlParts:      nil,
		urlPartsCount: 0,
		queries:       nil,
		uuid:          "",
	}
}

/*
WrapAsJSON wrap the response in JSON. If it fails it will just build the string
*/
func (p *RequestHandlerHelper) WrapAsJSON(name string, data string) string {
	s := jsonWrapper{
		Name: name,
		UUID: p.GetUUID(),
		Data: data,
	}
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("{\"name\":\"%s\", \"data\": \"%s\"}", name, data)
	}
	return string(b)
}

/*
GetServer returns the server instance
*/
func (p *RequestHandlerHelper) GetServer() *ServerInstanceData {
	return p.response.GetWrappedServer()
}

/*
GetResponseWriter returns the server instance
*/
func (p *RequestHandlerHelper) GetResponseWriter() *ResponseWriterWrapper {
	return p.response.GetWrappedWriter()
}

/*
GetStaticFileServerData get the path for a static name
*/
func (p *RequestHandlerHelper) GetStaticFileServerData() *StaticFileServerData {
	return p.GetServer().GetStaticFileServerData()
}

/*
GetStaticPathForName get the path for a static name
*/
func (p *RequestHandlerHelper) GetStaticPathForName(name string) *FileServerContainer {
	return p.GetServer().GetStaticPathForName(name)
}

/*
GetStaticPathForURL get the path for a static url
*/
func (p *RequestHandlerHelper) GetStaticPathForURL(url string) *FileServerContainer {
	return p.GetServer().GetStaticPathForURL(url)
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
func (p *RequestHandlerHelper) GetJSONBodyAsMap() map[string]interface{} {
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
func (p *RequestHandlerHelper) GetJSONBodyAsList() []interface{} {
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
func (p *RequestHandlerHelper) GetBodyString() string {
	return string(p.GetBody())
}

/*
GetBody read the body from the request. This can only be done ONCE!
*/
func (p *RequestHandlerHelper) GetBody() []byte {
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
	if p.url == "" {
		p.url = p.request.URL.Path
	}
	return p.url
}

/*
GetURLPart returns part by index or panics if not found and default is empty
*/
func (p *RequestHandlerHelper) GetURLPart(n int, defaultValue string) string {
	list := p.readParts()
	if (n >= 0) && (n < p.urlPartsCount) {
		return list[n]
	}
	if defaultValue == "" {
		ThrowPanic("E", 400, SCMissingURLParam, fmt.Sprintf("URL parameter '%d' missing", n), fmt.Sprintf("URL parameter at position '%d' returned an empty value.", n))
	}
	return defaultValue
}

/*
GetURLParts returns part by index or panics if not found and default is empty
*/
func (p *RequestHandlerHelper) GetURLParts() []string {
	return p.readParts()
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
	i, ok := p.response.names[name]
	if ok {
		return p.GetURLPart(i, defaultValue)
	}
	if defaultValue == "" {
		ThrowPanic("E", 400, SCMissingURLParam, fmt.Sprintf("URL parameter '%s' missing", name), fmt.Sprintf("URL parameter '%s' returned an empty value.", name))
	}
	return defaultValue
}

/*
GetNamedQuery returns part by name
*/
func (p *RequestHandlerHelper) GetNamedQuery(name string) string {
	return p.readQueries().Get(name)
}

/*
GetUUID returns part by name
*/
func (p *RequestHandlerHelper) GetUUID() string {
	if p.uuid == "" {
		p.uuid = uuid.New().String()
	}
	return p.uuid
}

/*
GetQueries - return the URL Query parameters as a map
*/
func (p *RequestHandlerHelper) GetQueries() map[string]string {
	m := make(map[string]string)
	for name, val := range p.readQueries() {
		m[name] = val[0]
	}
	return m
}

/*
GetMapOfRequestData - return the URL Query parameters, Named paaremeters and URL Positional parameters
*/
func (p *RequestHandlerHelper) GetMapOfRequestData() map[string]string {
	m := p.GetQueries()
	for name, index := range p.response.names {
		val := p.GetURLPart(index, "?")
		if val != "?" {
			m[name] = val
		}
	}
	for index, value := range p.GetURLParts() {
		m["url["+strconv.Itoa(index)+"]"] = value
	}
	m["uuid"] = p.GetUUID()
	return m
}

func (p *RequestHandlerHelper) readParts() []string {
	if p.urlParts == nil {
		url := p.GetURL()
		if strings.HasPrefix(url, "/") {
			url = url[1:]
		}
		p.urlParts = strings.Split(strings.TrimSpace(url), "/")
		p.urlPartsCount = len(p.urlParts)
	}
	return p.urlParts
}

func (p *RequestHandlerHelper) readQueries() url.Values {
	if p.queries == nil {
		p.queries = p.request.URL.Query()
	}
	return p.queries
}
