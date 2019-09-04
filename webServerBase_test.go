package main

import (
	"webServerBase/logging"
	"testing"
	"strconv"
	"strings"
	"net/http"
	"time"
	"io/ioutil"
	"webServerBase/test"
	"webServerBase/config"

)

var configData *config.Data
var testLog = logging.CreateTestLogger("Test-Logger")

func TestGetStatus(t *testing.T) {
	
	startServer(t)
	prc := configData.PanicResponseCode

	test.AssertStringContains(t, "", sendPost(t, 201, "path/static/file/testFile", "Hello"),[]string{"\"Created\":\"OK\""})
	test.AssertStringContains(t, "", sendPost(t, 404, "path/god/file/testFile", "Hello"),  []string{"\"Code\":404"})

	test.AssertStringContains(t, "", sendGet(t, 200, "status"), []string{"\"state\":\"RUNNING\"", "\"executable\":\"TestExe\"", "\"panics\":0"})
	test.AssertStringContains(t, "", sendGet(t, 404, "not-fo"), []string{"\"Code\":404"})

	test.AssertEqualString(t, "", "2", sendGet(t, 200, "calc/10/div/5"))
	test.AssertStringContains(t, "", sendGet(t, 404, "calc"), []string{"\"Code\":404"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10"), []string{"\"Code\":404"})
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10/div"), []string{"\"Code\":404"})		
	test.AssertStringContains(t, "", sendGet(t, 404, "calc/10/div/"), []string{"\"Code\":404"})		
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/10/div/ten"), []string{"\"Code\":400","invalid number ten"})		
	test.AssertStringContains(t, "", sendGet(t, 400, "calc/five/div/ten"), []string{"\"Code\":400","invalid number five"})		
	test.AssertStringContains(t, "", sendGet(t, prc, "calc/10/div/0"), []string{"\"Code\":"+strconv.Itoa(prc), "integer divide by zero", "MESSAGE:runtime error"})

	
	stopServer(t)
}

func stopServer(t *testing.T) {
	test.AssertStringContains(t, "", sendGet(t, 200, "stop"), []string{"\"state\":\"STOPPING\"", "\"executable\":\"TestExe\"", "\"panics\":1"})
}

func startServer(t *testing.T) {
	if (configData == nil) {
		err := config.LoadConfigData("webServerTest.json")
		if err != nil {
			test.Fail(t, "Read response Failed", err.Error())
		}
		configData = config.GetConfigDataInstance()
		go RunWithConfig(configData, "TestExe")
		time.Sleep(time.Millisecond * time.Duration(500))	
	}
}

func sendGet(t *testing.T, st int, url string) string {
	resp, err := http.Get("http://localhost:8080/"+url)
	if err != nil {
		test.Fail(t, "Get Failed", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		test.Fail(t, "Read response Failed", err.Error())
	}
	test.AssertEqualInt(t, "", st, resp.StatusCode)
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
