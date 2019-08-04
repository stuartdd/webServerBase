package dto

import (
	"net/http"
	"testing"
)

func TestCreateWithWC(t *testing.T) {
	AddPathMappingElement("a1/b3", "GET4", statusHandler)
	AddPathMappingElement("a1/b1/c1", "GET1", statusHandler)
	AddPathMappingElement("a1/*/c3", "GET2", statusHandler)
	AddPathMappingElement("a1/*/c1", "GET3", statusHandler)
	AddPathMappingElement("a2/b1/c1", "GET5", statusHandler)
	PrintTree()
	assertFound(t, "a1/xx/c3", "GET2")
	assertFound(t, "a1/b1/c1", "GET1")
	assertFound(t, "a1/xx/c1", "GET3")
	assertFound(t, "a1/b2/c1", "GET3")
	assertFound(t, "a1/b3", "GET4")
	assertFound(t, "a2/b1/c1", "GET5")
}

func TestCreateAmbiguous(t *testing.T) {
	AddPathMappingElement("a1/b1/c1", "GET", statusHandler)
	AddPathMappingElement("a1/*/c3", "GET", statusHandler)
	AddPathMappingElement("a1/*/c1", "GET", statusHandler)
	AddPathMappingElement("a1/b3", "GET", statusHandler)
}

/*
TestCreateSimple test simple (no wildcard) lookup creation
*/
func TestCreateSimple(t *testing.T) {
	AddPathMappingElement("a1/b1/c1", "GET", statusHandler)
	AddPathMappingElement("a1/b1/c3", "GET", statusHandler)
	AddPathMappingElement("a1/b2/c1", "GET", statusHandler)
	AddPathMappingElement("a1/b3", "GET", statusHandler)
	AddPathMappingElement("a2/b1/c1", "GET", statusHandler)

	assertFound(t, "a1/b1/c1", "GET")
	assertFound(t, "a1/b1/c3", "GET")
	assertFound(t, "a1/b2/c1", "GET")
	assertFound(t, "a1/b3", "GET")
	assertFound(t, "a2/b1/c1", "GET")

	assertNotFound(t, "a1/b1/c5")
	assertNotFound(t, "a1/b5")
	assertNotFound(t, "a1/b3/b6")
	assertNotFound(t, "a2/b3/b6")
	assertNotFound(t, "a2/b1/b6")
}

func assertFound(t *testing.T, url string, method string) {
	me, found := GetPathMappingElement(url)
	if !found {
		t.Fatalf("Mapping Not found! %s", url)
	}
	if me == nil {
		t.Fatalf("Mapping Found! %s but returned element is nil", url)
	}
	if me.RequestMethod != method {
		t.Fatalf("Mapping Found! %s but returned request method %s should be %s", url, me.RequestMethod, method)
	}
	if me.HandlerFunc == nil {
		t.Fatalf("Mapping Found! %s but returned handler function is nil", url)
	}
}

func assertNotFound(t *testing.T, url string) {
	me, found := GetPathMappingElement(url)
	if found {
		t.Fatalf("Mapping Found! %s. Should not be found", url)
	}
	if me != nil {
		t.Fatalf("Mapping Not Found! %s but returned element is NOT nil", url)
	}
}

func statusHandler(r *http.Request) *Response {
	return NewResponse(400, "", "", nil)
}
