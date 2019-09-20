package config

import (
	"testing"
	"webServerBase/test"
)

func TestConfigLoad(t *testing.T) {
	test.AssertErrorIsNil(t, "Should not return an error", LoadConfigData("../webServerExample.json"))
}
func TestConfigLoadFileNotFound(t *testing.T) {
	test.AssertPanicAndRecover(t, "xxx")
	test.AssertErrorIsNotExist(t, "File not found", LoadConfigData("debug.test.json"))
}

func TestConfigData(t *testing.T) {
	test.AssertErrorIsNil(t, "Should not return an error", LoadConfigData("../debug.test.json"))
	config := GetConfigDataInstance()
	test.AssertStringEquals(t, "", "../debug.test.json", config.ConfigName)
	test.AssertIntEqual(t, "", 8080, config.Port)
	test.AssertStringEquals(t, "", "utf-8", config.ContentTypeCharset)
}
