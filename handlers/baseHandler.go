package handlers

import (
	"fmt"
	"net/http"
	"webServerBase/dto"
	"webServerBase/logging"
	"webServerBase/state"
)

/*
HandlerData is the state of the server
*/
type HandlerData struct {
	before         handlerListData
	mappings       map[string]mappingData
	after          handlerListData
	errorHandler   func(http.ResponseWriter, *http.Request, *dto.Response)
	defaultHandler func(http.ResponseWriter, *http.Request, *dto.Response)
}

type handlerListData struct {
	handlerFunc func(*http.Request) *dto.Response
	next        *handlerListData
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
func NewHandlerData() *HandlerData {
	logger = logging.NewLogger("BaseHandler")
	handler := &HandlerData{
		before: handlerListData{
			handlerFunc: nil,
			next:        nil,
		},
		mappings: make(map[string]mappingData),
		after: handlerListData{
			handlerFunc: nil,
			next:        nil,
		},
		errorHandler:   defaultErrorResponseHandler,
		defaultHandler: defaultResponseHandler,
	}
	return handler
}

/*
SetErrorHandler handle an error response if one occurs
*/
func (p *HandlerData) SetErrorHandler(errorHandler func(http.ResponseWriter, *http.Request, *dto.Response)) {
	p.errorHandler = errorHandler
}

/*
SetHandler handle a NON error response
*/
func (p *HandlerData) SetHandler(handler func(http.ResponseWriter, *http.Request, *dto.Response)) {
	p.defaultHandler = handler
}

/*
AddMappedHandler creates a route to a function given a path
*/
func (p *HandlerData) AddMappedHandler(path string, method string, handlerFunc func(*http.Request) *dto.Response) {
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
func (p *HandlerData) AddBeforeHandler(beforeFunc func(*http.Request) *dto.Response) {
	bef := &p.before
	for bef.next != nil {
		bef = bef.next
	}
	bef.handlerFunc = beforeFunc
	bef.next = &handlerListData{
		handlerFunc: nil,
		next:        nil,
	}
}

/*
AddAfterHandler adds a function called after the mapping function
*/
func (p *HandlerData) AddAfterHandler(afterFunc func(*http.Request) *dto.Response) {
	aft := &p.after
	for aft.next != nil {
		aft = aft.next
	}
	aft.handlerFunc = afterFunc
	aft.next = &handlerListData{
		handlerFunc: nil,
		next:        nil,
	}
}

/*
ServeHTTP handle ALL calls
*/
func (p *HandlerData) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handlerResponse *dto.Response
	logger.LogDebugf("METHOD=%s: REQUEST=%s", r.Method, r.URL.String())
	/*
		Find the mapping
	*/
	mapping, found := p.mappings[r.URL.String()]
	if !found {
		/*
			Mapping was not found
		*/
		handlerResponse = dto.NewResponse(404, http.StatusText(404), "", nil)
		logHandlerResponse(r, handlerResponse)
		p.errorHandler(w, r, handlerResponse)
		return
	}

	handlerResponse = invokeAllHandlersInList(p, w, r, &p.before)
	if handlerResponse != nil {
		logHandlerResponse(r, handlerResponse)
		p.errorHandler(w, r, handlerResponse)
		return
	}

	handlerResponse = mapping.handlerFunc(r)
	if handlerResponse != nil {
		logHandlerResponse(r, handlerResponse)
		if handlerResponse != nil {
			if handlerResponse.IsError() {
				p.errorHandler(w, r, handlerResponse)
				return
			}
			p.defaultHandler(w, r, handlerResponse)
		}
	}

	handlerResponse = invokeAllHandlersInList(p, w, r, &p.after)
	if handlerResponse != nil {
		logHandlerResponse(r, handlerResponse)
		p.errorHandler(w, r, handlerResponse)
		return
	}
}

func defaultErrorResponseHandler(w http.ResponseWriter, request *http.Request, response *dto.Response) {
	http.Error(w, response.GetResp(), response.GetCode())
}

func defaultResponseHandler(w http.ResponseWriter, request *http.Request, response *dto.Response) {
	for key, value := range response.GetHeaders() {
		w.Header().Set(key, value)
	}
	if response.GetContentType() != "" {
		w.Header().Set("Content-Type", response.GetContentType())
	}
	w.Header().Set("Server", state.GetStatusDataExecutableName())
	w.WriteHeader(response.GetCode())
	fmt.Fprintf(w, response.GetResp())
}

/*
invokeAllHandlersInList
Invoke ALL handlers in the list UNTIL a handler returns a response.
Any response is considered an ERROR.
*/
func invokeAllHandlersInList(p *HandlerData, w http.ResponseWriter, r *http.Request, list *handlerListData) *dto.Response {
	for list.next != nil {
		if list.handlerFunc != nil {
			handlerResponse := list.handlerFunc(r)
			if handlerResponse != nil {
				return handlerResponse
			}
		}
		list = list.next
	}
	return nil
}

func logHandlerResponse(r *http.Request, response *dto.Response) {
	logger.LogDebugf("%s: STATUS=%d: RESP=%s", response.GetType(), response.GetCode(), response.GetResp())
}
