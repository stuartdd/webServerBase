package servermain

import "net/http"

/*
ResponseWriterWrapper replaces http.ResponseWriter

Methods are inherited! from http.ResponseWriter.
This allows us to pass ResponseWriterWrapper as a http.ResponseWriter to
any methods expecting an object with the http.ResponseWriter interface
*/
type ResponseWriterWrapper struct {
	responseWriter http.ResponseWriter
	statusCode     int
}

/*
GetStatusCode return the response status code.
*/
func (p *ResponseWriterWrapper) GetStatusCode() int {
	return p.statusCode
}

/*
NewResponseWriterWrapper Create a new ResponseWriterWrapper so we can write throught it!
*/
func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		responseWriter: w,
		statusCode:     http.StatusOK,
	}
}

/*
WriteHeader delegates to http.ResponseWriter.WriteHeader method.
Additional behaviour is to Store the status Code before passing it on.
*/
func (p *ResponseWriterWrapper) WriteHeader(code int) {
	p.statusCode = code
	p.responseWriter.WriteHeader(code)
}

/*
Header delegates to http.ResponseWriter.Header method.
*/
func (p *ResponseWriterWrapper) Header() http.Header {
	return p.responseWriter.Header()
}

/*
Write delegates to http.ResponseWriter.Write method.
*/
func (p *ResponseWriterWrapper) Write(b []byte) (n int, err error) {
	return p.responseWriter.Write(b)
}
