package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"webServerBase/logging"
)

/*
HandlerFunctionData is the state of the server
*/
type HandlerFunctionData struct {
	before             vetoHandlerListData
	after              vetoHandlerListData
	errorHandler       func(http.ResponseWriter, *http.Request, *Response)
	defaultHandler     func(http.ResponseWriter, *http.Request, *Response)
	fileServerList     *fileServerContainer
	redirections       map[string]string
	contentTypeCharset string
	contentTypeLookup  map[string]string
	handlerModuleName  string
}

type vetoHandlerListData struct {
	handlerFunc func(*http.Request, *Response) *Response
	next        *vetoHandlerListData
}

type fileServerContainer struct {
	path string
	root string
	fs   http.Handler
	next *fileServerContainer
}

var logger *logging.LoggerDataReference

/*
NewHandlerData Create new HandlerData object
*/
func NewHandlerData(contentTypeCharsetIn string, moduleName string) *HandlerFunctionData {

	logger = logging.NewLogger(moduleName)

	if contentTypeCharsetIn == "" {
		contentTypeCharsetIn = "utf=8"
	}

	return &HandlerFunctionData{
		before: vetoHandlerListData{
			handlerFunc: nil,
			next:        nil,
		},
		after: vetoHandlerListData{
			handlerFunc: nil,
			next:        nil,
		},
		errorHandler:   defaultErrorResponseHandler,
		defaultHandler: defaultResponseHandler,
		fileServerList: &fileServerContainer{
			path: "",
			root: "",
			fs:   nil,
			next: nil,
		},
		redirections:       make(map[string]string),
		contentTypeCharset: contentTypeCharsetIn,
		contentTypeLookup:  populateContentTypes(),
		handlerModuleName:  moduleName,
	}
}

/*
LookupContentType for a given url return the content type based on the .ext
*/
func (p *HandlerFunctionData) LookupContentType(url string) (string, string) {
	pos := strings.LastIndex(url, ".")
	if pos > 0 {
		ext := url[pos+1:]
		mapping, found := p.contentTypeLookup[ext]
		if found {
			return mapping, ext
		}
	}
	return "", ""
}

/*
AddContentTypeFromMap add to or update the contentType Map
*/
func (p *HandlerFunctionData) AddContentTypeFromMap(mimeTypeMap map[string]string) {
	for name, value := range mimeTypeMap {
		p.contentTypeLookup[name] = value
	}
}

/*
SetErrorHandler handle an error response if one occurs
*/
func (p *HandlerFunctionData) SetErrorHandler(errorHandler func(http.ResponseWriter, *http.Request, *Response)) {
	p.errorHandler = errorHandler
}

/*
SetErrorHandler handle an error response if one occurs
*/
func (p *HandlerFunctionData) SetRedirections(redirections map[string]string) {
	p.redirections = redirections
}

/*
AddFileServerDataFromMap creates a file servers for each mapping
*/
func (p *HandlerFunctionData) AddFileServerDataFromMap(mappings map[string]string) {
	for key, value := range mappings {
		p.AddFileServerData(key, value)
	}
}

/*
AddFileServerData creates a file server for a path and a root directory
*/
func (p *HandlerFunctionData) AddFileServerData(path string, root string) {
	container := p.fileServerList
	for container.next != nil {
		container = container.next
	}
	container.path = "/" + path + "/"
	container.root = root
	container.fs = http.FileServer(http.Dir(root))
	container.next = &fileServerContainer{
		path: "",
		root: "",
		fs:   nil,
		next: nil,
	}
}

/*
SetHandler handle a NON error response
*/
func (p *HandlerFunctionData) SetHandler(handler func(http.ResponseWriter, *http.Request, *Response)) {
	p.defaultHandler = handler
}

/*
AddMappedHandler creates a route to a function given a path
*/
func (p *HandlerFunctionData) AddMappedHandler(path string, method string, handlerFunc func(*http.Request) *Response) {
	AddPathMappingElement(path, method, handlerFunc)
}

/*
AddBeforeHandler adds a function called before the mapping function
*/
func (p *HandlerFunctionData) AddBeforeHandler(beforeFunc func(*http.Request, *Response) *Response) {
	bef := &p.before
	for bef.next != nil {
		bef = bef.next
	}
	bef.handlerFunc = beforeFunc
	bef.next = &vetoHandlerListData{
		handlerFunc: nil,
		next:        nil,
	}
}

/*
AddAfterHandler adds a function called after the mapping function
*/
func (p *HandlerFunctionData) AddAfterHandler(afterFunc func(*http.Request, *Response) *Response) {
	aft := &p.after
	for aft.next != nil {
		aft = aft.next
	}
	aft.handlerFunc = afterFunc
	aft.next = &vetoHandlerListData{
		handlerFunc: nil,
		next:        nil,
	}
}

