package handlers

import (
	"errors"
	"log"
	"net/http"
)

/*
HandlerData is the state of the server
*/
type HandlerData struct {
	before       handlerListData
	mappings     map[string]mappingData
	after        handlerListData
	errorHandler func(http.ResponseWriter, *http.Request, *ErrorResponse)
}

/*
ErrorResponse is the defininition of an  error
*/
type ErrorResponse struct {
	code int
	resp string
	err  error
}

type handlerListData struct {
	handlerFunc func(http.ResponseWriter, *http.Request) *ErrorResponse
	next        *handlerListData
}

type mappingData struct {
	handlerFunc   func(http.ResponseWriter, *http.Request) *ErrorResponse
	requestMethod string
	requestPath   string
}

/*
NewHandlerData Create new HandlerData object
*/
func NewHandlerData() *HandlerData {
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
	}
	return handler
}

/*
NewErrorResponse create an error response
*/
func NewErrorResponse(statusCode int, response string, goErr error) *ErrorResponse {
	if response == "" {
		response = http.StatusText(statusCode)
	}
	if goErr == nil {
		goErr = errors.New("")
	}
	return &ErrorResponse{
		code: statusCode,
		resp: response,
		err:  goErr,
	}
}

/*
SetErrorHandler handle an error is one occurs
*/
func (p *HandlerData) SetErrorHandler(errorHandler func(http.ResponseWriter, *http.Request, *ErrorResponse)) {
	p.errorHandler = errorHandler
}

/*
AddMappedHandler creates a route to a function given a path
*/
func (p *HandlerData) AddMappedHandler(path string, method string, handlerFunc func(http.ResponseWriter, *http.Request) *ErrorResponse) {
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
func (p *HandlerData) AddBeforeHandler(beforeFunc func(http.ResponseWriter, *http.Request) *ErrorResponse) {
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
func (p *HandlerData) AddAfterHandler(afterFunc func(http.ResponseWriter, *http.Request) *ErrorResponse) {
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
	var handlerError *ErrorResponse
	mapping, found := p.mappings[r.URL.String()]
	if !found {
		notFound := NewErrorResponse(404, http.StatusText(404), nil)
		if p.errorHandler == nil {
			logHandlerError(r, notFound)
			http.Error(w, notFound.resp, notFound.code)
		} else {
			p.errorHandler(w, r, notFound)
		}
		return
	}

	handlerError = iterateHandlerList(p, w, r, &p.before)
	if handlerError != nil {
		if p.errorHandler == nil {
			logHandlerError(r, handlerError)
			http.Error(w, handlerError.resp, handlerError.code)
		} else {
			p.errorHandler(w, r, handlerError)
		}
		return
	}

	handlerError = mapping.handlerFunc(w, r)
	if handlerError != nil {
		if p.errorHandler == nil {
			logHandlerError(r, handlerError)
			http.Error(w, handlerError.resp, handlerError.code)
		} else {
			p.errorHandler(w, r, handlerError)
		}
		return
	}

	handlerError = iterateHandlerList(p, w, r, &p.after)
	if handlerError != nil {
		if p.errorHandler == nil {
			logHandlerError(r, handlerError)
			http.Error(w, handlerError.resp, handlerError.code)
		} else {
			p.errorHandler(w, r, handlerError)
		}
		return
	}
}

func iterateHandlerList(p *HandlerData, w http.ResponseWriter, r *http.Request, list *handlerListData) *ErrorResponse {
	for list.next != nil {
		if list.handlerFunc != nil {
			handlerError := list.handlerFunc(w, r)
			if handlerError != nil {
				return handlerError
			}
		}
		list = list.next
	}
	return nil
}

func logHandlerError(r *http.Request, h *ErrorResponse) {
	log.Printf("ERROR: {\"URL\":\"%s\",\"code\":%d,\"err\":\"%s\",\"resp\":\"%s\"}", r.URL.String(), h.code, h.err, h.resp)
}
