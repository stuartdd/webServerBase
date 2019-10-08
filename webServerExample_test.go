package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
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
var port string

/* site/TestLargeFileRead-002.txt
0: \n
1: b\n
2: c\n
3: \n
4: d
*/
func TestLargeFileRead002_1(t *testing.T) {
	name := "site/TestLargeFileRead-002.txt"
	list := openInitial(name, 5)
	test.AssertStringEqualsUnix(t, "7", list.largeFileHandlerReader(0, 7), "\nb\nc\n\nd")
	test.AssertStringEqualsUnix(t, "6", list.largeFileHandlerReader(0, 6), "\nb\nc\n\nd")
	test.AssertStringEqualsUnix(t, "5", list.largeFileHandlerReader(0, 5), "\nb\nc\n\nd")
	test.AssertStringEqualsUnix(t, "4", list.largeFileHandlerReader(0, 4), "\nb\nc\n\n")
	test.AssertStringEqualsUnix(t, "3", list.largeFileHandlerReader(0, 3), "\nb\nc\n")
	test.AssertStringEqualsUnix(t, "2", list.largeFileHandlerReader(0, 2), "\nb\n")
	test.AssertStringEqualsUnix(t, "1", list.largeFileHandlerReader(0, 1), "\n")
	test.AssertStringEqualsUnix(t, "0", list.largeFileHandlerReader(0, 0), "")
}

/* site/TestLargeFileRead-001.txt
0: a\n
1: b\n
2: c\n
3: \n
4: d\n
*/
func TestLargeFileRead5(t *testing.T) {
	name := "site/TestLargeFileRead-001.txt"
	list := openInitial(name, 20)
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(6, 3), "")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(5, 3), "")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(5, 2), "")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(5, 1), "")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(5, 0), "")
}

func TestLargeFileRead4(t *testing.T) {
	name := "site/TestLargeFileRead-001.txt"
	list := openInitial(name, 5)
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(4, 4), "d\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(4, 3), "d\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(4, 2), "d\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(4, 1), "d\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(4, 0), "")
}

func TestLargeFileRead3(t *testing.T) {
	name := "site/TestLargeFileRead-001.txt"
	list := openInitial(name, 5)
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(3, 99), "\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(3, 3), "\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(3, 2), "\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(3, 1), "\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(3, 0), "")
}

func TestLargeFileRead1(t *testing.T) {
	name := "site/TestLargeFileRead-001.txt"
	list := openInitial(name, 5)
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(1, 99), "b\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(1, 5), "b\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(1, 4), "b\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(1, 3), "b\nc\n\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(1, 2), "b\nc\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(1, 1), "b\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(1, 0), "")
}

func TestLargeFileRead0(t *testing.T) {
	name := "site/TestLargeFileRead-001.txt"
	list := openInitial(name, 5)
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 99), "a\nb\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 7), "a\nb\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 6), "a\nb\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 5), "a\nb\nc\n\nd\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 4), "a\nb\nc\n\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 3), "a\nb\nc\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 2), "a\nb\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 1), "a\n")
	test.AssertStringEqualsUnix(t, "", list.largeFileHandlerReader(0, 0), "")
}

// func TestLargeParse(t *testing.T) {
// 	name := "site/TestLargeFileRead-001.txt"
// 	list := openInitial(name, 5)
// 	checkLFR(t, list)
// 	list = openInitial(name, 1)
// 	checkLFR(t, list)
// 	list = openInitial(name, 2)
// 	checkLFR(t, list)
// 	list = openInitial(name, 3)
// 	checkLFR(t, list)
// 	list = openInitial(name, 4)
// 	checkLFR(t, list)
// 	list = openInitial(name, 6)
// 	checkLFR(t, list)
// 	list = openInitial(name, 7)
// 	checkLFR(t, list)
// 	list = openInitial(name, 8)
// 	checkLFR(t, list)
// 	list = openInitial(name, 9)
// 	checkLFR(t, list)
// 	list = openInitial(name, 10)
// 	checkLFR(t, list)
// 	list = openInitial(name, 11)
// 	checkLFR(t, list)
// 	list = openInitial(name, 101)
// 	checkLFR(t, list)
// }

