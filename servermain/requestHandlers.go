package servermain

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/stuartdd/webServerBase/exec"
	"github.com/stuartdd/webServerBase/logging"
	"github.com/stuartdd/webServerBase/panicapi"
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
	count, err := strconv.Atoi(h.GetNamedURLPart("seconds", "2")) // Optional. Default value 2
	if err != nil {
		panicapi.ThrowError(400, panicapi.SCParamValidation, "Invalid stop period", err.Error())
	} else {
		response.GetWrappedServer().StopServerLater(count, fmt.Sprintf("Stopped by request. Delay %d seconds", count))
		response.SetResponse(200, response.GetWrappedServer().GetStatusData(), "application/json")
	}
}

/*
DefaultOSScriptHandler - Response handler for basic template processing
*/
func DefaultOSScriptHandler(request *http.Request, response *Response) {
	h := NewRequestHandlerHelper(request, response)
	server := h.GetServer()
	logger := server.GetServerLogger()
	/*
		Get the script name from the URL named parameters
	*/
	scriptName := h.GetNamedURLPart("script", "")
	/*
		Get the script data (command line arguments) for the script name
	*/
	data := server.GetOsScriptsData(scriptName)
	/*
		Run the script with the request data map to resolve substitutions.

		Also package the response in to a JSON message
	*/
	osData := exec.RunAndWait(server.GetOsScriptsPath(), data[0], h.GetMapOfRequestData(), data[1:]...)
	if osData.RetCode == 0 {
		if logging.IsDebug() {
			logger.LogDebugf("OS Script %s Executed OK", scriptName)
		}
		contentType := LookupContentType("json")
		response.SetResponse(200, h.WrapAsJSON(scriptName, osData.Stdout), contentType+"; charset="+server.contentTypeCharset)
		return
	}

	if logging.IsError() {
		errText := ""
		if osData.Err != nil {
			errText = osData.Err.Error()
		}
		logger.LogErrorf("OS Script %s Failed. RC:%d. stderr[%s] error[%s]", scriptName, osData.RetCode, osData.Stderr, errText)
	}
	response.SetErrorResponse(417, panicapi.SCScriptError, "Expectation Failed")
}

/*
DefaultTemplateFileHandler - Response handler for basic template processing
*/
func DefaultTemplateFileHandler(request *http.Request, response *Response) {
	h := NewRequestHandlerHelper(request, response)
	server := h.GetServer()
	name := h.GetNamedURLPart("template", "")
	if server.HasTemplate(name) {
		ww := h.GetResponseWriter()
		contentType := LookupContentType(name)
		if (contentType != "") && (ww.Header()[ContentTypeName] == nil) {
			ww.Header()[ContentTypeName] = []string{contentType + "; charset=" + server.contentTypeCharset}
		}
		server.TemplateWithWriter(ww, name, request, h.GetMapOfRequestData())
		response.Close()
		if logging.IsAccess() {
			response.GetWrappedServer().GetServerLogger().LogAccessf("<<< STATUS=%d: CODE=%d: RESP-FROM-FILE=%s: TYPE=%s", response.GetCode(), response.GetSubCode(), name, contentType)
			response.GetWrappedServer().LogHeaderMap(response.GetHeaders(), "<-<")
		}
	} else {
		response.SetError404(h.GetURL()+" "+server.ListTemplateNames("|"), panicapi.SCTemplateNotFound)
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
	pathData := h.GetStaticPathForURL(url)
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
	fileShort := url[len(pathData.URLPrefix):]
	filename := filepath.Join(pathData.FilePath, fileShort)
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
