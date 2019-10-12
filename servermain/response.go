package servermain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/stuartdd/webServerBase/panicapi"
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
	code         int
	subCode      int
	resp         interface{}
	contentType  string
	errorMessage string
	closed       bool
	txid         string
}

/*
Response is the defininition of a response
*/
type Response struct {
	response *responseState
	context  *responseContext
	headers  map[string][]string
	names    map[string]int
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
	if strings.Contains(p.GetContentType(), LookupContentType("json")) {
		return p.GetRespAsJSON()
	}
	return fmt.Sprintf("%v", p.response.resp)
}

/*
GetRespAsJSON returns the String response as JSON
*/
func (p *Response) GetRespAsJSON() string {
	json, err := json.Marshal(p.response.resp)
	if err != nil {
		panicapi.ThrowError(500, panicapi.SCJSONResponseErr, "Failed to marshal response to JSON", fmt.Sprintf("Marshal response type [%T] failed: %e", p.response.resp, err))
	}
	return string(json)
}

/*
GetErrorMessage returns the go error
*/
func (p *Response) GetErrorMessage() string {
	return p.response.errorMessage
}

/*
GetCode returns the http status code
*/
func (p *Response) GetCode() int {
	return p.response.code
}

/*
GetTransactionID returns the http status code
*/
func (p *Response) GetTransactionID() string {
	return p.response.txid
}

/*
GetSubCode returns the http status code
*/
func (p *Response) GetSubCode() int {
	return p.response.subCode
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
	return fmt.Sprintf("id=%s, status=%d, subCode=%d, errorMessage=%s, resp=%s, contentType=%s", p.response.txid, p.response.code, p.response.subCode, p.response.errorMessage, p.response.resp, p.response.contentType)
}

func (p *Response) toErrorJSON() string {
	return fmt.Sprintf("{\"ID\":\"%s\", \"Status\":%d,\"Code\":%d,\"Message\":\"%s\",\"Error\":\"%s\"}", p.response.txid, p.response.code, p.response.subCode, p.response.resp, p.response.errorMessage)
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
func NewResponse(w *ResponseWriterWrapper, s *ServerInstanceData, txid string) *Response {
	return &Response{
		response: &responseState{
			code:         200,
			subCode:      panicapi.SCSubCodeZero,
			resp:         nil,
			contentType:  "",
			errorMessage: "",
			closed:       false,
			txid:         txid,
		},
		context: &responseContext{
			writer: w,
			server: s,
		},
		headers: make(map[string][]string),
		names:   make(map[string]int),
	}
}

/*
SetError404 create an error response
*/
func (p *Response) SetError404(url string, subCode int) *Response {
	p.response = &responseState{
		code:         404,
		subCode:      subCode,
		resp:         http.StatusText(404),
		errorMessage: url,
		contentType:  p.GetContentType(),
		closed:       false,
		txid:         p.response.txid,
	}
	return p
}

/*
SetErrorResponse create an error response
*/
func (p *Response) SetErrorResponse(statusCode int, subCode int, errorMessage string) *Response {
	p.response = &responseState{
		code:         statusCode,
		subCode:      subCode,
		resp:         http.StatusText(statusCode),
		errorMessage: errorMessage,
		contentType:  p.GetContentType(),
		closed:       false,
		txid:         p.response.txid,
	}
	return p
}

/*
SetResponse set the content type. E.G. application/json
*/
func (p *Response) SetResponse(code int, resp interface{}, contentType string) *Response {
	p.response = &responseState{
		code:         code,
		subCode:      panicapi.SCSubCodeZero,
		resp:         resp,
		errorMessage: "",
		contentType:  contentType,
		closed:       false,
		txid:         p.response.txid,
	}
	return p
}
