package servermain

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"webServerBase/logging"
)

const contentTypeName = "Content-Type"

type vetoHandlerListData struct {
	handlerFunc func(*http.Request, *Response)
	next        *vetoHandlerListData
}

/*
ServerInstanceData is the state of the server
*/
type ServerInstanceData struct {
	before             vetoHandlerListData
	after              vetoHandlerListData
	errorHandler       func(*http.Request, *Response)
	responseHandler    func(*http.Request, *Response)
	redirections       map[string]string
	contentTypeCharset string
	contentTypeLookup  map[string]string
	server             *http.Server
	serverState        *statusData
	logger             *logging.LoggerDataReference
	panicStatusCode    int
	fileServerData     *FileServerData
	templates          *Templates
}

type statusData struct {
	unixTime     int64
	startTime    string
	executable   string
	state        string
	panicCounter int
}

/*
NewServerInstanceData Create new server data object
*/
func NewServerInstanceData(baseHandlerNameIn string, contentTypeCharsetIn string) *ServerInstanceData {

	if contentTypeCharsetIn == "" {
		contentTypeCharsetIn = "utf-8"
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

		redirections:       make(map[string]string),
		contentTypeCharset: contentTypeCharsetIn,
		contentTypeLookup:  getContentTypesMap(),
		serverState: &statusData{
			unixTime:     time.Now().Unix(),
			startTime:    time.Now().Format("2006-01-02 15:04:05"),
			executable:   baseHandlerNameIn,
			state:        "RUNNING",
			panicCounter: 0,
		},
		logger:          logging.NewLogger(baseHandlerNameIn),
		panicStatusCode: 500,
		fileServerData:  nil,
		templates:       nil,
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

func buildMapData(r *http.Request, p *ServerInstanceData) interface{} {
	data := make(map[string]interface{})
	return data
}

/*
ServeHTTP handle ALL calls
*/
func (p *ServerInstanceData) ServeHTTP(rw http.ResponseWriter, httpRequest *http.Request) {
	url := httpRequest.URL.Path
	/*
		Wrap the http.ResponseWriter so we can check statusCode. We also pass in a ref to the server
		so that the handlers can access the server data (ServerInstanceData)
	*/
	w := NewResponseWriterWrapper(rw)
	/*
		Log the request.
		Define ACCESS logging to see the request in the logs
		Define DEBUG and ACCESS to see the request and headers in the logs
	*/
	p.logRequest(httpRequest)
	/*
		Check for a matching url in the redirections map and redirect if found
	*/
	redirect := p.redirections[url]
	if redirect != "" {
		if p.logger.IsInfo() {
			p.logger.LogInfof(">>> REDIRECT: %s --> %s", url, redirect)
		}
		http.Redirect(w, httpRequest, redirect, http.StatusSeeOther)
		return
	}
	/*
		Create the response object so we can pass it to the handlers
	*/
	actualResponse := NewResponse(w, p)
	/*
		If a panic is thrown by ANY handler this defered method will clean up and LOG the event correctly.
	*/
	defer checkForPanicAndRecover(httpRequest, actualResponse)
	/*
		Find the mapping for the url (ReST style)
	*/
	mapping, found := GetPathMappingElement(url, httpRequest.Method)
	if !found {
		/*
			Mapping not found,
			delegate to the current error handler to manage the error
		*/
		p.errorHandler(httpRequest, actualResponse.SetError404(url))
		return
	}

	/*
		We found a matching function for the request so lets check each before handler to see if we can procceed.
		If a before handler changes the response to an error then we abandon the request and return it's response.
	*/
	p.invokeAllVetoHandlersInList(httpRequest, actualResponse, &p.before)
	if actualResponse.IsAnError() {
		if p.logger.IsWarn() {
			p.logger.LogWarnf("Request was Vetoed by 'Before' handler:%s", actualResponse.GetCSV())
		}
	} else {
		/*
			We found a matching function for the request so lets get the response.
			Do not return it immediatly as the after handlers may want to veto the response!
		*/
		mapping.HandlerFunc(httpRequest, actualResponse)
		/*
			If the handler changes the response to an error then we return it's response
			Otherwisw we see if an after handler wants to veto
		*/
		if actualResponse.IsNotAnError() {
			p.invokeAllVetoHandlersInList(httpRequest, actualResponse, &p.after)
			if actualResponse.IsAnError() {
				if p.logger.IsWarn() {
					p.logger.LogWarnf("Response was Vetoed by 'After' handler:%s", actualResponse.GetCSV())
				}
			}
		}
	}
	/*
		If the data is already sent there is nothing to do.
	*/
	if actualResponse.IsClosed() {
		return
	}
	/*
		If the response is not a 2xx status code then this is an error
	*/
	if actualResponse.IsAnError() {
		p.errorHandler(httpRequest, actualResponse)
	} else {
		p.responseHandler(httpRequest, actualResponse)
	}
}

/*
StopServerLater stop the server after N seconds
*/
func (p *ServerInstanceData) StopServerLater(waitForSeconds int) {
	p.serverState.state = "STOPPING"
	go p.stopServerThraed(waitForSeconds)
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
SetStaticFileServerData handle an error response if one occurs
*/
func (p *ServerInstanceData) SetStaticFileServerData(fileServerDataMap map[string]string) {
	p.fileServerData = NewStaticFileServer(fileServerDataMap)
}

/*
SetPathToTemplates initialise the template system
*/
func (p *ServerInstanceData) SetPathToTemplates(pathToTemplates string) {
	templ, err := LoadTemplates("../site")
	if err != nil {
		panic(err)
	}
	if templ.HasAnyTemplates() {
		p.templates = templ
	}
	panic("SetPathToTemplates did NOT ")
}

/*
SetErrorHandler handle an error response if one occurs
*/
func (p *ServerInstanceData) SetErrorHandler(errorHandler func(*http.Request, *Response)) {
	p.errorHandler = errorHandler
}

/*
SetResponseHandler handle a NON error response
*/
func (p *ServerInstanceData) SetResponseHandler(handler func(*http.Request, *Response)) {
	p.responseHandler = handler
}

/*
SetPanicStatusCode handle an error response if one occurs
*/
func (p *ServerInstanceData) SetPanicStatusCode(statusCode int) {
	p.panicStatusCode = statusCode
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

/*
SetRedirections add a map of re-directions
Example in config file: "redirections" : {"/":"/static/index.html"}
*/
func (p *ServerInstanceData) SetRedirections(redirections map[string]string) {
	p.redirections = redirections
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
AddMappedHandler creates a route to a function given a path
*/
func (p *ServerInstanceData) AddMappedHandler(path string, method string, handlerFunc func(*http.Request, *Response)) {
	AddPathMappingElement(path, method, handlerFunc)
}

/*
AddBeforeHandler adds a function called before the mapping function
*/
func (p *ServerInstanceData) AddBeforeHandler(beforeFunc func(*http.Request, *Response)) {
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
func (p *ServerInstanceData) AddAfterHandler(afterFunc func(*http.Request, *Response)) {
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
PreProcessResponse updates the http.ResponseWriter (aka ResponseWriterWrapper) using data from the Response

This is public to allow errorResponseHandlers and responsehandlers to make use of the standard processing

No logging is done here so the handler can define the logging strategy. Use Log
*/
func (p *ServerInstanceData) PreProcessResponse(request *http.Request, response *Response) {
	/*
		Add server id to the headers
	*/
	response.AddHeader("Server", []string{p.serverState.executable})
	/*
		Reflect connection keep-alive. Note - request.Header["Connection"] returns sn array!
	*/
	connection := request.Header["Connection"]
	if len(connection) > 0 {
		response.AddHeader("Connection", connection)
	}
	/*
		If a content type is defined in the response then add content-type to the headers.
	*/
	if response.GetContentType() != "" {
		response.GetHeaders()[contentTypeName] = []string{response.GetContentType() + "; charset=" + p.contentTypeCharset}
	}
	/*
		Push all of the headers in the response in to the http.ResponseWriter
	*/
	for key, value := range response.GetHeaders() {
		response.GetWrappedWriter().Header()[key] = value
	}
	/*
		Set the return code in http.ResponseWriter
	*/
	response.GetWrappedWriter().WriteHeader(response.GetCode())
}

/*
LogResponse logs the response and also call logHeaderMap
Define ACCESS logging to see the response in the logs
Define DEBUG and ACCESS to see the response and headers in the logs
*/
func (p *ServerInstanceData) LogResponse(response *Response) {
	if p.logger.IsAccess() {
		p.logger.LogAccessf("<<< STATUS=%d: RESP=%s", response.GetCode(), response.GetResp())
		p.logHeaderMap(response.GetHeaders(), "<-<")
	}
}

/*
ServeContent wraps the http.ServeContent. It opens the file first.
If the open fails it returns an error.
After that it delegates to http.ServeContent
*/
func ServeContent(w *ResponseWriterWrapper, r *http.Request, name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	http.ServeContent(w, r, name, time.Now(), file)
	return nil
}

func checkForPanicAndRecover(r *http.Request, response *Response) {
	server := response.GetWrappedServer()
	rec := recover()
	if rec != nil {
		server.serverState.panicCounter++
		text := fmt.Sprintf("REQUEST:%s MESSAGE:%s", r.URL.Path, rec)
		server.logger.LogErrorWithStackTrace("!!!", text)
		server.errorHandler(r, response.ChangeResponse(server.panicStatusCode, text, "", nil))
	}
}

func defaultTemplateFileHandler(w *ResponseWriterWrapper, r *http.Request) (bool, error) {
	return false, nil
}

func defaultStaticFileHandler(w *ResponseWriterWrapper, r *http.Request) (bool, error) {
	return false, nil
}

func defaultErrorResponseHandler(request *http.Request, response *Response) {
	server := response.GetWrappedServer()
	response.SetContentType(server.contentTypeLookup["json"])
	server.PreProcessResponse(request, response)
	server.LogResponse(response)
	fmt.Fprintf(response.GetWrappedWriter(), toErrorJSON(response.GetCode(), response.GetResp()))
}

func defaultResponseHandler(request *http.Request, response *Response) {
	server := response.GetWrappedServer()
	server.PreProcessResponse(request, response)
	server.LogResponse(response)
	fmt.Fprintf(response.GetWrappedWriter(), response.GetResp())
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
invokeAllHandlersInList
Invoke ALL handlers in the list UNTIL a handler returns a response.
Any response is considered an ERROR and will veto the normal response.
*/
func (p *ServerInstanceData) invokeAllVetoHandlersInList(httpRequest *http.Request, response *Response, list *vetoHandlerListData) {
	for list.next != nil {
		if list.handlerFunc != nil {
			list.handlerFunc(httpRequest, response)
			if response.IsAnError() {
				return
			}
		}
		list = list.next
	}
}

func (p *ServerInstanceData) logFileServerResponse(response *ResponseWriterWrapper, path string, ext string, mime string, fileName string) {
	if p.logger.IsAccess() {
		p.logger.LogAccessf("<<< STATUS=%d staticPath:%s ext:%s content-type:%s file:%s", response.GetStatusCode(), path, ext, mime, fileName)
		p.logHeaderMap(response.Header(), "<-<")
	}
}

/*
	logRequest - Log the request.
	Define ACCESS logging to see the request in the logs
	Define DEBUG and ACCESS to see the request and headers in the logs
*/
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

func toErrorJSON(code int, desc string) string {
	return fmt.Sprintf("{\"Code\":%d, \"Message\":\"%s\"}", code, desc)
}