/*
ServeStaticFile Read a file from a static file location and return it
*/
func (p *HandlerFunctionData) ServeStaticFile(w http.ResponseWriter, r *http.Request, url string) bool {
	fileServerMapping := p.fileServerList
	for fileServerMapping.fs != nil {
		if strings.HasPrefix(url, fileServerMapping.path) {
			contentType, ext := p.LookupContentType(url)
			if contentType != "" {
				w.Header()["Content-Type"] = []string{contentType + "; charset=" + p.contentTypeCharset}
			}
			filename := filepath.Join(fileServerMapping.root, url[len(fileServerMapping.path):])
			responseWriterWrapper := NewResponseWriterWrapper(w)
			http.ServeFile(responseWriterWrapper, r, filename)
			logFileServerResponse(responseWriterWrapper, fileServerMapping.path, ext, contentType, filename)
			return true
		}
		fileServerMapping = fileServerMapping.next
	}
	return false
}

/*
ServeHTTP handle ALL calls
*/
func (p *HandlerFunctionData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var mappingResponse *Response
	var url = r.URL.Path

	logRequest(r)
	trans := p.redirections[url]
	if trans != "" {
		if logger.IsInfo() {
			logger.LogInfof(">>> REDIRECT: %s --> %s", url, trans)
		}
		http.Redirect(w, r, trans, http.StatusSeeOther)
		return
	}
	/*
		Find the mapping
	*/
	mapping, found := GetPathMappingElement(url, r.Method)
	if !found {
		if p.ServeStaticFile(w, r, url) {
			return
		}
		/*
			Mapping was not found
		*/
		error404 := NewResponse(404, http.StatusText(404), "", nil)
		logResponse(error404)
		/*
			delegate to the current error handler to manage the error
		*/
		p.errorHandler(w, r, error404)
		return
	}

	if logger.IsDebug() {
		logger.LogDebugf("Request mapping found. METHOD:%s URL:%s", mapping.RequestMethod, url)
	}

	/*
		We found a matching function for the request so lets check each before handler to see if we can procceed.
		If a before handler returns a response then we abandon the request.
	*/
	mappingResponse = invokeAllVetoHandlersInList(p, w, r, nil, &p.before)
	if mappingResponse == nil {
		/*
			We found a matching function for the request so lets get the response.
			Do not return it immediatly as the after handlers may want to veto the response!
		*/
		mappingResponse = mapping.HandlerFunc(r)
		if mappingResponse == nil {
			/*
				Bad Request: The server cannot or will not process the request due to an
					apparent client error (e.g., malformed request syntax, size too large,
					invalid request message framing, or deceptive request routing)
			*/
			mappingResponse = NewResponse(400, http.StatusText(400), "", nil)
		}

		/*
			If an after handler returns a response then we abandon the request AND the response even if it is valid.
		*/
		afterVetoHandlerError := invokeAllVetoHandlersInList(p, w, r, mappingResponse, &p.after)
		if afterVetoHandlerError != nil {
			mappingResponse = afterVetoHandlerError
			if logger.IsWarn() {
				logger.LogWarnf("Response was Vetoed by 'After' handler:%s", mappingResponse.GetCSV())
			}
		}
	} else {
		if logger.IsWarn() {
			logger.LogWarnf("Request was Vetoed by 'Before' handler:%s", mappingResponse.GetCSV())
		}
	}
	/*
		None pof the 'after' handlers vetoed the response so return it!
	*/
	p.preProcessResponse(r, mappingResponse)
	logResponse(mappingResponse)
	if mappingResponse.IsNot200() {
		p.errorHandler(w, r, mappingResponse)
	} else {
		p.defaultHandler(w, r, mappingResponse)
	}
}

func (p *HandlerFunctionData) preProcessResponse(request *http.Request, response *Response) {
	if response.GetContentType() != "" {
		response.AddHeader("Content-Type", []string{response.GetContentType()})
	}

	response.AddHeader("Server", []string{p.handlerModuleName})

	connection := request.Header["Connection"]
	if len(connection) > 0 {
		response.AddHeader("Connection", connection)
	}
}

func defaultErrorResponseHandler(w http.ResponseWriter, request *http.Request, response *Response) {
	if logger.IsWarn() {
		logger.LogWarn("Using default Error ResponseHandler")
	}
	http.Error(w, response.GetResp(), response.GetCode())
}

func defaultResponseHandler(w http.ResponseWriter, request *http.Request, response *Response) {
	if logger.IsWarn() {
		logger.LogWarn("Using default ResponseHandler")
	}
	for key, value := range response.GetHeaders() {
		w.Header()[key] = value
	}
	if response.GetContentType() != "" {
		w.Header()["Content-Type"] = []string{response.GetContentType()}
	}
	w.WriteHeader(response.GetCode())
	fmt.Fprintf(w, response.GetResp())
}

