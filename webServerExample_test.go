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

	"github.com/stuartdd/webServerBase/config"
	"github.com/stuartdd/webServerBase/logging"
	"github.com/stuartdd/webServerBase/servermain"
	"github.com/stuartdd/webServerBase/test"
)

var configData *config.Data
var testLog = logging.CreateTestLogger("Test-Logger")

/*
Start server. Do loads of tests. Stop the server...
*/
func TestServer(t *testing.T) {
	startServer(t)
	test.AssertStringContains(t, "", sendGet(t, 200, "site/index1.html?Material=LEAD", headers("html", "")), []string{"<title>GOLD</title>"})
	test.AssertStringContains(t, "", sendGet(t, 404, "site/testfile", headers("json", "")), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCTemplateNotFound), "Not Found", "/site/testfile"})

	/*
		Test static file retrieval
	*/
	test.AssertStringContains(t, "", sendGet(t, 404, "static/testfile", headers("json", "")), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCContentNotFound), "Not Found", "URL:/static/testfile"})
	test.AssertStringContains(t, "", sendGet(t, 200, "static/testfile.json", headers("json", "13")), []string{"{\"test\":true}"})
	test.AssertStringContains(t, "", sendGet(t, 200, "static/testfile.xml", headers("xml", "17")), []string{"<test>true</test>"})
	test.AssertStringContains(t, "", sendGet(t, 200, "static/arduino.ico", headers("ico", "367958")), []string{})
	/*
		Test write file. Mainly to test POST processes
	*/
	testWriteFile1 := configData.GetConfigDataStaticFilePathForOS()["data"] + string(os.PathSeparator) + "createTestFile1.txt"
	testWriteFile2 := configData.GetConfigDataStaticFilePathForOS()["data"] + string(os.PathSeparator) + "createTestFile2.json"
	defer deleteFile(t, testWriteFile1) // Clean up the test data when done!
	defer deleteFile(t, testWriteFile2) // Clean up the test data when done!
	test.AssertStringContains(t, "", sendPost(t, 404, "status", "Hello.txt", headers("json", "")), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCPathNotFound), "POST URL:/status"})
	test.AssertStringContains(t, "", sendPost(t, 201, "path/data/file/createTestFile1", "Hello.txt", headers("json", "16")), []string{"\"Created\":\"OK\""})
	test.AssertStringContains(t, "", sendPost(t, 201, "path/data/file/createTestFile2/ext/json", "Hello.json", headers("json", "16")), []string{"\"Created\":\"OK\""})
	test.AssertStringContains(t, "", sendPost(t, 404, "path/god/file/createTestFile1", "Hello", headers("json", "")), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCStaticPathNotFound), "Entity:god Not Found"})
	test.AssertFileContains(t, "", testWriteFile1, []string{"Hello.txt"})
	test.AssertFileContains(t, "", testWriteFile2, []string{"Hello.json"})
	/*
		Test GET functions
	*/
	test.AssertStringContains(t, "", sendGet(t, 200, "status", headers("json", "")), []string{"\"State\":\"RUNNING\"", "\"Executable\":\"TestExe\"", "\"Panics\":0"})
	test.AssertStringContains(t, "", sendGet(t, 404, "not-fo", headers("json", "")), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCPathNotFound), "GET URL:/not-fo"})
	/*
		Test GET functions with calc
	*/
	test.AssertStringEquals(t, "", "2", sendGet(t, 200, "calc/10/div/5", headers("txt", "1")))
	test.AssertStringEquals(t, "", "50", sendGet(t, 200, "calc/100/div/2", headers("txt", "2")))
	test.AssertStringEquals(t, "", "50", sendGet(t, 200, "calc/100/div/2", headers("txt", "2")))
	test.AssertStringContains(t, "", sendGet(t, 404, "calc", headers("json", "")), []string{"\"Status\":404", "\"Code\":" + strconv.Itoa(servermain.SCPathNotFound), "GET URL:/calc"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10", headers("json", "")), []string{"\"Status\":404"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10/div", headers("json", "")), []string{"\"Status\":404"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10/div/", headers("json", "")), []string{"\"Status\":404"})
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/10/div/ten", headers("json", "")), []string{"\"Status\":400", "\"Code\":" + strconv.Itoa(SCParamValidation), "invalid number ten"})
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/five/div/ten", headers("json", "")), []string{"\"Status\":400", "\"Code\":" + strconv.Itoa(SCParamValidation), "invalid number five"})

	test.AssertStringEquals(t, "", "16", sendGet(t, 200, "calc/qube/2", headers("txt", "2")))
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/qube", headers("json", "")), []string{"\"Status\":404"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/qube/div/10", headers("json", "")), []string{"\"Status\":404"})
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/qube/div", headers("json", "")), []string{"\"Status\":400", "\"Code\":" + strconv.Itoa(SCParamValidation), "invalid number div"})
	/*
		Test PANIC responses
	*/
	prc := configData.PanicResponseCode
	test.AssertStringContains(t, "", sendGet(t, prc, "calc/10/div/0", headers("json", "")), []string{"\"Status\":" + strconv.Itoa(prc), "\"Code\":" + strconv.Itoa(servermain.SCRuntimeError), "integer divide by zero", "Internal Server Error"})

	stopServer(t)
}

func headers(ct string, cl string) map[string]string {
	headers := make(map[string]string)
	if ct != "" {
		headers[servermain.ContentTypeName] = servermain.LookupContentType(ct)
	}
	if cl != "" {
		headers[servermain.ContentLengthName] = cl
	}
	return headers
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
	test.AssertIntEqual(t, "", st, resp.StatusCode)
	if headers != nil {
		for name, value := range headers {
			if value != "" {
				headerList := resp.Header[name]
				if !strings.Contains(headerList[0], value) {
					test.Fail(t, "", fmt.Sprintf("GET: Header value '%s' does not contain '%s' in element[0]", headerList[0], value))
				}
			}
		}
	}
	return string(body)
}

func sendPost(t *testing.T, st int, url string, postBody string, headers map[string]string) string {
	resp, err := http.Post("http://localhost:8080/"+url, "application/json", strings.NewReader(postBody))
	if err != nil {
		test.Fail(t, "Post Failed", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		test.Fail(t, "Read response Failed", err.Error())
	}
	test.AssertIntEqual(t, "", st, resp.StatusCode)
	if headers != nil {
		for name, value := range headers {
			if value != "" {
				headerList := resp.Header[name]
				if !strings.Contains(headerList[0], value) {
					test.Fail(t, "", fmt.Sprintf("POST: Header value '%s' does not contain '%s' in element[0]", headerList[0], value))
				}
			}
		}
	}
	return string(body)
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

func stopServer(t *testing.T) {
	test.AssertStringContains(t, "", sendGet(t, 200, "stop", nil), []string{"\"State\":\"STOPPING\"", "\"Executable\":\"TestExe\"", "\"Panics\":1"})
}

func startServer(t *testing.T) {
	if configData == nil {
		err := config.LoadConfigData("webServerExample.json")
		if err != nil {
			test.Fail(t, "Read response Failed", err.Error())
		}
		configData = config.GetConfigDataInstance()
		go RunWithConfig(configData, "TestExe")
		time.Sleep(time.Millisecond * time.Duration(500))
	}
}
