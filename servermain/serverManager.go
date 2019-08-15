package servermain

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

type vetoHandlerListData struct {
	handlerFunc func(*http.Request, *Response) *Response
	next        *vetoHandlerListData
}

/*
ServerInstanceData is the state of the server
*/
type ServerInstanceData struct {
	before             vetoHandlerListData
	after              vetoHandlerListData
	errorHandler       func(*ResponseWriterWrapper, *http.Request, *Response)
	responseHandler    func(*ResponseWriterWrapper, *http.Request, *Response)
	fileServerList     *fileServerContainer
	redirections       map[string]string
	contentTypeCharset string
	contentTypeLookup  map[string]string
	baseHandlerName    string
	server             *http.Server
	serverState        *statusData
	logger             *logging.LoggerDataReference
	panicResponseCode  int
}

type statusData struct {
	unixTime     int64
	startTime    string
	executable   string
	state        string
	panicCounter int
}

type fileServerContainer struct {
	path string
	root string
	fs   http.Handler
	next *fileServerContainer
}

/*
NewServerInstanceData Create new server data object
*/
func NewServerInstanceData(baseHandlerNameIn string, contentTypeCharsetIn string) *ServerInstanceData {

	if contentTypeCharsetIn == "" {
		contentTypeCharsetIn = "utf=8"
	}

	return &ServerInstanceData{
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
		logger:            logging.NewLogger(baseHandlerNameIn),
		panicResponseCode: 500,
	}
}

/*
ListenAndServeOnPort start the server on a specific port
*/
func (p *ServerInstanceData) ListenAndServeOnPort(port int) {
	p.server = &http.Server{Addr: ":" + strconv.Itoa(port)}
	p.server.Handler = p
	err := p.server.ListenAndServe()
	if err != nil {
		p.logger.LogInfo(err.Error())
	} else {
		p.logger.LogInfo("http: Server closed")
	}
}

/*
StopServerLater stop the server after N seconds
*/
func (p *ServerInstanceData) StopServerLater(waitForSeconds int) {
	p.serverState.state = "STOPPING"
	go p.stopServerThraed(waitForSeconds)
}

