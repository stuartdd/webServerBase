package servermain

import (
	"bytes"
	"html/template"
	"testing"
	"webServerBase/logging"
	"webServerBase/test"
)

type Data struct {
	Count    int
	Material string
}

func TestLoadTemplatesSite(t *testing.T) {
	templ, err1 := LoadTemplates("../site")
	test.FailIfError(t, "", err1)
	data1 := Data{
		Count:    2,
		Material: "Metal",
	}
	data2 := Data{
		Count:    4,
		Material: "Zinc",
	}

	txt1, err2 := templ.Execute("simple1.html", data1)
	test.FailIfError(t, "", err2)
	logging.CreateTestLogger("data1").LogInfof("RESULT:[%s]", txt1)
	test.AssertEqualString(t, "Template result data1", "Count is 2 Material is Metal", txt1)

	txt2, err3 := templ.Execute("simple2.html", data2)
	test.FailIfError(t, "", err3)
	logging.CreateTestLogger("data1").LogInfof("RESULT:[%s]", txt2)
	test.AssertEqualString(t, "Template result data2", "Material is Zinc Count is 4", txt2)
}

func TestStaticTemplates(t *testing.T) {
	tmpl, err := template.New("helloWorld").Parse("Count is {{ .Count }} Material is {{ .Material }}")
	test.FailIfError(t, "", err)
	buf := new(bytes.Buffer)
	data1 := Data{
		Count:    7,
		Material: "Silk",
	}
	data2 := Data{
		Count:    9,
		Material: "Wool",
	}
	err1 := tmpl.Execute(buf, data1)
	test.FailIfError(t, "", err1)
	test.AssertEqualString(t, "Template result data1", "Count is 7 Material is Silk", buf.String())

	buf.Reset()
	err2 := tmpl.Execute(buf, data2)
	test.FailIfError(t, "", err2)
	test.AssertEqualString(t, "Template result data2", "Count is 9 Material is Wool", buf.String())
}

func TestFileTemplateNotFound(t *testing.T) {
	_, err := template.ParseFiles("../site/simple1.template.html.x")
	test.FailIfNilError(t, "File should NOT be found", err)
}

func TestFileTemplates(t *testing.T) {
	tmpl, err := template.ParseFiles("../site/simple1.template.html")
	test.FailIfError(t, "", err)
	buf := new(bytes.Buffer)
	data1 := Data{
		Count:    7,
		Material: "Silk",
	}
	data2 := Data{
		Count:    9,
		Material: "Wool",
	}
	err1 := tmpl.Execute(buf, data1)
	test.FailIfError(t, "", err1)
	test.AssertEqualString(t, "Template result data1", "Count is 7 Material is Silk", buf.String())

	buf.Reset()
	err2 := tmpl.Execute(buf, data2)
	test.FailIfError(t, "", err2)
	test.AssertEqualString(t, "Template result data2", "Count is 9 Material is Wool", buf.String())
}
