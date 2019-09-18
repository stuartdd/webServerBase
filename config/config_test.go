package config

import (
	"testing"
	"webServerBase/test"
)

func TestConfigLoad(t *testing.T) {
	test.AssertErrorIsNil(t, "Should not return an error", LoadConfigData("../webServerTest.json"))
	test.AssertErrorIsNotExist(t, "File not found", LoadConfigData("webServerTest.json"))
}

func TestConfigData(t *testing.T) {
	test.AssertErrorIsNil(t, "Should not return an error", LoadConfigData("../webServerTest.json"))
	config := GetConfigDataInstance()
	test.AssertStringEquals(t, "", "../webServerTest.json", config.ConfigName)
	test.AssertIntEqual(t, "", 8080, config.Port)
	test.AssertStringEquals(t, "", "utf-8", config.ContentTypeCharset)
}
