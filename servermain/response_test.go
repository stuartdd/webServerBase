package servermain

import (
	"testing"
	"webServerBase/test"
)

type testStruct struct {
	A string
	B bool
	C float64
	D int
}

func TestRespAsString(t *testing.T) {
	resp := NewResponse(nil, nil).SetResponse(200, "String", LookupContentType("json"))
	test.AssertEqualString(t, "Response should be String", "String", resp.GetResp())
	test.AssertFalse(t, "", resp.IsAnError())
	test.AssertEqualString(t, "Response ContentType should be String", LookupContentType("json"), resp.GetContentType())
	test.AssertEmptyString(t, "", resp.GetErrorMessage())
}

func TestCanAddHeaders(t *testing.T) {
	resp := NewResponse(nil, nil)
	resp.AddHeader("H1", []string{"a", "b"})
	test.AssertEqualInt(t, "", 2, len(resp.GetHeaders()["H1"]))
	test.AssertEqualString(t, "", "a", resp.GetHeaders()["H1"][0])
	test.AssertEqualString(t, "", "b", resp.GetHeaders()["H1"][1])
}

func TestIsError(t *testing.T) {
	resp := NewResponse(nil, nil)
	test.AssertFalse(t, "", resp.IsAnError())
	test.AssertFalse(t, "", resp.IsClosed())
	test.AssertEqualInt(t, "", 200, resp.GetCode())
}

func TestRespAsInt(t *testing.T) {
	resp := NewResponse(nil, nil).SetErrorResponse(300, 90, "MeError")
	test.AssertEqualString(t, "", "Multiple Choices", resp.GetResp())
	test.AssertTrue(t, "", resp.IsAnError())
	test.AssertEmptyString(t, "", resp.GetContentType())
	test.AssertEqualString(t, "", "MeError", resp.GetErrorMessage())
}

func TestRespAsStructWithHeader(t *testing.T) {
	resp := NewResponse(nil, nil).SetResponse(299, testStruct{
		A: "A",
		B: true,
		C: 72.8,
		D: 99,
	}, "application/json")
	json := resp.GetResp()
	test.AssertEqualString(t, "GetResp()", "{\"A\":\"A\",\"B\":true,\"C\":72.8,\"D\":99}", json)
	test.AssertFalse(t, "", resp.IsAnError())
	test.AssertEqualString(t, "GetContentType()", "application/json", resp.GetContentType())
	test.AssertEmptyString(t, "", resp.GetErrorMessage())

	resp.AddHeader("HI", []string{"A", "B"})
	hi := resp.GetHeaders()["HI"]
	test.AssertEqualInt(t, "", 2, len(hi))
	test.AssertEqualString(t, "Header[HI][0]", "A", hi[0])
	test.AssertEqualString(t, "Header[HI][1]", "B", hi[1])

	test.AssertEqualString(t, "", "{\"A\":\"A\",\"B\":true,\"C\":72.8,\"D\":99}", resp.GetResp())
}
