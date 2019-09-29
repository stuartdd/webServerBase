package servermain

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stuartdd/webServerBase/test"
)

func TestUrlParamNamesNotEnoughNames(t *testing.T) {
	url := []string{"a", "?", "c", "?"}
	names := []string{"namea"}
	defer test.AssertPanicAndRecover(t, "Not enough names")
	validateNames(url, names)
}

func TestUrlParamNamesNilNames(t *testing.T) {
	url := []string{"a", "?", "c", "?"}
	defer test.AssertPanicAndRecover(t, "No names")
	validateNames(url, nil)
}

func TestUrlParamNamesNoNames(t *testing.T) {
	url := []string{"a", "?"}
	names := []string{}
	defer test.AssertPanicAndRecover(t, "No names")
	validateNames(url, names)
}

func TestUrlParamNamesNoQM(t *testing.T) {
	url := []string{"a", "b", "c", "d"}
	names := []string{"namea"}
	defer test.AssertPanicAndRecover(t, "Too many names")
	validateNames(url, names)
}

func TestUrlParamNamesTooManyNames(t *testing.T) {
	url := []string{"a", "?", "c", "?"}
	names := []string{"namea", "nameb", "namec"}
	defer test.AssertPanicAndRecover(t, "Too many names")
	validateNames(url, names)
}

func TestUrlParamNamesDuplicates(t *testing.T) {
	url := []string{"a", "?", "c", "?", "d", "?"}
	names := []string{"namea", "nameb", "namea"}
	defer test.AssertPanicAndRecover(t, "Duplicate names")
	validateNames(url, names)
}

func TestUrlParamNames(t *testing.T) {
	url := []string{"a", "?", "c", "?"}
	names := []string{"namea", "namec"}
	m := validateNames(url, names)
	test.AssertIntEqual(t, "", 2, len(m))
	test.AssertIntEqual(t, "", 1, m["namea"])
	test.AssertIntEqual(t, "", 3, m["namec"])
}

func TestUrlParamNames3(t *testing.T) {
	url := []string{"a", "?", "c", "?", "?"}
	names := []string{"namea", "namec", "named"}
	m := validateNames(url, names)
	test.AssertIntEqual(t, "", 3, len(m))
	test.AssertIntEqual(t, "", 1, m["namea"])
	test.AssertIntEqual(t, "", 3, m["namec"])
	test.AssertIntEqual(t, "", 4, m["named"])
}

func TestCreateWithWC(t *testing.T) {
	names := []string{"ABC"}
	m := NewMappingElements(nil)
	m.AddPathMappingElement("a1/b3", "GET4", statusHandler)
	m.AddPathMappingElement("/a1/b1/c1", "GET1", statusHandler)
	m.AddPathMappingElementWithNames("/a1/?/c3", "GET2", statusHandler, names)
	m.AddPathMappingElementWithNames("a1/?/c1", "GET3", statusHandler, names)
	m.AddPathMappingElement("a2/b1/c1", "GET5", statusHandler)
	m.AddPathMappingElementWithNames("a3/b1/?", "GET6", statusHandler, names)
	m.AddPathMappingElementWithNames("a3/b2/?/", "GET7", statusHandler, names)
	m.AddPathMappingElementWithNames("a3//b3//?//", "GET8", statusHandler, names)
	m.AddPathMappingElement("a4/*", "GET98", statusHandler)
	m.AddPathMappingElement("a5", "GET99", statusHandler)
	fmt.Println(m.GetMappingElementTreeString("----------------- TestCreateWithWC -----------------"))
	assertFound(t, m, "a4/xx/aa", "GET98")
	assertFound(t, m, "a4/xx", "GET98")
	assertFound(t, m, "a4/1/2/3/4/5/6", "GET98")
	assertFound(t, m, "a5", "GET99")
	assertNotFound(t, m, "a5/a6", "")
	assertNotFound(t, m, "a3/b1/xx/aa", "")
	assertFound(t, m, "a3/b1/xx", "GET6")
	assertFound(t, m, "a3/b2/123", "GET7")
	assertFound(t, m, "a3/b3/1234", "GET8")
	assertNotFound(t, m, "a3/b3/xx/aa", "")
	assertFound(t, m, "/a1/xx/c3", "GET2")
	assertFound(t, m, "/a1/b1/c1", "GET1")
	assertFound(t, m, "a1/xx/c1", "GET3")
	assertFound(t, m, "a1/b2/c1", "GET3")
	assertFound(t, m, "a1/b3", "GET4")
	assertFound(t, m, "a2/b1/c1", "GET5")
	assertNotFound(t, m, "a1/b5", "")
}

