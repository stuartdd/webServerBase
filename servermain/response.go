package servermain

import (
	"fmt"
	"net/http"

	jsonconfig "github.com/stuartdd/tools_jsonconfig"
)

/*
Response is the defininition of a response
*/
type Response struct {
	code        int
	resp        interface{}
	contentType string
	headers     map[string][]string
	err         error
}

/*
IsNot200 returns true is the response is NOT a 2xx
*/
func (p *Response) IsNot200() bool {
	return (p.code < 200) || (p.code > 299)
}

/*
GetContentType returns the String content type. E.G. application/json
*/
func (p *Response) GetContentType() string {
	return p.contentType
}

/*
GetHeaders returns the list of headers
*/
func (p *Response) GetHeaders() map[string][]string {
	return p.headers
}

/*
AddHeader adds an array/slice of strings to a header
*/
func (p *Response) AddHeader(name string, value []string) {
	p.GetHeaders()[name] = value
}

/*
GetResp returns the String response
*/
func (p *Response) GetResp() string {
	str, ok := p.resp.(string)
	if ok {
		return str
	}
	json, err := jsonconfig.StringJson(p.resp)
	if err == nil {
		return json
	}
	return fmt.Sprintf("%v", p.resp)
}

/*
GetError returns the go error
*/
func (p *Response) GetError() error {
	return p.err
}

/*
GetCode returns the http status code
*/
func (p *Response) GetCode() int {
	return p.code
}

/*
GetCSV returns JSON string representing this object
*/
func (p *Response) GetCSV() string {
	return fmt.Sprintf("code=%d, contentType=%s, err=%s, resp=%s", p.code, p.contentType, p.err, p.resp)
}

/*
NewResponse create an error response
*/
func NewResponse(statusCode int, response interface{}, contentType string, goErr error) *Response {
	if response == "" {
		response = http.StatusText(statusCode)
	}
	return &Response{
		code:        statusCode,
		resp:        response,
		contentType: contentType,
		headers:     make(map[string][]string),
		err:         goErr,
	}
}
