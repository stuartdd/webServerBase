package config

import (
	"testing"
	"webServerBase/test"
)

func TestConfigLoad(t *testing.T) {
	test.AssertErrorIsNil(t, "Should not return an error", LoadConfigData("../webServerTest.json"))
	test.AssertErrorTextContains(t, "File not found", LoadConfigData("webServerTest.json"), "The system cannot find the file specified")
}

func TestConfigDat(t *testing.T) {
	test.AssertErrorIsNil(t, "Should not return an error", LoadConfigData("../webServerTest.json"))
	config := GetConfigDataInstance()
	test.AssertEqualString(t, "", "../webServerTest.json", config.ConfigName)
	test.AssertEqualInt(t, "", 8080, config.Port)
}
