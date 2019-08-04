package dto

import (
	"net/http"
	"strings"
)

var mappingElementInstance *MappingElements

/*
MappingElements searchable tree
*/
type MappingElements struct {
	elements      map[string]*MappingElements
	HandlerFunc   func(*http.Request) *Response
	RequestMethod string
}

/*
GetMappingElementInstance c
reate a note. It roo is nil then create it
*/
func GetMappingElementInstance() *MappingElements {
	if mappingElementInstance == nil {
		mappingElementInstance = NewMappingElement()
	}
	return mappingElementInstance
}

/*
NewMappingElement c
reate a note. It roo is nil then create it
*/
func NewMappingElement() *MappingElements {
	return &MappingElements{
		elements: make(map[string]*MappingElements),
	}
}

/*
AddPathMappingElement Add a path to the mapping
*/
func AddPathMappingElement(url string) {
	var me *MappingElements
	var found bool

	parts := strings.Split(url, "/")
	currentElement := GetMappingElementInstance()
	for _, val := range parts {
		me, found = currentElement.elements[val]
		if !found {
			me = NewMappingElement()
			currentElement.elements[val] = me
		}
		currentElement = me
	}
}

/*
AddPath Add a path to the mapping
*/
func GetPathMappingElement(url string) (*MappingElements, bool) {
	return GetMappingElementInstance(), true
}
