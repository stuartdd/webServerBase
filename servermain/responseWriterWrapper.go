package servermain

import "net/http"

/*
ResponseWriterWrapper replaces http.ResponseWriter
*/
type ResponseWriterWrapper struct {
	responseWriter http.ResponseWriter
	server         *ServerInstanceData
	statusCode     int
}

/*
GetStatusCode return the response status code.
*/
func (p *ResponseWriterWrapper) GetStatusCode() int {
	return p.statusCode
}

/*
Is2XX returns true is the response is NOT a 2xx
*/
func (p *ResponseWriterWrapper) Is2XX() bool {
	return (p.statusCode > 199) && (p.statusCode < 300)
}

/*
GetServer returns the serverData instance
*/
func (p *ResponseWriterWrapper) GetServer() *ServerInstanceData {
	return p.server
}

/*
NewResponseWriterWrapper Create a new ResponseWriterWrapper so we can write throught it!
*/
func NewResponseWriterWrapper(w http.ResponseWriter, p *ServerInstanceData) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		responseWriter: w,
		server:         p,
		statusCode:     http.StatusOK,
	}
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

func (p *ResponseWriterWrapper) Write(b []byte) (n int, err error) {
	return p.responseWriter.Write(b)
}
