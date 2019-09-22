package config

import (
	"testing"

	"github.com/stuartdd/webServerBase/test"
)

func TestConfigLoad(t *testing.T) {
	test.AssertErrorIsNil(t, "Should not return an error", LoadConfigData("../webServerExample.json"))
}
func TestConfigLoadFileNotFound(t *testing.T) {
	defer test.AssertPanicAndRecover(t, "Failed to read config data file")
	test.AssertErrorIsNotExist(t, "File not found", LoadConfigData("debug.test.json"))
}
