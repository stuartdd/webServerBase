package handlers

import "net/http"

/*
LoggingResponseWriter replaces http.ResponseWriter
*/
type LoggingResponseWriter struct {
	responseWriter http.ResponseWriter
	statusCode     int
}

/*
GetStatusCode return the response status code.
*/
func (p *LoggingResponseWriter) GetStatusCode() int {
	return p.statusCode
}

/*
IsNot200 returns true is the response is NOT a 2xx
*/
func (p *LoggingResponseWriter) Is2XX() bool {
	return (p.statusCode > 199) && (p.statusCode < 300)
}

/*
NewLoggingResponseWriter Create a new loggingResponseWriter so we can write throught it!
*/
func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK}
}

/*
WriteHeader replace the WriteHeader method.
 Store the sattus Code.
 Pass it to the ResponseWriter
*/
func (p *LoggingResponseWriter) WriteHeader(code int) {
	p.statusCode = code
	p.responseWriter.WriteHeader(code)
}

/*
Header replace the WriteHeader method.
 Store the sattus Code.
 Pass it to the ResponseWriter
*/
func (p *LoggingResponseWriter) Header() http.Header {
	return p.responseWriter.Header()
}

func (p *LoggingResponseWriter) Write(b []byte) (int, error) {
	return p.responseWriter.Write(b)
}
