package servermain

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/stuartdd/webServerBase/logging"
)

/*
SCSubCodeZero and these constants are used as unique subcodes in error responses
*/
const (
	SCSubCodeZero = iota
	SCPathNotFound
	SCStaticPathNotFound
	SCContentNotFound
	SCContentReadFailed
	SCServerShutDown
	SCInvalidJSONRequest
	SCReadJSONRequest
	SCJSONResponseErr
	SCMissingURLParam
	SCStaticFileInit
	SCTemplateNotFound
	SCTemplateError
	SCRuntimeError
	SCStaticPath
	SCWriteFile
	SCParamValidation
	SCScriptNotFound
	SCScriptError
	SCMax
)

/*
ContentTypeName - so we always get it right!
*/
const ContentTypeName = "Content-Type"

/*
ContentLengthName - so we always get it right!
*/
const ContentLengthName = "Content-Length"

type vetoHandlerListData struct {
	handlerFunc func(*http.Request, *Response)
	next        *vetoHandlerListData
}

/*
ServerInstanceData is the state of the server
*/
type ServerInstanceData struct {
	mappingElements    *MappingElements
	before             vetoHandlerListData
	after              vetoHandlerListData
	errorHandler       func(*http.Request, *Response)
	responseHandler    func(*http.Request, *Response)
	redirections       map[string]string
	contentTypeCharset string
	server             *http.Server
	serverState        *StatusData
	logger             *logging.LoggerDataReference
	panicStatusCode    int
	fileServerData     *StaticFileServerData
	templates          *Templates
	serverReturnCode   int
	serverClosedReason string
	osScriptsPath      string
	osScripts          map[string][]string
}

/*
StatusData the state of the server
*/
type StatusData struct {
	UnixTime   int64
	StartTime  string
	Executable string
	State      string
	Panics     int
	Uptime     int64
}

/*
NewServerInstanceData Create new server data object
*/
func NewServerInstanceData(baseHandlerNameIn string, contentTypeCharsetIn string) *ServerInstanceData {

	if contentTypeCharsetIn == "" {
		contentTypeCharsetIn = "utf-8"
	}

	return &ServerInstanceData{
		mappingElements: NewMappingElements(nil),
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
		serverState: &StatusData{
			UnixTime:   time.Now().Unix(),
			StartTime:  time.Now().Format("2006-01-02 15:04:05"),
			Executable: baseHandlerNameIn,
			State:      "RUNNING",
			Panics:     0,
			Uptime:     0,
		},
		logger:             logging.NewLogger(baseHandlerNameIn),
		panicStatusCode:    500,
		fileServerData:     nil,
		templates:          nil,
		serverReturnCode:   1,
		serverClosedReason: "",
	}
}

/*
ListenAndServeOnPort start the server on a specific port
*/
func (p *ServerInstanceData) ListenAndServeOnPort(port int) {
	p.server = &http.Server{Addr: ":" + strconv.Itoa(port)}
	p.server.Handler = p
	err := p.server.ListenAndServe()
	if p.GetServerClosedReason() != "" {
		p.logger.LogInfof("Server Halted: %s", p.GetServerClosedReason())
		if err != nil {
			p.logger.LogInfof("Server Response: %s", err.Error())
		}
		p.serverReturnCode = 0
	} else {
		if err != nil {
			p.logger.LogErrorWithStackTrace("FAILED: Server terminated. Error: %s", err.Error())
			p.serverReturnCode = 1
		} else {
			p.logger.LogInfo("Server Halted.")
			p.serverReturnCode = 2
		}
	}
}

func buildMapData(r *http.Request, p *ServerInstanceData) interface{} {
	data := make(map[string]interface{})
	return data
}

