package servermain

import (
	"net/url"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strings"
)

/*
URLDetails contains details of url parameters

Dont fetch anything until asked (lazy load)
*/
type URLDetails struct {
	request  *http.Request
	url      string
	urlParts []string
	urlPartsCount int
	queries url.Values
}

/*
NewURLDetails create a new url details with a url
*/
func NewURLDetails(r *http.Request) *URLDetails {
	return &URLDetails{
		request:  r,
		url:      "",
		urlParts: nil,
		urlPartsCount: 0,
		queries: nil,
	}
}

/*
GetJSONBodyAsObject return an object, populated from a known JSON structure. This can only be done ONCE!
Example: (see urlDetails_test.go)
	testStruct := &TestStruct{}
	err = d.GetJSONBodyAsObject(testStruct)
*/
func (p *URLDetails) GetJSONBodyAsObject(configObject interface{}) error {
	jsonBytes, err := p.GetBody()
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, configObject)
	if err != nil {
		return err
	}
	return nil
}
/*
GetJSONBodyAsMap read the body from the request. This can only be done ONCE!
Use this method if the expected Json starts with {
Example: (see urlDetails_test.go)
	aMap, err := d.GetJSONBodyAsMap()
*/
func (p *URLDetails) GetJSONBodyAsMap() (map[string]interface{}, error) {
	jsonBytes, err := p.GetBody()
	if err != nil {
		return nil , err
	}
	var v interface{}
	err = json.Unmarshal(jsonBytes, &v)
	if err != nil {
		return nil, err
	}
	return v.(map[string]interface{}) ,nil
}
/*
GetJSONBodyAsList read the body from the request. This can only be done ONCE!
Use this method if the expected Json starts with [
	aList, err := d.GetJSONBodyAsList()
*/
func (p *URLDetails) GetJSONBodyAsList() ([]interface{}, error) {
	jsonBytes, err := p.GetBody()
	if err != nil {
		return nil , err
	}
	var v interface{}
	err = json.Unmarshal(jsonBytes, &v)
	if err != nil {
		return nil, err
	}
	return v.([]interface{}), nil
}

/*
GetBodyString read the body from the request. This can only be done ONCE!
*/
func (p *URLDetails) GetBodyString() (string, error) {
	bodyBytes, err := p.GetBody()
	if err != nil {
		return "" , err
	}
	return string(bodyBytes), nil
}

/*
GetBody read the body from the request. This can only be done ONCE!
*/
func (p *URLDetails) GetBody() ([]byte, error) {
	bodyBytes, err := ioutil.ReadAll(p.request.Body)
	defer p.request.Body.Close()
	if err != nil {
		return nil , err
	}
	return bodyBytes, nil
}

/*
GetURL returns the URL
*/
func (p *URLDetails) GetURL() string {
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
func (p *URLDetails) GetURLPart(n int, defaultValue string) string {
	list := p.readParts()
	if ((n>=0 ) && (n<p.urlPartsCount)) {
		return list[n]
	}
	return defaultValue
}

/*
GetPartsCount returns the number of parts in the URL
*/
func (p *URLDetails) GetPartsCount() int {
	p.readParts()
	return p.urlPartsCount
}

/*
GetNamedPart returns part by name
*/
func (p *URLDetails) GetNamedPart(name string, defaultValue string) string {
	list := p.readParts()
	for index, val := range list {
		if (val == name) {
			return p.GetURLPart(index+1, defaultValue)
		}
	}
	return defaultValue
}
/*
GetNamedQuery returns part by name
*/
func (p *URLDetails) GetNamedQuery(name string) string {
	return p.readQueries().Get(name)
}

func (p *URLDetails) readParts() []string {
	if (p.urlParts==nil) {
		p.urlParts = strings.Split(strings.TrimSpace(p.GetURL()), "/")
		p.urlPartsCount = len(p.urlParts)
	}
	return p.urlParts
}

func (p *URLDetails) readQueries() url.Values {
	if (p.queries==nil) {
		p.queries = p.request.URL.Query()
	}
	return p.queries
}