func (p *ServerInstanceData) stopServerThraed(waitForSeconds int) {
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
func (p *ServerInstanceData) LookupContentType(url string) (string, string) {
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
func (p *ServerInstanceData) AddContentTypeFromMap(mimeTypeMap map[string]string) {
	for name, value := range mimeTypeMap {
		p.contentTypeLookup[name] = value
	}
}

/*
SetErrorHandler handle an error response if one occurs
*/
func (p *ServerInstanceData) SetErrorHandler(errorHandler func(*ResponseWriterWrapper, *http.Request, *Response)) {
	p.errorHandler = errorHandler
}

/*
SetPanicResponseCode handle an error response if one occurs
*/
func (p *ServerInstanceData) SetPanicResponseCode(responseCode int) {
	p.panicResponseCode = responseCode
}

/*
SetRedirections add a map of re-directions
Example in config file: "redirections" : {"/":"/static/index.html"}
*/
func (p *ServerInstanceData) SetRedirections(redirections map[string]string) {
	p.redirections = redirections
}

/*
SetFileServerDataFromMap creates a file servers for each mapping
*/
func (p *ServerInstanceData) SetFileServerDataFromMap(mappings map[string]string) {
	for key, value := range mappings {
		p.AddFileServerData(key, value)
	}
}

/*
AddFileServerData creates a file server for a path and a root directory
*/
func (p *ServerInstanceData) AddFileServerData(path string, root string) {
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
func (p *ServerInstanceData) SetResponseHandler(handler func(*ResponseWriterWrapper, *http.Request, *Response)) {
	p.responseHandler = handler
}

/*
AddMappedHandler creates a route to a function given a path
*/
func (p *ServerInstanceData) AddMappedHandler(path string, method string, handlerFunc func(*http.Request) *Response) {
	AddPathMappingElement(path, method, handlerFunc)
}

/*
AddBeforeHandler adds a function called before the mapping function
*/
func (p *ServerInstanceData) AddBeforeHandler(beforeFunc func(*http.Request, *Response) *Response) {
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
func (p *ServerInstanceData) AddAfterHandler(afterFunc func(*http.Request, *Response) *Response) {
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
func (p *ServerInstanceData) ServeStaticFile(w *ResponseWriterWrapper, r *http.Request, url string) bool {
	fileServerMapping := p.fileServerList
	for fileServerMapping.fs != nil {
		if strings.HasPrefix(url, fileServerMapping.path) {
			contentType, ext := p.LookupContentType(url)
			if contentType != "" {
				p.setContentTypeHeader(w, contentType)
			}
			filename := filepath.Join(fileServerMapping.root, url[len(fileServerMapping.path):])
			http.ServeFile(w, r, filename)
			p.logFileServerResponse(w, fileServerMapping.path, ext, contentType, filename)
			return true
		}
		fileServerMapping = fileServerMapping.next
	}
	return false
}

/*
ServeHTTP handle ALL calls
*/
func (p *ServerInstanceData) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	w := NewResponseWriterWrapper(rw, p)
	defer checkPanicIsThrown(w, r)
	var actualResponse *Response
	var url = r.URL.Path

	p.logRequest(r)
	trans := p.redirections[url]
	if trans != "" {
		if p.logger.IsInfo() {
			p.logger.LogInfof(">>> REDIRECT: %s --> %s", url, trans)
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
			delegate to the current error handler to manage the error
		*/
		p.errorHandler(w, r, NewResponse(404, http.StatusText(404), "", nil))
		return
	}

	/*
		We found a matching function for the request so lets check each before handler to see if we can procceed.
		If a before handler returns a response then we abandon the request.
	*/
	actualResponse = p.invokeAllVetoHandlersInList(r, nil, &p.before)
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
		afterVetoHandlerError := p.invokeAllVetoHandlersInList(r, actualResponse, &p.after)
		if afterVetoHandlerError != nil {
			actualResponse = afterVetoHandlerError
			if p.logger.IsWarn() {
				p.logger.LogWarnf("Response was Vetoed by 'After' handler:%s", actualResponse.GetCSV())
			}
		}
	} else {
		if p.logger.IsWarn() {
			p.logger.LogWarnf("Request was Vetoed by 'Before' handler:%s", actualResponse.GetCSV())
		}
	}

	if actualResponse.IsNot200() {
		p.errorHandler(w, r, actualResponse)
	} else {
		p.responseHandler(w, r, actualResponse)
	}
}

func checkPanicIsThrown(w *ResponseWriterWrapper, r *http.Request) {
	server := w.GetServer()
	rec := recover()
	if rec != nil {
		server.serverState.panicCounter++
		text := fmt.Sprintf("REQUEST:%s MESSAGE:%s", r.URL.Path, rec)
		server.logger.LogErrorWithStackTrace("!!!", text)
		server.errorHandler(w, r, NewResponse(server.panicResponseCode, text, "", nil))
	}
}

/*
GetStatusDataJSON server status as a JSON string
*/
func (p *ServerInstanceData) GetStatusDataJSON() string {
	uptime := time.Now().Unix() - p.serverState.unixTime
	return fmt.Sprintf("{\"state\":\"%s\",\"startTime\":\"%s\",\"executable\":\"%s\",\"uptime\":%d,\"panics\":%d}",
		p.serverState.state,
		p.serverState.startTime,
		p.serverState.executable,
		uptime,
		p.serverState.panicCounter,
	)
}

func defaultErrorResponseHandler(w *ResponseWriterWrapper, request *http.Request, response *Response) {
	w.GetServer().preProcessResponse(request, response)
	w.GetServer().logResponse(response)
	for key, value := range response.GetHeaders() {
		w.Header()[key] = value
	}
	w.Header()[contentTypeName] = []string{w.GetServer().contentTypeLookup["json"]}
	w.WriteHeader(response.GetCode())
	fmt.Fprintf(w, toErrorJSON(response.GetCode(), response.GetResp()))
}

func defaultResponseHandler(w *ResponseWriterWrapper, request *http.Request, response *Response) {
	w.GetServer().preProcessResponse(request, response)
	w.GetServer().logResponse(response)
	for key, value := range response.GetHeaders() {
		w.Header()[key] = value
	}
	w.Header()[contentTypeName] = []string{response.GetContentType()}
	w.WriteHeader(response.GetCode())
	fmt.Fprintf(w, response.GetResp())
}

func (p *ServerInstanceData) preProcessResponse(request *http.Request, response *Response) {
	response.AddHeader("Server", []string{p.baseHandlerName})
	connection := request.Header["Connection"]
	if len(connection) > 0 {
		response.AddHeader("Connection", connection)
	}
}

/*
invokeAllHandlersInList
Invoke ALL handlers in the list UNTIL a handler returns a response.
Any response is considered an ERROR.
*/
func (p *ServerInstanceData) invokeAllVetoHandlersInList(r *http.Request, response *Response, list *vetoHandlerListData) *Response {
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

func (p *ServerInstanceData) logResponse(response *Response) {
	if p.logger.IsAccess() {
		p.logger.LogAccessf("<<< STATUS=%d: RESP=%s", response.GetCode(), response.GetResp())
		p.logHeaderMap(response.GetHeaders(), "<-<")
	}
}

func (p *ServerInstanceData) logFileServerResponse(response *ResponseWriterWrapper, path string, ext string, mime string, fileName string) {
	if p.logger.IsAccess() {
		p.logger.LogAccessf("<<< STATUS=%d staticPath:%s ext:%s content-type:%s file:%s", response.GetStatusCode(), path, ext, mime, fileName)
		p.logHeaderMap(response.Header(), "<-<")
	}
}

func (p *ServerInstanceData) logRequest(r *http.Request) {
	if p.logger.IsAccess() {
		p.logger.LogAccessf(">>> METHOD=%s: REQUEST=%s", r.Method, r.URL.Path)
		p.logHeaderMap(r.Header, ">->")
	}
}

func (p *ServerInstanceData) logHeaderMap(headers map[string][]string, dir string) {
	if p.logger.IsDebug() {
		for k, v := range headers {
			p.logger.LogDebugf("%s HEADER=%s=%s", dir, k, v)
		}
	}
}

func (p *ServerInstanceData) setContentTypeHeader(w *ResponseWriterWrapper, contentType string) {
	w.Header()[contentTypeName] = []string{contentType + "; charset=" + p.contentTypeCharset}
}

func toErrorJSON(code int, desc string) string {
	return fmt.Sprintf("{\"Code\":%d, \"Message\":\"%s\"}", code, desc)
}
