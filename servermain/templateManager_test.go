package servermain

import (
	"testing"
	"webServerBase/logging"
)

type Data struct {
	Count    int
	Material string
}

func TestLoadTemplatesSite(t *testing.T) {
	logger := logging.CreateTestLogger("templates")
	templ, _ := LoadTemplates("../site")
	a := Data{
		Count:    7,
		Material: "Silk",
	}

	txt, err1 := templ.Execute("index1.html", a)
	logger.LogInfof("%s, %s", txt, err1.Error())
	//	test.AssertEqualString(t, "fred", "Template error: template: \"index1.html\" is an incomplete or empty template", err1.Error())
}
