package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"webServerBase/config"
	"webServerBase/logging"
	"webServerBase/servermain"
	"webServerBase/test"
)

var configData *config.Data
var testLog = logging.CreateTestLogger("Test-Logger")

func TestServer(t *testing.T) {
	headers := make(map[string]string)

	startServer(t)
	test.AssertStringContains(t, "", sendGet(t, 404, "static/testfile", nil), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCContentNotFound), "Not Found", "URL:/static/testfile"})

	headers[servermain.ContentTypeName] = servermain.LookupContentType("json")
	headers[servermain.ContentLengthName] = "13"
	test.AssertStringContains(t, "", sendGet(t, 200, "static/testfile.json", headers), []string{"{\"test\":true}"})

	headers[servermain.ContentTypeName] = servermain.LookupContentType("xml")
	headers[servermain.ContentLengthName] = "17"
	test.AssertStringContains(t, "", sendGet(t, 200, "static/testfile.xml", headers), []string{"<test>true</test>"})

	headers[servermain.ContentTypeName] = servermain.LookupContentType("ico")
	headers[servermain.ContentLengthName] = "367958"
	test.AssertStringContains(t, "", sendGet(t, 200, "static/arduino.ico", headers), []string{})

	testFile := configData.GetConfigDataStaticFilePathForOS()["static"] + string(os.PathSeparator) + "testFile.txt"
	defer deleteFile(t, testFile)

	prc := configData.PanicResponseCode
	test.AssertStringContains(t, "", sendPost(t, 404, "path/god/file/testFile", "Hello"), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(SCStaticPath), "Not Found", "Parameter 'god' Not Found"})
	test.AssertStringContains(t, "", sendPost(t, 404, "status", "Hello"), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCPathNotFound), "POST URL:/status"})

	test.AssertStringContains(t, "", sendPost(t, 201, "path/static/file/testFile", "Hello"), []string{"\"Created\":\"OK\""})
	test.AssertFileContains(t, "", testFile, []string{"Hello"})

	test.AssertStringContains(t, "", sendGet(t, 200, "status", nil), []string{"\"state\":\"RUNNING\"", "\"executable\":\"TestExe\"", "\"panics\":0"})
	test.AssertStringContains(t, "", sendGet(t, 404, "not-fo", nil), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCPathNotFound), "GET URL:/not-fo"})

	test.AssertEqualString(t, "", "2", sendGet(t, 200, "calc/10/div/5", nil))
	test.AssertStringContains(t, "", sendGet(t, 404, "calc", nil), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCPathNotFound), "GET URL:/calc"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10", nil), []string{"\"Status\":404"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10/div", nil), []string{"\"Status\":404"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10/div/", nil), []string{"\"Status\":404"})
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/10/div/ten", nil), []string{"\"Status\":400", "\"Code\":" + strconv.Itoa(SCParamValidation), "invalid number ten"})
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/five/div/ten", nil), []string{"\"Status\":400", "\"Code\":" + strconv.Itoa(SCParamValidation), "invalid number five"})
	test.AssertStringContains(t, "", sendGet(t, prc, "calc/10/div/0", nil), []string{"\"Status\":" + strconv.Itoa(prc), "\"Code\":" + strconv.Itoa(servermain.SCRuntimeError), "integer divide by zero", "Internal Server Error"})

	stopServer(t)
}

func stopServer(t *testing.T) {
	test.AssertStringContains(t, "", sendGet(t, 200, "stop", nil), []string{"\"state\":\"STOPPING\"", "\"executable\":\"TestExe\"", "\"panics\":1"})
}

func deleteFile(t *testing.T, name string) {
	err := os.Remove(name)
	if err != nil {
		if os.IsNotExist(err) {
			test.Fail(t, "", "Failed. File "+name+" was not created")
		} else {
			test.Fail(t, "", "Failed to remove file "+name+". "+err.Error())
		}
	}
}

func startServer(t *testing.T) {
	if configData == nil {
		err := config.LoadConfigData("webServerTest.json")
		if err != nil {
			test.Fail(t, "Read response Failed", err.Error())
		}
		configData = config.GetConfigDataInstance()
		go RunWithConfig(configData, "TestExe")
		time.Sleep(time.Millisecond * time.Duration(500))
	}
}

func sendGet(t *testing.T, st int, url string, headers map[string]string) string {
	resp, err := http.Get("http://localhost:8080/" + url)
	if err != nil {
		test.Fail(t, "Get Failed", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		test.Fail(t, "Read response Failed", err.Error())
	}
	test.AssertEqualInt(t, "", st, resp.StatusCode)
	if headers != nil {
		for name, value := range headers {
			headerList := resp.Header[name]
			if !strings.Contains(headerList[0], value) {
				test.Fail(t, "", fmt.Sprintf("Header value '%s' does not contain '%s' in element[0]", headerList[0], value))
			}
		}
	}
	return string(body)
}

func sendPost(t *testing.T, st int, url string, postBody string) string {
	resp, err := http.Post("http://localhost:8080/"+url, "application/json", strings.NewReader(postBody))
	if err != nil {
		test.Fail(t, "Post Failed", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		test.Fail(t, "Read response Failed", err.Error())
	}
	test.AssertEqualInt(t, "", st, resp.StatusCode)
	return string(body)
}
