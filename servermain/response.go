package servermain

import (
	"fmt"
	"net/http"
	jsonconfig "github.com/stuartdd/tools_jsonconfig"
)

/*
Context contained by the Response
*/
type Context struct {
	writer      *ResponseWriterWrapper
	server 		*ServerInstanceData
}
/*
Response is the defininition of a response
*/
type Response struct {
	code        int
	resp        interface{}
	contentType string
	headers     map[string][]string
	err         error
	context 	*Context
}

/*
IsAnError returns true is the response is NOT a 2xx
*/
func (p *Response) IsAnError() bool {
	return (p.code < 200) || (p.code > 299)
}

/*
SetContentType set the content type. E.G. application/json
*/
func (p *Response) SetContentType(contentType string) {
	p.contentType = contentType
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
GetWrappedServer returns the ServerInstanceData wrapped in the response context
*/
func (p *Response) GetWrappedServer() *ServerInstanceData {
	return p.context.server
}
/*
GetWrappedWriter returns the io.Writer wrapped in the response context
*/
func (p *Response) GetWrappedWriter() *ResponseWriterWrapper {
	return p.context.writer
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
func NewResponse(w *ResponseWriterWrapper, s *ServerInstanceData, statusCode int, response interface{}, contentType string, goErr error) *Response {
	if response == "" {
		response = http.StatusText(statusCode)
	}
	return &Response{
		code:        statusCode,
		resp:        response,
		contentType: contentType,
		headers:     make(map[string][]string),
		err:         goErr,
		context:     &Context{
			writer: w,
			server: s,
		},
	}
}
