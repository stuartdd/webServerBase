package servermain

import (
	"testing"

	"github.com/stuartdd/webServerBase/test"
)

type testStruct struct {
	A string
	B bool
	C float64
	D int
}

func TestRespAsString(t *testing.T) {
	resp := NewResponse(nil, nil, "TXID").SetResponse(200, "String", LookupContentType("json"))
	test.AssertStringEquals(t, "Response should be String", "String", resp.GetResp())
	test.AssertBoolFalse(t, "", resp.IsAnError())
	test.AssertStringEquals(t, "Response ContentType should be String", LookupContentType("json"), resp.GetContentType())
	test.AssertStringEmpty(t, "", resp.GetErrorMessage())
}

func TestCanAddHeaders(t *testing.T) {
	resp := NewResponse(nil, nil, "TXID")
	resp.AddHeader("H1", []string{"a", "b"})
	test.AssertIntEqual(t, "", len(resp.GetHeaders()["H1"]), 2)
	test.AssertStringEquals(t, "", "a", resp.GetHeaders()["H1"][0])
	test.AssertStringEquals(t, "", "b", resp.GetHeaders()["H1"][1])
}

func TestIsError(t *testing.T) {
	resp := NewResponse(nil, nil, "TXID")
	test.AssertBoolFalse(t, "", resp.IsAnError())
	test.AssertBoolFalse(t, "", resp.IsClosed())
	test.AssertIntEqual(t, "", resp.GetCode(), 200)
}

func TestRespAsInt(t *testing.T) {
	resp := NewResponse(nil, nil, "TXID").SetErrorResponse(300, 90, "MeError")
	test.AssertStringEquals(t, "", "Multiple Choices", resp.GetResp())
	test.AssertBoolTrue(t, "", resp.IsAnError())
	test.AssertStringEmpty(t, "", resp.GetContentType())
	test.AssertStringEquals(t, "", "MeError", resp.GetErrorMessage())
}

func TestRespAsStructWithHeader(t *testing.T) {
	resp := NewResponse(nil, nil, "TXID").SetResponse(299, testStruct{
		A: "A",
		B: true,
		C: 72.8,
		D: 99,
	}, "application/json")
	json := resp.GetResp()
	test.AssertStringEquals(t, "GetResp()", "{\"A\":\"A\",\"B\":true,\"C\":72.8,\"D\":99}", json)
	test.AssertBoolFalse(t, "", resp.IsAnError())
	test.AssertStringEquals(t, "GetContentType()", "application/json", resp.GetContentType())
	test.AssertStringEmpty(t, "", resp.GetErrorMessage())

	resp.AddHeader("HI", []string{"A", "B"})
	hi := resp.GetHeaders()["HI"]
	test.AssertIntEqual(t, "", len(hi), 2)
	test.AssertStringEquals(t, "Header[HI][0]", "A", hi[0])
	test.AssertStringEquals(t, "Header[HI][1]", "B", hi[1])

	test.AssertStringEquals(t, "", "{\"A\":\"A\",\"B\":true,\"C\":72.8,\"D\":99}", resp.GetResp())
}
