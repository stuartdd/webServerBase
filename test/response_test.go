package test

import (
	"errors"
	"testing"
	"webServerBase/handlers"
)

type testStruct struct {
	A string
	B bool
	C float64
	D int
}

func TestRespAsString(t *testing.T) {
	resp := handlers.NewResponse(200, "String", "String", nil)
	if "String" != resp.GetResp() {
		t.Fatal("Response ne String")
	}
	if resp.IsNot200() {
		t.Fatal("Response IsNot200 should be false")
	}
	if resp.GetContentType() != "String" {
		t.Fatal("Response ContentType should be String")
	}
	if resp.GetError() != nil {
		t.Fatal("Response GetError should be nil")
	}
}

func TestRespAsInt(t *testing.T) {
	resp := handlers.NewResponse(300, 90, "Int", errors.New("MeError"))
	if "90" != resp.GetResp() {
		t.Fatal("Response ne 90")
	}
	if !resp.IsNot200() {
		t.Fatal("Response IsNot200 should be true")
	}
	if resp.GetContentType() != "Int" {
		t.Fatal("Response ContentType should be Int")
	}
	if resp.GetError() == nil {
		t.Fatal("Response GetError should NOT be nil")
	}
	if resp.GetError().Error() != "MeError" {
		t.Fatal("Response GetError should return Error=MeError")
	}
}

func TestRespAsStruct(t *testing.T) {
	resp := handlers.NewResponse(299, testStruct{
		A: "A",
		B: true,
		C: 72.8,
		D: 99,
	}, "testStruct", nil)
	json := resp.GetResp()
	if "{\"A\":\"A\",\"B\":true,\"C\":72.8,\"D\":99}" != json {
		t.Fatal("Response ne JSON")
	}
	if resp.IsNot200() {
		t.Fatal("Response IsNot200 should be false")
	}
	if resp.GetContentType() != "testStruct" {
		t.Fatal("Response ContentType should be String")
	}
	if resp.GetError() != nil {
		t.Fatal("Response GetError should be nil")
	}

	resp.AddHeader("HI", []string{"A", "B"})

	hi := resp.GetHeaders()["HI"]
	if len(hi) != 2 {
		t.Fatal("Response GetHeader should be string array len 2")
	}
	if hi[0] != "A" {
		t.Fatal("Response GetHeader index 0 not A")
	}
	if hi[1] != "B" {
		t.Fatal("Response GetHeader index 1 not B")
	}
}
