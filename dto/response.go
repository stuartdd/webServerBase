package dto

import (
	"errors"
	"fmt"
	"net/http"
)

/*
ErrorResponse is the defininition of an  error
*/
type Response struct {
	code        int
	resp        string
	contentType string
	err         error
}

/*
GetResp returns the String response
*/
func (p *Response) IsError() bool {
	return (p.code < 200) || (p.code > 299)
}

/*
GetContentType returns the String content type. E.G. application/json
*/
func (p *Response) GetContentType() string {
	return p.contentType
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
	return fmt.Sprintf("{\"code\":%d,\"resp\":\"%s\",\"contentType\":\"%s\",\"err\":\"%s\"}", p.code, p.resp, p.contentType, p.err)
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
		err:         goErr,
	}
}
