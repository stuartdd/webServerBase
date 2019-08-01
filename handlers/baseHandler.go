package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"webServerBase/dto"
	"webServerBase/logging"
	"webServerBase/state"
)

/*
HandlerFunctionData is the state of the server
*/
type HandlerFunctionData struct {
	before         vetoHandlerListData
	mappings       map[string]mappingData
	after          vetoHandlerListData
	errorHandler   func(http.ResponseWriter, *http.Request, *dto.Response)
	defaultHandler func(http.ResponseWriter, *http.Request, *dto.Response)
	fileServerList *fileServerContainer
}

type vetoHandlerListData struct {
	handlerFunc func(*http.Request, *dto.Response) *dto.Response
	next        *vetoHandlerListData
}

type fileServerContainer struct {
	path string
	root string
	fs   http.Handler
	next *fileServerContainer
}

type mappingData struct {
	handlerFunc   func(*http.Request) *dto.Response
	requestMethod string
	requestPath   string
}

var logger *logging.LoggerDataReference

/*
NewHandlerData Create new HandlerData object
*/
func NewHandlerData() *HandlerFunctionData {
	logger = logging.NewLogger("BaseHandler")
	handler := &HandlerFunctionData{
		before: vetoHandlerListData{
			handlerFunc: nil,
			next:        nil,
		},
		mappings: make(map[string]mappingData),
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
	}
	return handler
}

/*
SetErrorHandler handle an error response if one occurs
*/
func (p *HandlerFunctionData) SetErrorHandler(errorHandler func(http.ResponseWriter, *http.Request, *dto.Response)) {
	p.errorHandler = errorHandler
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
	container.path = path
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
func (p *HandlerFunctionData) SetHandler(handler func(http.ResponseWriter, *http.Request, *dto.Response)) {
	p.defaultHandler = handler
}

/*
AddMappedHandler creates a route to a function given a path
*/
func (p *HandlerFunctionData) AddMappedHandler(path string, method string, handlerFunc func(*http.Request) *dto.Response) {
	mapping := mappingData{
		handlerFunc:   handlerFunc,
		requestMethod: method,
		requestPath:   path,
	}
	p.mappings[path] = mapping
}

/*
AddBeforeHandler adds a function called before the mapping function
*/
func (p *HandlerFunctionData) AddBeforeHandler(beforeFunc func(*http.Request, *dto.Response) *dto.Response) {
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
func (p *HandlerFunctionData) AddAfterHandler(afterFunc func(*http.Request, *dto.Response) *dto.Response) {
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
ServeHTTP handle ALL calls
*/
func (p *HandlerFunctionData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var mappingResponse *dto.Response
	var url = r.URL.String()

	logRequest(r)

	fileServerMapping := p.fileServerList

	for fileServerMapping.fs != nil {
		if strings.HasPrefix(url, fileServerMapping.path) {
			loggingResponseWriter := NewLoggingResponseWriter(w)
			http.StripPrefix(fileServerMapping.path, fileServerMapping.fs).ServeHTTP(loggingResponseWriter, r)
			if loggingResponseWriter.Is2XX() {
				contentType := getContentType(url)
				if contentType != nil {
					loggingResponseWriter.Header()["Content-Type"] = contentType
				}
			}
			logFileServerFesponse(loggingResponseWriter, fileServerMapping)
			return
		}
		fileServerMapping = fileServerMapping.next
	}

	/*
		Find the mapping
	*/
	mapping, found := p.mappings[url]
	if !found {
		/*
			Mapping was not found
		*/
		error404 := dto.NewResponse(404, http.StatusText(404), "", nil)
		logResponse(error404)
		/*
			delegate to the current error handler to manage the error
		*/
		p.errorHandler(w, r, error404)
		return
	}

	if logger.IsDebug() {
		logger.LogDebugf("Request mapping found. METHOD:%s PATH:%s", mapping.requestMethod, mapping.requestPath)
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
		mappingResponse = mapping.handlerFunc(r)
		if mappingResponse == nil {
			/*
				Bad Request: The server cannot or will not process the request due to an
					apparent client error (e.g., malformed request syntax, size too large,
					invalid request message framing, or deceptive request routing)
			*/
			mappingResponse = dto.NewResponse(400, http.StatusText(400), "", nil)
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
	preProcessResponse(r, mappingResponse)
	logResponse(mappingResponse)
	if mappingResponse.IsNot200() {
		p.errorHandler(w, r, mappingResponse)
	} else {
		p.defaultHandler(w, r, mappingResponse)
	}
}

func preProcessResponse(request *http.Request, response *dto.Response) {
	if response.GetContentType() != "" {
		response.AddHeader("Content-Type", []string{response.GetContentType()})
	}

	response.AddHeader("Server", []string{state.GetStatusDataExecutableName()})

	connection := request.Header["Connection"]
	if len(connection) > 0 {
		response.AddHeader("Connection", connection)
	}
}

func defaultErrorResponseHandler(w http.ResponseWriter, request *http.Request, response *dto.Response) {
	if logger.IsWarn() {
		logger.LogWarn("Using default Error ResponseHandler")
	}
	http.Error(w, response.GetResp(), response.GetCode())
}

func defaultResponseHandler(w http.ResponseWriter, request *http.Request, response *dto.Response) {
	if logger.IsWarn() {
		logger.LogWarn("Using default ResponseHandler")
	}
	for key, value := range response.GetHeaders() {
		w.Header()[key] = value
	}
	if response.GetContentType() != "" {
		w.Header()["Content-Type"] = []string{response.GetContentType()}
	}
	w.Header()["Server"] = []string{state.GetStatusDataExecutableName()}

	w.WriteHeader(response.GetCode())
	fmt.Fprintf(w, response.GetResp())
}

/*
invokeAllHandlersInList
Invoke ALL handlers in the list UNTIL a handler returns a response.
Any response is considered an ERROR.
*/
func invokeAllVetoHandlersInList(p *HandlerFunctionData, w http.ResponseWriter, r *http.Request, response *dto.Response, list *vetoHandlerListData) *dto.Response {
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

func logResponse(response *dto.Response) {
	if logger.IsAccess() {
		logger.LogAccessf("<<< STATUS=%d: RESP=%s", response.GetCode(), response.GetResp())
		logHeaderMap(response.GetHeaders(), "<-<")
	}
}

func logFileServerFesponse(response *LoggingResponseWriter, fileServerMapping *fileServerContainer) {
	if logger.IsAccess() {
		logger.LogAccessf("<<< STATUS=%d staticPath:%s root:%s", response.GetStatusCode(), fileServerMapping.path, fileServerMapping.root)
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

func getContentType(url string) []string {
	pos := strings.LastIndex(url, ".")
	if pos > 0 {
		ext := url[pos:]
		mapping, found := state.GetConfigDataInstance().ContentTypes[ext]
		if found {
			return []string{mapping, "charset=" + state.GetConfigDataInstance().ContentTypeCharset}
		}
	}
	return nil
}
