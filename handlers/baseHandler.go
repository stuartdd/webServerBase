package handlers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	baseHandlerName    string
	server             *http.Server
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
func NewHandlerData(baseHandlerNameIn string, contentTypeCharsetIn string) *HandlerFunctionData {

	logger = logging.NewLogger(baseHandlerNameIn)

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
		contentTypeLookup:  getContentTypesMap(),
		baseHandlerName:    baseHandlerNameIn,
	}
}

/*
ListenAndServeOnPort start the server on a specific port
*/
func (p *HandlerFunctionData) ListenAndServeOnPort(port int) {
	p.server = &http.Server{Addr: ":" + strconv.Itoa(port)}
	p.server.Handler = p
	err := p.server.ListenAndServe()
	if err != nil {
		logger.LogInfo(err.Error())
	} else {
		logger.LogInfo("http: Server closed")
	}
}

/*
StopServer stop the server immediatly or after 500 ms
*/
func (p *HandlerFunctionData) StopServer(immediate bool) {
	if !immediate {
		time.Sleep(time.Millisecond * 500)
	}
	err := p.server.Shutdown(context.TODO())
	if err != nil {
		panic(err)
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
SetRedirections add a map of re-directions
Example in config file: "redirections" : {"/":"/static/index.html"}
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
			p.logFileServerResponse(responseWriterWrapper, fileServerMapping.path, ext, contentType, filename)
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

	p.logRequest(r)
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
		p.logResponse(error404)
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
	mappingResponse = p.invokeAllVetoHandlersInList(w, r, nil, &p.before)
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
		afterVetoHandlerError := p.invokeAllVetoHandlersInList(w, r, mappingResponse, &p.after)
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
	p.logResponse(mappingResponse)
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

	response.AddHeader("Server", []string{p.baseHandlerName})

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
func (p *HandlerFunctionData) invokeAllVetoHandlersInList(w http.ResponseWriter, r *http.Request, response *Response, list *vetoHandlerListData) *Response {
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

func (p *HandlerFunctionData) logResponse(response *Response) {
	if logger.IsAccess() {
		logger.LogAccessf("<<< STATUS=%d: RESP=%s", response.GetCode(), response.GetResp())
		p.logHeaderMap(response.GetHeaders(), "<-<")
	}
}

func (p *HandlerFunctionData) logFileServerResponse(response *ResponseWriterWrapper, path string, ext string, mime string, fileName string) {
	if logger.IsAccess() {
		logger.LogAccessf("<<< STATUS=%d staticPath:%s ext:%s content-type:%s file:%s", response.GetStatusCode(), path, ext, mime, fileName)
		p.logHeaderMap(response.Header(), "<-<")
	}
}

func (p *HandlerFunctionData) logRequest(r *http.Request) {
	if logger.IsAccess() {
		logger.LogAccessf(">>> METHOD=%s: REQUEST=%s", r.Method, r.URL.Path)
		p.logHeaderMap(r.Header, ">->")
	}
}

func (p *HandlerFunctionData) logHeaderMap(headers map[string][]string, dir string) {
	if logger.IsDebug() {
		for k, v := range headers {
			logger.LogDebugf("%s HEADER=%s=%s", dir, k, v)
		}
	}
}
