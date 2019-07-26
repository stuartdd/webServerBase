package dto

import (
	"errors"
	"fmt"
	"net/http"
)

/*
Response is the defininition of a response
*/
type Response struct {
	code        int
	resp        string
	contentType string
	headers     map[string]string
	err         error
}


/*
IsError returns the String response
*/
func (p *Response) IsError() bool {
	return (p.code < 200) || (p.code > 299)
}

/*
GetType returns the response type as a String
*/
func (p *Response) GetType() string {
	if p.IsError() {
		return "RESP_ERR"
	}
	return "RESP__OK"
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
func (p *Response) GetHeaders() map[string]string {
	return p.headers
}

/*
AddHeader returns the String response
*/
func (p *Response) AddHeader(name string, value string) {
	p.GetHeaders()[name] = value
}

/*
GetResp returns the String response
*/
func (p *Response) GetResp() string {
	return p.resp
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
GetJSON returns JSON string representing this object
*/
func (p *Response) GetJSON() string {
	return fmt.Sprintf("{\"code\":%d,\"contentType\":\"%s\",\"err\":\"%s\",\"resp\":\"%s\"}", p.code, p.resp, p.contentType, p.err)
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
func NewResponse(statusCode int, response string, contentType string, goErr error) *Response {
	if response == "" {
		response = http.StatusText(statusCode)
	}
	if goErr == nil {
		goErr = errors.New("")
	}
	return &Response{
		code:        statusCode,
		resp:        response,
		contentType: contentType,
		headers:     make(map[string]string),
		err:         goErr,
	}
}
