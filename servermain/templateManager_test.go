package servermain

import (
	"bytes"
	"html/template"
	"testing"
	"webServerBase/test"
)

type Data struct {
	Count    int
	Material string
}

func TestLoadTemplatesFromSite(t *testing.T) {
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
	test.AssertEqualString(t, "Template result data1", "Count is 2 Material is Metal", txt1)

	txt2, err3 := templ.Execute("simple2.html", data2)
	test.FailIfError(t, "", err3)
	test.AssertEqualString(t, "Template result data2", "Material is Zinc Count is 4", txt2)

	_, err4 := templ.Execute("simple3.html", data2)
	test.FailIfNilErrorAndContains(t, "", "can't evaluate field Type", err4)

	txt4, err5 := templ.Execute("simple3.html", map[string]string{"Type": "Fire", "Count": "SIX"})
	test.FailIfError(t, "", err5)
	test.AssertEqualString(t, "Template result mapped", "Material is Fire Count is SIX", txt4)

	mapData := make(map[string]string)
	mapData["Type"] = "Water"
	mapData["Count"] = "4"
	mapData["Extra"] = "More"
	txt5, err6 := templ.Execute("simple3.html", mapData)
	test.FailIfError(t, "", err6)
	test.AssertEqualString(t, "Template result mapped", "Material is Water Count is 4", txt5)
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
