package main

import (
	"webServerBase/logging"
	"testing"
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
	test.AssertStringContains(t, "", sendGet(t, 200, "status"), []string{"\"state\":\"RUNNING\"", "\"executable\":\"TestExe\"", "\"panics\":0"})
	test.AssertStringContains(t, "", sendGet(t, 404, "not-fo"), []string{"\"Code\":404"})
	stopServer(t)
}

func stopServer(t *testing.T) {
	test.AssertStringContains(t, "", sendGet(t, 200, "stop"), []string{"\"state\":\"STOPPING\"", "\"executable\":\"TestExe\"", "\"panics\":0"})
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

