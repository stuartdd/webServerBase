package servermain

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestCreateWithWC(t *testing.T) {
	ResetMappingElementTree()
	AddPathMappingElement("a1/b3", "GET4", statusHandler)
	AddPathMappingElement("/a1/b1/c1", "GET1", statusHandler)
	AddPathMappingElement("/a1/?/c3", "GET2", statusHandler)
	AddPathMappingElement("a1/?/c1", "GET3", statusHandler)
	AddPathMappingElement("a2/b1/c1", "GET5", statusHandler)
	AddPathMappingElement("a3/b1/?", "GET6", statusHandler)
	AddPathMappingElement("a3/b2/?/", "GET7", statusHandler)
	AddPathMappingElement("a3//b3//?//", "GET8", statusHandler)
	AddPathMappingElement("a4/*", "GET98", statusHandler)
	AddPathMappingElement("a5", "GET99", statusHandler)
	fmt.Println(GetMappingElementTreeString("----------------- TestCreateWithWC -----------------"))
	assertFound(t, "a4/xx/aa", "GET98")
	assertFound(t, "a4/xx", "GET98")
	assertFound(t, "a4/1/2/3/4/5/6", "GET98")
	assertFound(t, "a5", "GET99")
	assertNotFound(t, "a5/a6", "")
	assertNotFound(t, "a3/b1/xx/aa", "")
	assertFound(t, "a3/b1/xx", "GET6")
	assertFound(t, "a3/b2/123", "GET7")
	assertFound(t, "a3/b3/1234", "GET8")
	assertNotFound(t, "a3/b3/xx/aa", "")
	assertFound(t, "/a1/xx/c3", "GET2")
	assertFound(t, "/a1/b1/c1", "GET1")
	assertFound(t, "a1/xx/c1", "GET3")
	assertFound(t, "a1/b2/c1", "GET3")
	assertFound(t, "a1/b3", "GET4")
	assertFound(t, "a2/b1/c1", "GET5")
	assertNotFound(t, "a1/b5", "")
}

/*
TestCreateSimple test simple (no wildcard) lookup creation
*/
func TestCreateSimple(t *testing.T) {
	ResetMappingElementTree()
	AddPathMappingElement("/a1/b1/c1", http.MethodGet, statusHandler)
	AddPathMappingElement("/a1/b1/c3/", http.MethodGet, statusHandler)
	AddPathMappingElement("/a1/b2/c1", "post", statusHandler)
	AddPathMappingElement("/a1/b3", http.MethodPost, statusHandler)
	AddPathMappingElement("a2/b1/c1", "get", statusHandler)
	fmt.Println(GetMappingElementTreeString("----------------- TestCreateSimple -----------------"))
	assertNotFound(t, "/a1/b5", "")
	assertNotFound(t, "/a1/b5", http.MethodGet)
	assertNotFound(t, "/a1/b5", http.MethodPost)
	assertNotFound(t, "/a1/b1/c1", "Post")
	assertFound(t, "/a1/b1/c1", http.MethodGet)
	assertFound(t, "/a1/b1/c3", http.MethodGet)
	assertFound(t, "/a1/b2/c1", http.MethodPost)
	assertFound(t, "a1/b3", "Post")
	assertNotFound(t, "a1/b3", http.MethodGet)
	assertFound(t, "a2/b1/c1", http.MethodGet)
	assertNotFound(t, "a1/b1/c5", "")
	assertNotFound(t, "a1/b3/b6", "")
	assertNotFound(t, "a2/b3/b6", "")
	assertNotFound(t, "a2/b1/b6", "")
}

func assertFound(t *testing.T, url string, method string) {
	me, found := GetPathMappingElement(url, method)
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

func assertNotFound(t *testing.T, url string, method string) {
	me, found := GetPathMappingElement(url, method)
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
