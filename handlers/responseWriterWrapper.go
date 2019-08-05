package handlers

import "net/http"

/*
ResponseWriterWrapper replaces http.ResponseWriter
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
IsNot200 returns true is the response is NOT a 2xx
*/
func (p *ResponseWriterWrapper) Is2XX() bool {
	return (p.statusCode > 199) && (p.statusCode < 300)
}

/*
NewResponseWriterWrapper Create a new ResponseWriterWrapper so we can write throught it!
*/
func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{w, http.StatusOK}
}

/*
WriteHeader replace the WriteHeader method.
 Store the sattus Code.
 Pass it to the ResponseWriter
*/
func (p *ResponseWriterWrapper) WriteHeader(code int) {
	p.statusCode = code
	p.responseWriter.WriteHeader(code)
}

/*
Header replace the WriteHeader method.
 Store the sattus Code.
 Pass it to the ResponseWriter
*/
func (p *ResponseWriterWrapper) Header() http.Header {
	return p.responseWriter.Header()
}

func (p *ResponseWriterWrapper) Write(b []byte) (int, error) {
	return p.responseWriter.Write(b)
}
