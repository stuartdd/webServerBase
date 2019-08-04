package dto

import (
	"testing"
)

/*
TestCreate test lookup creation
*/
func TestCreate(*testing.T) {
	GetMappingElementInstance()
	AddPathMappingElement("a/b/c")
}