/*
invokeAllHandlersInList
Invoke ALL handlers in the list UNTIL a handler returns a response.
Any response is considered an ERROR.
*/
func invokeAllVetoHandlersInList(p *HandlerFunctionData, w http.ResponseWriter, r *http.Request, response *Response, list *vetoHandlerListData) *Response {
	for list.next != nil {
		if list.handlerFunc != nil {
			handlerResponse := list.handlerFunc(r, response)
			if handlerResponse != nil {
				return handlerResponse
			}
		}
		list = list.next
	}
	return nil
}

func logResponse(response *Response) {
	if logger.IsAccess() {
		logger.LogAccessf("<<< STATUS=%d: RESP=%s", response.GetCode(), response.GetResp())
		logHeaderMap(response.GetHeaders(), "<-<")
	}
}

func logFileServerResponse(response *ResponseWriterWrapper, path string, ext string, mime string, fileName string) {
	if logger.IsAccess() {
		logger.LogAccessf("<<< STATUS=%d staticPath:%s ext:%s content-type:%s file:%s", response.GetStatusCode(), path, ext, mime, fileName)
		logHeaderMap(response.Header(), "<-<")
	}
}

func logRequest(r *http.Request) {
	if logger.IsAccess() {
		logger.LogAccessf(">>> METHOD=%s: REQUEST=%s", r.Method, r.URL.Path)
		logHeaderMap(r.Header, ">->")
	}
}

func logHeaderMap(headers map[string][]string, dir string) {
	if logger.IsDebug() {
		for k, v := range headers {
			logger.LogDebugf("%s HEADER=%s=%s", dir, k, v)
		}
	}
}

/*
from : https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Complete_list_of_MIME_types
*/
func populateContentTypes() map[string]string {
	mime := make(map[string]string)
	mime["aac"] = "audio/aac"
	mime["abw"] = "application/x-abiword"
	mime["arc"] = "application/x-freearc"
	mime["avi"] = "video/x-msvideo"
	mime["azw"] = "application/vnd.amazon.ebook"
	mime["bin"] = "application/octet-stream"
	mime["bmp"] = "image/bmp"
	mime["bz"] = "application/x-bzip"
	mime["bz2"] = "application/x-bzip2"
	mime["csh"] = "application/x-csh"
	mime["css"] = "text/css"
	mime["csv"] = "text/csv"
	mime["doc"] = "application/msword"
	mime["docx"] = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	mime["eot"] = "application/vnd.ms-fontobject"
	mime["epub"] = "application/epub+zip"
	mime["gif"] = "image/gif"
	mime["htm"] = "text/html"
	mime["html"] = "text/html"
	mime["ico"] = "image/vnd.microsoft.icon" // Some browsers use image/x-icon. Add to config data to override!
	mime["ics"] = "text/calendar"
	mime["jar"] = "application/java-archive"
	mime["jpeg"] = "image/jpeg"
	mime["jpg"] = "image/jpeg"
	mime["js"] = "text/javascript"
	mime["json"] = "application/json"
	mime["jsonld"] = "application/ld+json"
	mime["mid"] = "audio/midi audio/x-midi"
	mime["midi"] = "audio/midi audio/x-midi"
	mime["mjs"] = "text/javascript"
	mime["mp3"] = "audio/mpeg"
	mime["mpeg"] = "video/mpeg"
	mime["mpkg"] = "application/vnd.apple.installer+xml"
	mime["odp"] = "application/vnd.oasis.opendocument.presentation"
	mime["ods"] = "application/vnd.oasis.opendocument.spreadsheet"
	mime["odt"] = "application/vnd.oasis.opendocument.text"
	mime["oga"] = "audio/ogg"
	mime["ogv"] = "video/ogg"
	mime["ogx"] = "application/ogg"
	mime["otf"] = "font/otf"
	mime["png"] = "image/png"
	mime["pdf"] = "application/pdf"
	mime["ppt"] = "application/vnd.ms-powerpoint"
	mime["pptx"] = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	mime["rar"] = "application/x-rar-compressed"
	mime["rtf"] = "application/rtf"
	mime["sh"] = "application/x-sh"
	mime["svg"] = "image/svg+xml"
	mime["swf"] = "application/x-shockwave-flash"
	mime["tar"] = "application/x-tar"
	mime["tif"] = "image/tiff"
	mime["tiff"] = "image/tiff"
	mime["ts"] = "video/mp2t"
	mime["ttf"] = "font/ttf"
	mime["txt"] = "text/plain"
	mime["vsd"] = "application/vnd.visio"
	mime["wav"] = "audio/wav"
	mime["weba"] = "audio/webm"
	mime["webm"] = "video/webm"
	mime["webp"] = "image/webp"
	mime["woff"] = "font/woff"
	mime["woff2"] = "font/woff2"
	mime["xhtml"] = "application/xhtml+xml"
	mime["xls"] = "application/vnd.ms-excel"
	mime["xlsx"] = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	mime["xml"] = "application/xml"
	mime["xul"] = "application/vnd.mozilla.xul+xml"
	mime["zip"] = "application/zip"
	mime["7z"] = "application/x-7z-compressed"
	return mime
}
