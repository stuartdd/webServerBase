package servermain

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

/*
MappingElements searchable tree
*/
type MappingElements struct {
	elements      map[string]*MappingElements
	HandlerFunc   func(*http.Request, *Response)
	RequestMethod string
	names         map[string]int
	parent        *MappingElements
}

/*
NewMappingElements Create a new MappingElements with an empty tree
*/
func NewMappingElements(parent *MappingElements) *MappingElements {
	return &MappingElements{
		elements:      make(map[string]*MappingElements),
		HandlerFunc:   nil,
		RequestMethod: "",
		names:         make(map[string]int),
		parent:        parent,
	}
}

func (p *MappingElements) findRoot() *MappingElements {
	for p.parent != nil {
		p = p.parent
	}
	return p
}

/*
ResetMappingElementTree Clear ALL mappings
*/
func (p *MappingElements) ResetMappingElementTree() {
	p.elements = make(map[string]*MappingElements)
	p.HandlerFunc = nil
	p.RequestMethod = ""
}

func validateNames(urlParts []string, names []string) map[string]int {
	noNames := ((names == nil) || (len(names) == 0))
	m := make(map[string]int)
	namePos := 0
	for urlPos, str := range urlParts {
		if str == "?" {
			if noNames {
				panic("AddPathMappingElementWithNames: No names were provided for the url parameters")
			}
			if namePos >= len(names) {
				panic("AddPathMappingElementWithNames: Not enough names for the number of url parameters")
			}
			_, ok := m[names[namePos]]
			if ok {
				panic("AddPathMappingElementWithNames: Duplicate names for url parameters")
			}
			m[names[namePos]] = urlPos
			namePos++
		}
	}
	if len(m) != len(names) {
		panic("AddPathMappingElementWithNames: Too many names for the number of url parameters")
	}
	return m
}

/*
AddPathMappingElement Add a path to the mapping
*/
func (p *MappingElements) AddPathMappingElement(url string, method string, handlerFunc func(*http.Request, *Response)) {
	p.AddPathMappingElementWithNames(url, method, handlerFunc, nil)
}

/*
AddPathMappingElementWithNames Add a path to the mapping
*/
func (p *MappingElements) AddPathMappingElementWithNames(url string, method string, handlerFunc func(*http.Request, *Response), names []string) {
	var me *MappingElements
	var found bool
	parts := strings.Split(strings.Trim(url, "/"), "/")
	if len(parts) == 0 {
		panic("AddPathMappingElement: Url is empty")
	}
	if parts[0] == "?" {
		panic("AddPathMappingElement: Path cannot start with a wildcard '?'")
	}
	currentElement := p
	for _, val := range parts {
		if val != "" {
			me, found = currentElement.elements[val]
			if !found {
				me = NewMappingElements(currentElement)
				currentElement.elements[val] = me
			}
			currentElement = me
		}
		if val == "*" {
			break
		}
	}
	currentElement.HandlerFunc = handlerFunc
	currentElement.names = validateNames(parts, names)
	currentElement.RequestMethod = strings.ToUpper(method)
}

/*
GetPathMappingElement Get a path from the mapping
*/
func (p *MappingElements) GetPathMappingElement(url string, method string) (*MappingElements, bool) {
	return p.getPathMappingElement(strings.Split(strings.Trim(url, "/"), "/"), strings.ToUpper(method), 0, p)
}

/*
GetMappingElementTreeString returns a String representing the mapping structure
*/
func (p *MappingElements) GetMappingElementTreeString(heading string) string {
	var b bytes.Buffer
	b.WriteString(heading)
	b.WriteString("\n")
	getMappingElementTreeString(p, 0, &b)
	return b.String()
}

func (p *MappingElements) getPathMappingElement(parts []string, method string, pos int, me *MappingElements) (*MappingElements, bool) {
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
		return p.getPathMappingElement(parts, method, pos+1, foundMe)
	}
	foundMe, found = me.elements["?"]
	if found {
		return p.getPathMappingElement(parts, method, pos+1, foundMe)
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
