package dto

import (
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
func AddPathMappingElement(url string, method string, handlerFunc func(*http.Request) *Response) {
	var me *MappingElements
	var found bool

	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		panic("AddPathMappingElement: Url is empty")
	}
	if parts[0] == "*" {
		panic("AddPathMappingElement: Path cannot start with a wildcard '*'")
	}
	currentElement := GetMappingElementInstance()
	for _, val := range parts {
		me, found = currentElement.elements[val]
		if !found {
			me = NewMappingElement()
			currentElement.elements[val] = me
		}
		currentElement = me
	}
	currentElement.HandlerFunc = handlerFunc
	currentElement.RequestMethod = method
}

func getPathMappingElement(parts []string, pos int, me *MappingElements) (*MappingElements, bool) {
	if pos >= len(parts) {
		return me, true
	}
	me2, found2 := me.elements[parts[pos]]
	if found2 {
		return getPathMappingElement(parts, pos+1, me2)
	}
	me1, found1 := me.elements["*"]
	if found1 {
		return getPathMappingElement(parts, pos+1, me1)
	}
	return nil, false
}

/*
GetPathMappingElement Get a path from the mapping
*/
func GetPathMappingElement(url string) (*MappingElements, bool) {
	me, found := getPathMappingElement(strings.Split(url, "/"), 0, GetMappingElementInstance())
	return me, found
}

func printTree(ce *MappingElements, ind int) {
	for key, val := range ce.elements {
		fmt.Printf("%s[%s] key:%s method:%s size:%d\n", strings.Repeat(".", ind), ce.RequestMethod, key, val.RequestMethod, len(val.elements))
		printTree(val, ind+(1*4))
	}
}

/*
PrintTree prints the structure of the mappings
*/
func PrintTree() {
	ce := GetMappingElementInstance()
	printTree(ce, 0)
}
