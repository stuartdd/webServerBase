package servermain

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/stuartdd/webServerBase/logging"
)

/*
StatusHandler returns the server status as a JSON string
*/
func StatusHandler(request *http.Request, response *Response) {
	response.SetResponse(200, response.GetWrappedServer().GetStatusData(), "application/json")
}

/*
StopServerInstance - Stops the server in N seconds defined by optional URL parameter.
Note that the delay is so the response can be processed and returned to the client (or browser)
/stop
/stop/?
both invoke this function
*/
func StopServerInstance(request *http.Request, response *Response) {
	h := NewRequestHandlerHelper(request, response)
	count, err := strconv.Atoi(h.GetNamedURLPart("stop", "2")) // Optional. Default value 2
	if err != nil {
		ThrowPanic("E", 400, SCParamValidation, "Invalid stop period", err.Error())
	} else {
		response.GetWrappedServer().StopServerLater(count, fmt.Sprintf("Stopped by request. Delay %d seconds", count))
		response.SetResponse(200, response.GetWrappedServer().GetStatusData(), "application/json")
	}
}

/*
DefaultTemplateFileHandler - Response handler for basic template processing
*/
func DefaultTemplateFileHandler(request *http.Request, response *Response) {
	h := NewRequestHandlerHelper(request, response)
	server := h.GetServer()
	name := h.GetNamedURLPart("site", "")
	if server.HasTemplate(name) {
		ww := h.GetResponseWriter()
		contentType := LookupContentType(name)
		if (contentType != "") && (ww.Header()[ContentTypeName] == nil) {
			ww.Header()[ContentTypeName] = []string{contentType + "; charset=" + server.contentTypeCharset}
		}
		m := h.GetQueries()
		server.TemplateWithWriter(ww, name, request, m)
		response.Close()
		if logging.IsAccess() {
			response.GetWrappedServer().GetServerLogger().LogAccessf("<<< STATUS=%d: CODE=%d: RESP-FROM-FILE=%s: TYPE=%s", response.GetCode(), response.GetSubCode(), name, contentType)
			response.GetWrappedServer().LogHeaderMap(response.GetHeaders(), "<-<")
		}
	} else {
		response.SetError404(h.GetURL()+" "+server.ListTemplateNames("|"), SCTemplateNotFound)
	}
}

/*
DefaultStaticFileHandler Read a file from a static file location and return it
*/
func DefaultStaticFileHandler(request *http.Request, response *Response) {
	h := NewRequestHandlerHelper(request, response)
	url := h.GetURL()
	server := h.GetServer()
	/*
		get the file path for the url and the matched url. Panics if not found with 404 so no need to check.
	*/
	container := h.GetStaticPathForURL(url)
	/*
		Forward the response headers in to the wrapped http.ResponseWriter
	*/
	ww := h.GetResponseWriter()
	for name, value := range response.GetHeaders() {
		ww.Header()[name] = value
	}
	/*
		Work out the content type from the file name extension and add it to the wrapped http.ResponseWriter
		Dont overwrite a content type that already exists!
	*/
	contentType := LookupContentType(url)
	if (contentType != "") && (ww.Header()[ContentTypeName] == nil) {
		ww.Header()[ContentTypeName] = []string{contentType + "; charset=" + server.contentTypeCharset}
	}
	/*
		derive the file name from the url and the path in the fileServerList
	*/
	fileShort := url[len(container.URLPrefix):]
	filename := filepath.Join(container.FilePath, fileShort)
	/*
		Implemented in servermain/serverInstanceData.go. This wraps the http.ServeContent to return the file contents.
		Panics 404 if file not found. Panics 500 if file cannot be read
	*/
	ServeContent(ww, request, filename)
	/*
		The file is being written to the response writer.
		Close the response to prevent further writes to the response writer
	*/
	response.Close()
	/*
		Specific ACCESS log entry for data returned from a file. Dont echo the response as this is a stream. Just indicate the file.
		Dont log the full file name as this reveals the server file system structure and can lead to vulnerabilities.
	*/
	if logging.IsAccess() {
		response.GetWrappedServer().GetServerLogger().LogAccessf("<<< STATUS=%d: CODE=%d: RESP-FROM-FILE=%s: TYPE=%s", response.GetCode(), response.GetSubCode(), fileShort, contentType)
		response.GetWrappedServer().LogHeaderMap(response.GetHeaders(), "<-<")
	}
	return
}
