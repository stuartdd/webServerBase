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

const contentTypeName = "Content-Type"

/*
HandlerFunctionData is the state of the server
*/
type HandlerFunctionData struct {
	before             vetoHandlerListData
	after              vetoHandlerListData
	errorHandler       func(http.ResponseWriter, *http.Request, *Response)
	responseHandler    func(http.ResponseWriter, *http.Request, *Response)
	fileServerList     *fileServerContainer
	redirections       map[string]string
	contentTypeCharset string
	contentTypeLookup  map[string]string
	baseHandlerName    string
	server             *http.Server
	serverState        *statusData
	panicResponseCode  int
}

type statusData struct {
	unixTime     int64
	startTime    string
	executable   string
	state        string
	panicCounter int
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
		errorHandler:    defaultErrorResponseHandler,
		responseHandler: defaultResponseHandler,
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
		serverState: &statusData{
			unixTime:     time.Now().Unix(),
			startTime:    time.Now().Format("2006-01-02 15:04:05"),
			executable:   baseHandlerNameIn,
			state:        "RUNNING",
			panicCounter: 0,
		},
		panicResponseCode: 500,
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
StopServerLater stop the server after N seconds
*/
func (p *HandlerFunctionData) StopServerLater(waitForSeconds int) {
	p.serverState.state = "STOPPING"
	go p.stopServerThraed(waitForSeconds)
}

func (p *HandlerFunctionData) stopServerThraed(waitForSeconds int) {
	if waitForSeconds > 0 {
		time.Sleep(time.Second * time.Duration(waitForSeconds))
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
SetPanicResponseCode handle an error response if one occurs
*/
func (p *HandlerFunctionData) SetPanicResponseCode(responseCode int) {
	p.panicResponseCode = responseCode
}

/*
SetRedirections add a map of re-directions
Example in config file: "redirections" : {"/":"/static/index.html"}
*/
func (p *HandlerFunctionData) SetRedirections(redirections map[string]string) {
	p.redirections = redirections
}

/*
SetFileServerDataFromMap creates a file servers for each mapping
*/
func (p *HandlerFunctionData) SetFileServerDataFromMap(mappings map[string]string) {
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
SetResponseHandler handle a NON error response
*/
func (p *HandlerFunctionData) SetResponseHandler(handler func(http.ResponseWriter, *http.Request, *Response)) {
	p.responseHandler = handler
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
	fileServerMa pping := p.fileServerList
	for fileServerMapping.fs != nil {
		if strings.HasPrefix(url, fileServerMapping.path) {
			contentType, ext := p.LookupContentType(url)
			if contentType != "" {
				p.setContentTypeHeader(w, contentType)
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
	defer checkPanicIsThrown(p, w, r)
	var actualResponse *Response
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
		/*
			delegate to the current error handler to manage the error
		*/
		p.errorHandler(w, r, error404)
		return
	}

	/*
		We found a matching function for the request so lets check each before handler to see if we can procceed.
		If a before handler returns a response then we abandon the request.
	*/
	actualResponse = p.invokeAllVetoHandlersInList(w, r, nil, &p.before)
	if actualResponse == nil {
		/*
			We found a matching function for the request so lets get the response.
			Do not return it immediatly as the after handlers may want to veto the response!
		*/
		actualResponse = mapping.HandlerFunc(r)
		if actualResponse == nil {
			/*
				Bad Request: The server cannot or will not process the request due to an
					apparent client error (e.g., malformed request syntax, size too large,
					invalid request message framing, or deceptive request routing)
			*/
			actualResponse = NewResponse(400, http.StatusText(400), "", nil)
		}

		/*
			If an after handler returns a response then we abandon the request AND the response even if it is valid.
		*/
		afterVetoHandlerError := p.invokeAllVetoHandlersInList(w, r, actualResponse, &p.after)
		if afterVetoHandlerError != nil {
			actualResponse = afterVetoHandlerError
			if logger.IsWarn() {
				logger.LogWarnf("Response was Vetoed by 'After' handler:%s", actualResponse.GetCSV())
			}
		}
	} else {
		if logger.IsWarn() {
			logger.LogWarnf("Request was Vetoed by 'Before' handler:%s", actualResponse.GetCSV())
		}
	}
	/*
		None pof the 'after' handlers vetoed the response so return it!
	*/
	p.preProcessResponse(r, actualResponse)
	p.logResponse(actualResponse)
	if actualResponse.IsNot200() {
		p.errorHandler(w, r, actualResponse)
	} else {
		p.responseHandler(w, r, actualResponse)
	}
}

func checkPanicIsThrown(p *HandlerFunctionData, w http.ResponseWriter, r *http.Request) {
	rec := recover()
	if rec != nil {
		p.serverState.panicCounter++
		logger.LogErrorWithStackTrace("!!!", fmt.Sprintf("REQUEST:%s MESSAGE:%s", r.URL.Path, rec))
		s := fmt.Sprintf("%s", rec)
		j := toErrorJSON(p.panicResponseCode, s)
		if logger.IsAccess() {
			logger.LogAccessf("<<< STATUS=%d: RESP=%s", p.panicResponseCode, j)
		}
		p.setContentTypeHeader(w, p.contentTypeLookup["json"])
		w.WriteHeader(p.panicResponseCode)
		fmt.Fprintf(w, j)
	}
}

/*
GetStatusDataJSON server status as a JSON string
*/
func (p *HandlerFunctionData) GetStatusDataJSON() string {
	uptime := time.Now().Unix() - p.serverState.unixTime
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\",\"uptime\":%d,\"panics\":%d}",
		p.serverState.state,
		p.serverState.startTime,
		p.serverState.executable,
		uptime,
		p.serverState.panicCounter,
	)
}

func (p *HandlerFunctionData) preProcessResponse(request *http.Request, response *Response) {
	if response.GetContentType() != "" {
		response.AddHeader(contentTypeName, []string{response.GetContentType()})
	}

	response.AddHeader("Server", []string{p.baseHandlerName})

	connection := request.Header["Connection"]
	if len(connection) > 0 {
		response.AddHeader("Connection", connection)
	}
}

func defaultErrorResponseHandler(w http.ResponseWriter, r *http.Request, response *Response) {
	j := toErrorJSON(response.GetCode(), response.GetResp())
	if logger.IsAccess() {
		logger.LogAccessf("<<< STATUS=%d: RESP=%s", response.GetCode(), j)
	}
	for key, value := range response.GetHeaders() {
		w.Header()[key] = value
	}
	w.Header()[contentTypeName] = []string{"application/json"}
	w.WriteHeader(response.GetCode())
	fmt.Fprintf(w, j)
}

func defaultResponseHandler(w http.ResponseWriter, request *http.Request, response *Response) {
	if logger.IsWarn() {
		logger.LogWarn("Using default ResponseHandler")
	}
	for key, value := range response.GetHeaders() {
		w.Header()[key] = value
	}
	if response.GetContentType() != "" {
		w.Header()[contentTypeName] = []string{response.GetContentType()}
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

func (p *HandlerFunctionData) setContentTypeHeader(w http.ResponseWriter, contentType string) {
	w.Header()[contentTypeName] = []string{contentType + "; charset=" + p.contentTypeCharset}
}

func toErrorJSON(code int, desc string) string {
	return fmt.Sprintf("{\"Code\":%d, \"Message\":\"%s\"}", code, desc)
}