/*
TestCreateSimple test simple (no wildcard) lookup creation
*/
func TestCreateSimple(t *testing.T) {
	m := NewMappingElements(nil)
	m.AddPathMappingElement("/a1/b1/c1", http.MethodGet, statusHandler)
	m.AddPathMappingElement("/a1/b1/c3/", http.MethodPost, statusHandler)
	m.AddPathMappingElement("/a1/b2/c1", "post", statusHandler)
	m.AddPathMappingElement("/a1/b3", http.MethodGet, statusHandler)
	m.AddPathMappingElement("a2/b1/c1", "get", statusHandler)
	fmt.Println(m.GetMappingElementTreeString("----------------- TestCreateSimple -----------------"))
	assertFound(t, m, "/a1/b1/c1", http.MethodGet)
	assertNotFound(t, m, "/a1/b1/c1", http.MethodPost)
	assertNotFound(t, m, "/a1/b1/c1", "")
	assertNotFound(t, m, "/a1/b5", http.MethodGet)
	assertNotFound(t, m, "/a1/b5", http.MethodPost)
	assertNotFound(t, m, "/a1/b5", "")

	assertFound(t, m, "/a1/b1/c3", http.MethodPost)
	assertNotFound(t, m, "/a1/b1/c3", http.MethodGet)
	assertFound(t, m, "/a1/b1/c3", "post") // Test not case sensitive
	assertFound(t, m, "/a1/b1/c3", "Post") // Test not case sensitive

	assertNotFound(t, m, "a1/b3", http.MethodPost)
	assertFound(t, m, "a1/b3", "Get")
	assertFound(t, m, "a1/b3", http.MethodGet)
	assertFound(t, m, "a2/b1/c1", http.MethodGet)

	assertNotFound(t, m, "a1/b6", http.MethodPost)

}

func TestFindRoot(t *testing.T) {
	meRoot := NewMappingElements(nil)
	meRoot.RequestMethod = "ROOT"
	test.AssertNil(t, "", meRoot.parent)
	test.AssertStringEquals(t, "", "ROOT", meRoot.RequestMethod)
	test.AssertStringEquals(t, "", "ROOT", meRoot.findRoot().RequestMethod)
	test.AssertNil(t, "", meRoot.findRoot().parent)

	meNext := NewMappingElements(meRoot)
	meNext.RequestMethod = "NEXT"
	test.AssertStringEquals(t, "", "NEXT", meNext.RequestMethod)
	test.AssertNil(t, "", meRoot.parent)
	test.AssertStringEquals(t, "", "ROOT", meNext.parent.RequestMethod)
	test.AssertStringEquals(t, "", "ROOT", meNext.findRoot().RequestMethod)
	test.AssertNil(t, "", meNext.findRoot().parent)

	meLast := NewMappingElements(meNext)
	meLast.RequestMethod = "LAST"
	test.AssertStringEquals(t, "", "LAST", meLast.RequestMethod)
	test.AssertNil(t, "", meRoot.parent)
	test.AssertStringEquals(t, "", "ROOT", meRoot.RequestMethod)
	test.AssertStringEquals(t, "", "ROOT", meNext.parent.RequestMethod)
	test.AssertStringEquals(t, "", "NEXT", meLast.parent.RequestMethod)
	test.AssertStringEquals(t, "", "ROOT", meLast.findRoot().RequestMethod)
	test.AssertNil(t, "", meLast.findRoot().parent)
}

func assertFound(t *testing.T, p *MappingElements, url string, method string) {
	root := p.findRoot()

	me, found := root.GetPathMappingElement(url, method)
	if !found {
		t.Fatalf("Mapping Not found! %s", url)
	}
	if me == nil {
		t.Fatalf("Mapping Found! %s but returned element is nil", url)
	}
	if me.RequestMethod != strings.ToUpper(method) {
		t.Fatalf("Mapping Found! %s but returned request method %s should be %s", url, me.RequestMethod, strings.ToUpper(method))
	}
	if me.HandlerFunc == nil {
		t.Fatalf("Mapping Found! %s but returned handler function is nil", url)
	}
}

func assertNotFound(t *testing.T, p *MappingElements, url string, method string) {
	root := p.findRoot()

	me, found := root.GetPathMappingElement(url, method)
	if me != nil {
		t.Fatalf("Mapping Not Found! %s but returned element is NOT nil", url)
	}
	if found {
		t.Fatalf("Mapping Found! %s. Should not be found", url)
	}
}

func statusHandler(request *http.Request, response *Response) {
	response.SetErrorResponse(400, 0, "")
}
