package servermain

import (
	"net/url"
	"net/http"
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