// func checkLFR(t *testing.T, l *largeFileData) {
// 	test.AssertIntEqual(t, "", l.LineCount, 7)
// 	test.AssertInt64Equal(t, "", l.Offsets[0], 0)
// 	test.AssertInt64Equal(t, "", l.Offsets[1], 1)
// 	test.AssertInt64Equal(t, "", l.Offsets[2], 3)
// 	test.AssertInt64Equal(t, "", l.Offsets[3], 5)
// 	test.AssertInt64Equal(t, "", l.Offsets[4], 6)
// 	test.AssertInt64Equal(t, "", l.Offsets[5], 8)
// 	test.AssertInt64Equal(t, "", l.Offsets[6], 9)
// }

/*
Start server. Do loads of tests. Stop the server...
*/
func TestServer(t *testing.T) {
	startServer(t)
	defer stopServer(t)
	scriptType := "sh"
	if (runtime.GOOS) == "windows" {
		scriptType = "bat"
	}

	test.AssertStringContains(t, "", sendGet(t, 200, "script/list?abc=123&xxx=ABC", headers("json", "")), "ARG 0: echop."+scriptType, "ARG 1: 123-ABC", "TYPE="+scriptType)
	test.AssertStringContains(t, "", sendGet(t, 200, "script/list?abc=123", headers("json", "")), "ARG 0: echop."+scriptType, "ARG 1: 123-${xxx}", "TYPE="+scriptType)
	test.AssertStringContains(t, "", sendGet(t, 404, "script/abc", headers("json", "")), "\"Status\":404,\"Code\":"+strconv.Itoa(servermain.SCScriptNotFound)+"")

	test.AssertStringContains(t, "", sendGet(t, 200, "site/index1.html?Material=LEAD", headers("html", "")), "<title>Soot</title>")
	test.AssertStringContains(t, "", sendGet(t, 404, "site/testfile", headers("json", "")), "\"Status\":404", "\"Code\":"+strconv.Itoa(servermain.SCTemplateNotFound), "Not Found", "/site/testfile")

	/*
		Test static file retrieval
	*/
	test.AssertStringContains(t, "", sendGet(t, 404, "static/testfile", headers("json", "")), "\"Status\":404", "\"Code\":"+strconv.Itoa(servermain.SCContentNotFound), "Not Found", "URL:/static/testfile")
	test.AssertStringContains(t, "", sendGet(t, 200, "static/testfile.json", headers("json", "13")), "{\"test\":true}")
	test.AssertStringContains(t, "", sendGet(t, 200, "static/testfile.xml", headers("xml", "17")), "<test>true</test>")
	test.AssertStringContains(t, "", sendGet(t, 200, "static/arduino.ico", headers("ico", "367958")))
	/*
		Test write file. Mainly to test POST processes
	*/
	testWriteFile1 := configData.GetConfigDataStaticFilePathForOS()["data"] + string(os.PathSeparator) + "createTestFile1.txt"
	testWriteFile2 := configData.GetConfigDataStaticFilePathForOS()["data"] + string(os.PathSeparator) + "createTestFile2.json"
	defer deleteFile(t, testWriteFile1) // Clean up the test data when done!
	defer deleteFile(t, testWriteFile2) // Clean up the test data when done!
	test.AssertStringContains(t, "", sendPost(t, 404, "status", "Hello.txt", headers("json", "")), "\"Status\":404", "\"Code\":"+strconv.Itoa(servermain.SCPathNotFound), "POST URL:/status")
	test.AssertStringContains(t, "", sendPost(t, 201, "path/data/file/createTestFile1", "Hello.txt", headers("json", "16")), "\"Created\":\"OK\"")
	test.AssertStringContains(t, "", sendPost(t, 201, "path/data/file/createTestFile2/ext/json", "Hello.json", headers("json", "16")), "\"Created\":\"OK\"")
	test.AssertStringContains(t, "", sendPost(t, 404, "path/god/file/createTestFile1", "Hello", headers("json", "")), "\"Status\":404", "\"Code\":"+strconv.Itoa(servermain.SCStaticPathNotFound), "Entity:god Not Found")
	test.AssertFileContains(t, "", testWriteFile1, "Hello.txt")
	test.AssertFileContains(t, "", testWriteFile2, "Hello.json")
	/*
		Test GET functions
	*/
	test.AssertStringContains(t, "", sendGet(t, 200, "status", headers("json", "")), "\"State\":\"RUNNING\"", "\"Executable\":\"TestExe\"", "\"Panics\":0")
	test.AssertStringContains(t, "", sendGet(t, 404, "not-fo", headers("json", "")), "\"Status\":404", "\"Code\":"+strconv.Itoa(servermain.SCPathNotFound), "GET URL:/not-fo")
	/*
		Test GET functions with calc
	*/
	test.AssertStringEquals(t, "", sendGet(t, 200, "calc/10/div/5", headers("txt", "1")), "2")
	test.AssertStringEquals(t, "", sendGet(t, 200, "calc/100/div/2", headers("txt", "2")), "50")
	test.AssertStringEquals(t, "", sendGet(t, 200, "calc/100/div/2", headers("txt", "2")), "50")
	test.AssertStringContains(t, "", sendGet(t, 404, "calc", headers("json", "")), "\"Status\":404", "\"Code\":"+strconv.Itoa(servermain.SCPathNotFound), "GET URL:/calc")
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10", headers("json", "")), "\"Status\":404")
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10/div", headers("json", "")), "\"Status\":404")
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10/div/", headers("json", "")), "\"Status\":404")
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/10/div/ten", headers("json", "")), "\"Status\":400", "\"Code\":"+strconv.Itoa(servermain.SCParamValidation), "invalid number ten")
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/five/div/ten", headers("json", "")), "\"Status\":400", "\"Code\":"+strconv.Itoa(servermain.SCParamValidation), "invalid number five")

	test.AssertStringEquals(t, "", sendGet(t, 200, "calc/qube/2", headers("txt", "2")), "16")
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/qube", headers("json", "")), "\"Status\":404")
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/qube/div/10", headers("json", "")), "\"Status\":404")
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/qube/div", headers("json", "")), "\"Status\":400", "\"Code\":"+strconv.Itoa(servermain.SCParamValidation), "invalid number div")
	/*
		Test PANIC responses
	*/
	prc := configData.PanicResponseCode
	test.AssertStringContains(t, "", sendGet(t, prc, "calc/10/div/0", headers("json", "")), "\"Status\":"+strconv.Itoa(prc), "\"Code\":"+strconv.Itoa(servermain.SCRuntimeError), "integer divide by zero", "Internal Server Error")

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
	resp, err := http.Get("http://localhost:" + port + "/" + url)
	if err != nil {
		test.Fail(t, "Get Failed", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		test.Fail(t, "Read response Failed", err.Error())
	}
	test.AssertIntEqual(t, "", resp.StatusCode, st)
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
	resp, err := http.Post("http://localhost:"+port+"/"+url, "application/json", strings.NewReader(postBody))
	if err != nil {
		test.Fail(t, "Post Failed", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		test.Fail(t, "Read response Failed", err.Error())
	}
	test.AssertIntEqual(t, "", resp.StatusCode, st)
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
	test.AssertStringContains(t, "", sendGet(t, 200, "stop", nil), "\"State\":\"STOPPING\"", "\"Executable\":\"TestExe\"", "\"Panics\":1")
	testLog.LogDebug("SHUT DOWN STARTED")
}

func startServer(t *testing.T) {
	if configData == nil {
		err := config.LoadConfigData("webServerExample.json")
		if err != nil {
			test.Fail(t, "Read response Failed", err.Error())
		}
		configData = config.GetConfigDataInstance()
		port = fmt.Sprintf("%d", configData.Port)
		logging.CreateLogWithFilenameAndAppID(configData.DefaultLogFileName, "TEST:"+strconv.Itoa(configData.Port), 1, configData.LoggerLevels)
		testLog = logging.CreateTestLogger("CONTROL")
		go RunWithConfig(configData, "TestExe")
		testLog.LogDebug("SERVER STARTING")
		time.Sleep(time.Millisecond * time.Duration(500))
	}
}
