package servermain

import (
	"errors"
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
	resp := NewResponse(200, "String", "String", nil)
	test.AssertEqualString(t, "Response should be String", "String", resp.GetResp())
	test.AssertFalse(t, "", resp.IsAnError())
	test.AssertEqualString(t, "Response ContentType should be String", "String", resp.GetResp())
	test.AssertNil(t, "", resp.GetError())
}

func TestRespAsInt(t *testing.T) {
	resp := NewResponse(300, 90, "Int", errors.New("MeError"))
	test.AssertEqualString(t, "GetResp()", "90", resp.GetResp())
	test.AssertTrue(t, "", resp.IsAnError())
	test.AssertEqualString(t, "GetContentType()", "Int", resp.GetContentType())
	test.AssertNotNil(t, "", resp.GetError())
	test.AssertErrorString(t, "", resp.GetError(), "MeError")
}

func TestRespAsStructWithHeader(t *testing.T) {
	resp := NewResponse(299, testStruct{
		A: "A",
		B: true,
		C: 72.8,
		D: 99,
	}, "testStruct", nil)
	json := resp.GetResp()
	test.AssertEqualString(t, "GetResp()", "{\"A\":\"A\",\"B\":true,\"C\":72.8,\"D\":99}", json)
	test.AssertFalse(t, "", resp.IsAnError())
	test.AssertEqualString(t, "GetContentType()", "testStruct", resp.GetContentType())
	test.AssertNil(t, "", resp.GetError())

	resp.AddHeader("HI", []string{"A", "B"})
	hi := resp.GetHeaders()["HI"]
	test.AssertEqualInt(t, "", 2, len(hi))
	test.AssertEqualString(t, "Header[HI][0]", "A", hi[0])
	test.AssertEqualString(t, "Header[HI][1]", "B", hi[1])
}
