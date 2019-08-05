package handlers

import (
	"bytes"
	"fmt"
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
ResetMappingElementTree Clear ALL mappings
*/
func ResetMappingElementTree() {
	mappingElementInstance = nil
}

/*
AddPathMappingElement Add a path to the mapping
*/
func AddPathMappingElement(url string, method string, handlerFunc func(*http.Request) *Response) {
	var me *MappingElements
	var found bool
	parts := strings.Split(strings.Trim(url, "/"), "/")
	if len(parts) == 0 {
		panic("AddPathMappingElement: Url is empty")
	}
	if parts[0] == "?" {
		panic("AddPathMappingElement: Path cannot start with a wildcard '?'")
	}
	currentElement := getMappingElementInstance()
	for _, val := range parts {
		if val != "" {
			me, found = currentElement.elements[val]
			if !found {
				me = newMappingElement()
				currentElement.elements[val] = me
			}
			currentElement = me
		}
		if val == "*" {
			break
		}
	}
	currentElement.HandlerFunc = handlerFunc
	currentElement.RequestMethod = strings.ToUpper(method)
}

/*
GetPathMappingElement Get a path from the mapping
*/
func GetPathMappingElement(url string, method string) (*MappingElements, bool) {
	me, found := getPathMappingElement(strings.Split(strings.Trim(url, "/"), "/"), strings.ToUpper(method), 0, getMappingElementInstance())
	return me, found
}

/*
GetMappingElementTreeString returns a String representing the mapping structure
*/
func GetMappingElementTreeString(heading string) string {
	var b bytes.Buffer
	b.WriteString(heading)
	b.WriteString("\n")
	ce := getMappingElementInstance()
	getMappingElementTreeString(ce, 0, &b)
	return b.String()
}

func getPathMappingElement(parts []string, method string, pos int, me *MappingElements) (*MappingElements, bool) {
	if pos >= len(parts) {
		if me.RequestMethod != method || me.RequestMethod == "" {
			return nil, false
		}
		return me, true
	}
	var foundMe *MappingElements
	var found bool

	foundMe, found = me.elements[parts[pos]]
	if found {
		return getPathMappingElement(parts, method, pos+1, foundMe)
	}
	foundMe, found = me.elements["?"]
	if found {
		return getPathMappingElement(parts, method, pos+1, foundMe)
	}
	foundMe, found = me.elements["*"]
	if found {
		if foundMe.RequestMethod == "" {
			return nil, false
		}
		return foundMe, true
	}
	return nil, false
}

func getMappingElementTreeString(ce *MappingElements, ind int, b *bytes.Buffer) {
	for key, val := range ce.elements {
		b.WriteString(fmt.Sprintf("%s[%s] key:%s method:%s size:%d\n", strings.Repeat(".", ind), ce.RequestMethod, key, val.RequestMethod, len(val.elements)))
		getMappingElementTreeString(val, ind+(1*4), b)
	}
}

func newMappingElement() *MappingElements {
	return &MappingElements{
		elements:      make(map[string]*MappingElements),
		HandlerFunc:   nil,
		RequestMethod: "",
	}
}

func getMappingElementInstance() *MappingElements {
	if mappingElementInstance == nil {
		mappingElementInstance = newMappingElement()
	}
	return mappingElementInstance
}
