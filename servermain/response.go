package servermain

import (
	"fmt"
	"net/http"

	jsonconfig "github.com/stuartdd/tools_jsonconfig"
)

/*
ResponseContext contained by the Response
*/
type responseContext struct {
	writer *ResponseWriterWrapper
	server *ServerInstanceData
}

/*
ResponseState contains the current response
*/
type responseState struct {
	code        int
	resp        interface{}
	contentType string
	err         error
	closed      bool
}

/*
Response is the defininition of a response
*/
type Response struct {
	response *responseState
	context  *responseContext
	headers  map[string][]string
}

/*
SetContentType set the content type. E.G. application/json
*/
func (p *Response) SetContentType(contentType string) {
	p.response.contentType = contentType
}

/*
AddHeader adds an array/slice of strings to a header
*/
func (p *Response) AddHeader(name string, value []string) {
	p.GetHeaders()[name] = value
}

/*
IsNotAnError returns true is the response is NOT a 2xx
*/
func (p *Response) IsNotAnError() bool {
	return !p.IsAnError()
}

/*
IsAnError returns true is the response is NOT a 2xx
*/
func (p *Response) IsAnError() bool {
	return (p.response.code < 200) || (p.response.code > 299)
}

/*
GetContentType returns the String content type. E.G. application/json
*/
func (p *Response) GetContentType() string {
	return p.response.contentType
}

/*
GetHeaders returns the list of headers
*/
func (p *Response) GetHeaders() map[string][]string {
	return p.headers
}

/*
GetResp returns the String response
*/
func (p *Response) GetResp() string {
	str, ok := p.response.resp.(string)
	if ok {
		return str
	}
	json, err := jsonconfig.StringJson(p.response.resp)
	if err == nil {
		return json
	}
	return fmt.Sprintf("%v", p.response.resp)
}

/*
GetError returns the go error
*/
func (p *Response) GetError() error {
	return p.response.err
}

/*
GetCode returns the http status code
*/
func (p *Response) GetCode() int {
	return p.response.code
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
	return fmt.Sprintf("code=%d, contentType=%s, err=%s, resp=%s", p.response.code, p.response.contentType, p.response.err, p.response.resp)
}

/*
Close - A closed response wil NOT write anything to the response stream. So if
the response is written then close so no further data is sent.
*/
func (p *Response) Close() {
	p.response.closed = true
}

/*
IsClosed - A closed response wil NOT write anything to the response stream. So if
the response is written then close so no further data is sent.
*/
func (p *Response) IsClosed() bool {
	return p.response.closed
}

/*
NewResponse create an error response
*/
func NewResponse(w *ResponseWriterWrapper, s *ServerInstanceData) *Response {
	return &Response{
		response: nil,
		context: &responseContext{
			writer: w,
			server: s,
		},
	}
}

/*
SetError404 create an error response
*/
func (p *Response) SetError404(url string) *Response {
	p.response = &responseState{
		code:        404,
		resp:        url + " - " + http.StatusText(404),
		contentType: "",
		err:         nil,
		closed:      false,
	}
	return p
}

/*
ChangeResponse create an error response
*/
func (p *Response) ChangeResponse(statusCode int, response interface{}, contentType string, goErr error) *Response {
	p.response = &responseState{
		code:        statusCode,
		resp:        response,
		contentType: contentType,
		err:         goErr,
		closed:      false,
	}
	return p
}
