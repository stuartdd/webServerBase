package config

import (
	"runtime"
	"testing"

	"github.com/stuartdd/webServerBase/test"
)

func TestConfigLoad(t *testing.T) {
	test.AssertErrorIsNil(t, "Should not return an error", LoadConfigData("../webServerExample.json"))

	cfg := GetConfigDataInstance()
	if runtime.GOOS == "windows" {
		test.AssertStringEquals(t, "", cfg.GetScriptDataForOS().Path, "site\\scripts\\")
		test.AssertStringEquals(t, "", cfg.GetScriptDataForOS().Data["list"][0], "cmd")
	} else {
		test.AssertStringEquals(t, "", cfg.GetScriptDataForOS().Path, "site/scripts/")
		test.AssertStringEquals(t, "", cfg.GetScriptDataForOS().Data["list"][0], "sh")
	}
}
func TestConfigLoadFileNotFound(t *testing.T) {
	defer test.AssertPanicAndRecover(t, "Failed to read config data file")
	test.AssertErrorIsNotExist(t, "File not found", LoadConfigData("debug.test.json"))
}
