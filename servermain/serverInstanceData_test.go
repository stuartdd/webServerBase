package servermain

import (
	"testing"

	"github.com/stuartdd/webServerBase/test"
)

func TestOsScriptsConfigNoPath(t *testing.T) {
	server := NewServerInstanceData("ServerName", "utf-8")
	defer test.AssertPanicAndRecover(t, "A path is required")
	server.SetOsScriptsData("", nil)
}
func TestOsScriptsConfigNoData(t *testing.T) {
	server := NewServerInstanceData("ServerName", "utf-8")
	defer test.AssertPanicAndRecover(t, "data cannot be empty")
	server.SetOsScriptsData("/", nil)
}

func TestOsScriptsConfigBadPath(t *testing.T) {
	server := NewServerInstanceData("ServerName", "utf-8")
	defer test.AssertPanicAndRecover(t, "could not be found")
	server.SetOsScriptsData("/abc123", make(map[string][]string))
}

func TestOsScriptsConfigEmptyMap(t *testing.T) {
	server := NewServerInstanceData("ServerName", "utf-8")
	defer test.AssertPanicAndRecover(t, "data cannot be empty")
	server.SetOsScriptsData("/", make(map[string][]string))
}

func TestOsScriptsConfigEmptyMapData(t *testing.T) {
	server := NewServerInstanceData("ServerName", "utf-8")
	m := make(map[string][]string)
	m["empty"] = []string{}
	defer test.AssertPanicAndRecover(t, "data for script [empty]")
	server.SetOsScriptsData("/", m)
}