/***********************************************************************************************
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
		Create the response object so we can pass it to the handlers
	*/
	actualResponse := NewResponse(w, p)
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
		s := httpRequest.URL.RawQuery
		if s != "" {
			redirect = redirect + "?redirect=true&" + s
		} else {
			redirect = redirect + "?redirect=true"
		}
		if logging.IsInfo() {
			p.logger.LogInfof(">>> REDIRECT: %s --> %s", url, redirect)
		}
		http.Redirect(w, httpRequest, redirect, http.StatusSeeOther)
		return
	}
	/*
		If a panic is thrown by ANY handler this defered method will clean up and LOG the event correctly.
	*/
	defer checkForPanicAndRecover(httpRequest, actualResponse)
	/*
		Find the mapping for the url (ReST style)
	*/
	mapping, found := p.mappingElements.GetPathMappingElement(url, httpRequest.Method)
	if !found {
		/*
			Mapping not found,
			delegate to the current error handler to manage the error
		*/
		ThrowPanic("W", 404, SCPathNotFound, fmt.Sprintf("%s URL:%s", httpRequest.Method, url), fmt.Sprintf("METHOD:%s URL:%s is not mapped", httpRequest.Method, url))
	}
	/*
		Add any url parameter names and indexes to the response so we can get ? values
	*/
	actualResponse.names = mapping.names
	/*
		We found a matching function for the request so lets check each before handler to see if we can procceed.
		If a before handler changes the response to an error then we abandon the request and return it's response.
	*/
	p.invokeAllVetoHandlersInList(httpRequest, actualResponse, &p.before)
	if actualResponse.IsAnError() {
		if logging.IsWarn() {
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
				if logging.IsWarn() {
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
func (p *ServerInstanceData) StopServerLater(waitForSeconds int, reason string) {
	p.serverState.State = "STOPPING"
	p.serverClosedReason = reason
	p.serverReturnCode = 0
	go p.stopServerThread(waitForSeconds)
}

/*
SetOsScriptsData - Configure and Validate OS script data
*/
func (p *ServerInstanceData) SetOsScriptsData(path string, scriptsData map[string][]string) {
	if path == "" {
		panic("SetOsScriptsData: A path is required if any OS scripts or applications are to be executed!")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("SetOsScriptsData: The path [" + path + "] for OS scripts or applications could not be found: " + err.Error())
	}
	p.osScriptsPath = path
	if (scriptsData == nil) || (len(scriptsData) == 0) {
		panic("SetOsScriptsData: OS scripts or applications data cannot be empty")
	}
	for name, value := range scriptsData {
		if (value == nil) || (len(value) == 0) {
			panic("SetOsScriptsData: OS scripts or applications data for script [" + name + "]. Parameters cannot be empty")
		}
	}
	p.osScripts = scriptsData
}

/*
GetOsScriptsData - Get the OS Script data for a specific script name
*/
func (p *ServerInstanceData) GetOsScriptsData(scriptName string) []string {
	data := p.osScripts[scriptName]
	if data == nil {
		ThrowPanic("E", 404, SCScriptNotFound, "Not found", "GetOsScriptsData: OS Script ["+scriptName+"] was not found")
	}
	return data
}

/*
GetOsScriptsPath - Get the OS Script data path. This is where all the scripts are.
*/
func (p *ServerInstanceData) GetOsScriptsPath() string {
	return p.osScriptsPath
}

/*
SetStaticFileServerData handle an error response if one occurs
*/
func (p *ServerInstanceData) SetStaticFileServerData(fileServerDataMap map[string]string) {
	p.fileServerData = NewStaticFileServerData(fileServerDataMap)
}

/*
GetStaticFileServerData get the path for a static name
*/
func (p *ServerInstanceData) GetStaticFileServerData() *StaticFileServerData {
	if p.fileServerData == nil {
		ThrowPanic("E", 500, SCStaticFileInit, fmt.Sprintf("File Server Data is undefined"), "Static File Server Data has not been defined.")
	}
	return p.fileServerData
}

/*
GetStaticPathForName get the path for a static name
*/
func (p *ServerInstanceData) GetStaticPathForName(name string) *FileServerContainer {
	return p.GetStaticFileServerData().GetStaticPathForName(name)
}

/*
GetStaticPathForURL get the path for a static url
*/
func (p *ServerInstanceData) GetStaticPathForURL(url string) *FileServerContainer {
	return p.GetStaticFileServerData().GetStaticPathForURL(url)
}

/*
SetPathToTemplates initialise the template system
*/
func (p *ServerInstanceData) SetPathToTemplates(pathToTemplates string) {
	templ, err := loadTemplates(pathToTemplates)
	if err != nil {
		panic(err)
	}
	if templ.HasAnyTemplates() {
		p.templates = templ
		return
	}
	panic("SetPathToTemplates: [" + pathToTemplates + "] did NOT contain any templates")
}

/*
AddTemplateDataProvider add a data provider for a template
*/
func (p *ServerInstanceData) AddTemplateDataProvider(provider func(*http.Request, string, interface{})) {
	if p.templates != nil {
		p.templates.AddDataProvider(provider)
		return
	}
	panic("AddTemplateDataProvider: Templates have not been defined")
}

/*
ListTemplateNames list template names
*/
func (p *ServerInstanceData) ListTemplateNames(delim string) string {
	if p.templates != nil {
		return ListTemplateNames(", ", p.templates.templates)
	}
	return ""
}

/*
HasAnyTemplates list template names
*/
func (p *ServerInstanceData) HasAnyTemplates() bool {
	if p.templates != nil {
		return p.templates.HasAnyTemplates()
	}
	return false
}

/*
HasTemplate list template names
*/
func (p *ServerInstanceData) HasTemplate(templateName string) bool {
	if p.templates != nil {
		return p.templates.HasTemplate(templateName)
	}
	return false
}

/*
TemplateAsString list template names
*/
func (p *ServerInstanceData) TemplateAsString(templateName string, r *http.Request, data interface{}) string {
	if p.HasTemplate(templateName) {
		p.templates.executeDataProvider(templateName, r, data)
		return p.templates.executeString(templateName, data)
	}
	ThrowPanic("W", 404, SCTemplateNotFound, "Not Found", "Template "+templateName+" was not found")
	return ""
}

/*
TemplateWithWriter write a template to the writer
*/
func (p *ServerInstanceData) TemplateWithWriter(w io.Writer, templateName string, r *http.Request, data interface{}) {
	if p.HasTemplate(templateName) {
		p.templates.executeDataProvider(templateName, r, data)
		p.templates.executeWriter(w, templateName, data)
		return
	}
	ThrowPanic("W", 404, SCTemplateNotFound, "Not Found", "Template "+templateName+" was not found")

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
GetServerReturnCode handle an error response if one occurs
*/
func (p *ServerInstanceData) GetServerReturnCode() int {
	return p.serverReturnCode
}

/*
GetServerLogger handle an error response if one occurs
*/
func (p *ServerInstanceData) GetServerLogger() *logging.LoggerDataReference {
	return p.logger
}

/*
GetServerClosedReason handle an error response if one occurs
*/
func (p *ServerInstanceData) GetServerClosedReason() string {
	return p.serverClosedReason
}

/*
GetStatusData server status
*/
func (p *ServerInstanceData) GetStatusData() *StatusData {
	p.serverState.Uptime = time.Now().Unix() - p.serverState.UnixTime
	return p.serverState
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
		AddNewContentTypeToMap(name, value)
	}
}

/*
AddMappedHandler creates a route to a function given a path
*/
func (p *ServerInstanceData) AddMappedHandler(path string, method string, handlerFunc func(*http.Request, *Response)) {
	p.mappingElements.AddPathMappingElement(path, method, handlerFunc)
}

/*
AddMappedHandlerWithNames creates a route to a function given a path and a set of names for each ? in the mapping
*/
func (p *ServerInstanceData) AddMappedHandlerWithNames(path string, method string, handlerFunc func(*http.Request, *Response), names []string) {
	p.mappingElements.AddPathMappingElementWithNames(path, method, handlerFunc, names)
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
	response.AddHeader("Server", []string{p.serverState.Executable})
	/*
		Reflect connection keep-alive. Note - request.Header["Connection"] returns sn array!
	*/
	connection := request.Header["Connection"]
	if len(connection) > 0 {
		response.AddHeader("Connection", connection)
	}
	/*
		If a content type is defined in the response then add Content-Type to the headers.
	*/
	if response.GetContentType() != "" {
		response.GetHeaders()[ContentTypeName] = []string{response.GetContentType() + "; charset=" + p.contentTypeCharset}
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
	if logging.IsAccess() {
		errText := response.GetErrorMessage()
		if errText != "" {
			errText = ": ERROR=" + errText
		}
		p.logger.LogAccessf("<<< STATUS=%d: CODE=%d: RESP=%s%s", response.GetCode(), response.GetSubCode(), response.GetResp(), errText)
		p.LogHeaderMap(response.GetHeaders(), "<-<")
	}
}

/*
ThrowPanic - Throw a PANIC that is handled by the checkForPanicAndRecover method.
This results in the panic being 'recovered' and down graded to an error response.
Parameter level:
 I for log at INFO level
 W for log at WARN level
 Anything else is ERROR level

 Note ALL data is logged.
 StatusCode, errorMessage are returned to the client as Code and Error. For example;
 {"Code":404,"Message":"Not Found","Error":"Parameter XXX Not Found"}
 The Message is derived from the statusCode standard messages (404==Not Found)

 Equivilent to panic("Level|statusCode|Error|info)
*/
func ThrowPanic(level string, statusCode, subCode int, errorText string, logMessage string) {
	panic(fmt.Sprintf("%s|%d|%d|%s|%s", level, statusCode, subCode, errorText, logMessage))
}

/*
ServeContent wraps the http.ServeContent. It opens the file first.
If the open fails it returns an error.
After that it delegates to http.ServeContent
*/
func ServeContent(w *ResponseWriterWrapper, r *http.Request, name string) {
	file, err := os.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			ThrowPanic("W", 404, SCContentNotFound, fmt.Sprintf("URL:%s", r.URL.Path), err.Error())
		}
		ThrowPanic("E", 500, SCContentReadFailed, fmt.Sprintf("URL:%s", r.URL.Path), err.Error())
	}
	defer file.Close()
	http.ServeContent(w, r, name, time.Now(), file)
}

/*
LogHeaderMap logs the header data.

dir will usually be '<-<' or '>->'. Keep it to 3 chars or the logs will look untidy!
*/
func (p *ServerInstanceData) LogHeaderMap(headers map[string][]string, dir string) {
	if logging.IsDebug() {
		for k, v := range headers {
			p.logger.LogDebugf("%s HEADER=%s=%s", dir, k, v)
		}
	}
}

func checkForPanicAndRecover(r *http.Request, response *Response) {
	server := response.GetWrappedServer()
	rec := recover()
	if rec != nil {
		recStr := fmt.Sprintf("%s", rec)
		parts := strings.Split(recStr, "|")
		if len(parts) > 1 {
			rc, err1 := strconv.Atoi(parts[1])
			sub, err2 := strconv.Atoi(parts[2])
			if (err1 == nil) && (err2 == nil) {
				switch strings.ToUpper(parts[0]) {
				case "I":
					if logging.IsInfo() {
						server.logger.LogInfo("--- PANIC|" + recStr[2:])
					}
					break
				case "W":
					if logging.IsWarn() {
						server.logger.LogWarn("--- PANIC|" + recStr[2:])
					}
					break
				case "E":
					if logging.IsError() {
						server.logger.LogError(fmt.Errorf("--- PANIC|%s", recStr[2:]))
					}
					break
				default:
					if logging.IsError() {
						server.logger.LogError(fmt.Errorf("--- PANIC|%s", recStr))
					}
					break
				}
				server.errorHandler(r, response.SetErrorResponse(rc, sub, parts[3]))
				return
			}
		}
		server.serverState.Panics++
		text := fmt.Sprintf("REQUEST:%s MESSAGE:%s", r.URL.Path, recStr)
		server.logger.LogErrorWithStackTrace("!!!", text)
		server.errorHandler(r, response.SetErrorResponse(server.panicStatusCode, SCRuntimeError, recStr))
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
	response.SetContentType(LookupContentType("json"))
	server.PreProcessResponse(request, response)
	server.LogResponse(response)
	fmt.Fprintf(response.GetWrappedWriter(), response.toErrorJSON())
}

func defaultResponseHandler(request *http.Request, response *Response) {
	server := response.GetWrappedServer()
	server.PreProcessResponse(request, response)
	server.LogResponse(response)
	fmt.Fprintf(response.GetWrappedWriter(), response.GetResp())
}

func (p *ServerInstanceData) stopServerThread(waitForSeconds int) {
	if waitForSeconds > 0 {
		time.Sleep(time.Second * time.Duration(waitForSeconds))
	}
	err := p.server.Shutdown(context.TODO())
	if err != nil {
		ThrowPanic("E", 500, SCServerShutDown, "Server Shutdown Failed", err.Error())
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
	if logging.IsAccess() {
		p.logger.LogAccessf("<<< STATUS=%d staticPath:%s ext:%s Content-Type:%s file:%s", response.GetStatusCode(), path, ext, mime, fileName)
		p.LogHeaderMap(response.Header(), "<-<")
	}
}

/*
	logRequest - Log the request.
	Define ACCESS logging to see the request in the logs
	Define DEBUG and ACCESS to see the request and headers in the logs
*/
func (p *ServerInstanceData) logRequest(r *http.Request) {
	if logging.IsAccess() {
		p.logger.LogAccessf(">>> METHOD=%s: REQUEST=%s", r.Method, r.URL.Path)
		p.LogHeaderMap(r.Header, ">->")
	}
}
