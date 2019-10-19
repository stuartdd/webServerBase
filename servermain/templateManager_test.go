package servermain

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/stuartdd/webServerBase/logging"
	"github.com/stuartdd/webServerBase/test"
)

type Data struct {
	Count    int
	Material string
}
type DataError struct {
	count  int
	MadeOf string
}

var templ *Templates
var err1 error

func TestLoadTemplatesNested(t *testing.T) {
	if templ == nil {
		logging.CreateTestLogger("TestTemplates")
		templ, err1 = loadTemplates("../example/site/templates")
		test.AssertErrorIsNil(t, "", err1)
	}
	mapData := make(map[string]string)
	mapData["Type"] = "Water"
	mapData["Count"] = "4"
	mapData["Extra"] = "More"
	mapData["Material"] = "Silk"
	test.AssertBoolFalse(t, "", templ.HasTemplate("import1.html"))
	test.AssertBoolFalse(t, "", templ.HasTemplate("part1.html"))
	test.AssertBoolFalse(t, "", templ.HasTemplate("part2.html"))
	test.AssertBoolTrue(t, "", templ.HasTemplate("composite1.html"))
	test.AssertBoolTrue(t, "", templ.HasTemplate("simple1.html"))
	test.AssertBoolTrue(t, "", templ.HasTemplate("simple2.html"))
	txt1 := templ.executeString("composite1.html", mapData)
	test.AssertStringEquals(t, "Template result data1", txt1, "Imp 1.0 - Count is 4 Material is More Silk --PART 1 Silk-- --PART 2 is 4 of Silk -- --PART 3 wants More--")

}

func TestFileGroupTemplates(t *testing.T) {
	tmpl, err := template.ParseFiles("../example/site/templates/import1.html", "../example/site/templates/part1.html", "../example/site/templates/part2.html")
	test.AssertErrorIsNil(t, "", err)
	buf := new(bytes.Buffer)
	mapData := make(map[string]string)
	mapData["Type"] = "Water"
	mapData["Count"] = "4"
	mapData["Extra"] = "More"
	mapData["Material"] = "Silk"
	err1 := tmpl.Execute(buf, mapData)
	test.AssertErrorIsNil(t, "", err1)
	test.AssertStringEquals(t, "Template result data1", buf.String(), "Imp 1.0 - Count is 4 Material is More Silk --PART 1 Silk-- --PART 2 is 4 of Silk -- --PART 3 wants More--")

}
func TestWriterNotFoundPanic(t *testing.T) {
	if templ == nil {
		templ, err1 = loadTemplates("../example/site/templates")
	}
	test.AssertErrorIsNil(t, "", err1)
	defer test.AssertPanicAndRecover(t, "Template 'simple4.html' not found")
	templ.executeString("simple4.html", make(map[string]string))
}

func TestWriterPanic(t *testing.T) {
	if templ == nil {
		templ, err1 = loadTemplates("../site")
	}
	test.AssertErrorIsNil(t, "", err1)
	data1 := DataError{
		count:  2,
		MadeOf: "Metal",
	}
	defer test.AssertPanicAndRecover(t, "can't evaluate field Count")
	templ.executeString("simple1.html", data1)
}

func TestLoadTemplatesFromPath(t *testing.T) {
	if templ == nil {
		templ, err1 = loadTemplates("../example/site/templates")
	}

	test.AssertErrorIsNil(t, "", err1)
	data1 := Data{
		Count:    2,
		Material: "Metal",
	}
	data2 := Data{
		Count:    4,
		Material: "Zinc",
	}
	test.AssertBoolTrue(t, "", templ.HasTemplate("simple1.html"))
	txt1 := templ.executeString("simple1.html", data1)
	test.AssertStringEquals(t, "Template result data1", txt1, "S1 - Count is 2 Material is Metal")

	test.AssertBoolTrue(t, "", templ.HasTemplate("simple2.html"))
	txt2 := templ.executeString("simple2.html", data2)
	test.AssertStringEquals(t, "Template result data2", txt2, "S2 - Material is Zinc Count is 4")

	test.AssertBoolFalse(t, "simpleXXX.html", templ.HasTemplate("simpleXXX.html"))

	test.AssertBoolTrue(t, "", templ.HasTemplate("simple3.html"))
	txt4 := templ.executeString("simple3.html", map[string]string{"Type": "Fire", "Count": "SIX"})
	test.AssertStringEquals(t, "S3 - Template result mapped", txt4, "S3 - Material is Fire Count is SIX")

	mapData := make(map[string]string)
	mapData["Type"] = "Water"
	mapData["Count"] = "4"
	mapData["Extra"] = "More"
	txt5 := templ.executeString("simple3.html", mapData)
	test.AssertStringEquals(t, "S3 - Template result mapped", txt5, "S3 - Material is Water Count is 4")
}

func TestStaticTemplates(t *testing.T) {
	tmpl, err := template.New("helloWorld").Parse("Count is {{ .Count }} Material is {{ .Material }}")
	test.AssertErrorIsNil(t, "", err)
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
	test.AssertErrorIsNil(t, "", err1)
	test.AssertStringEquals(t, "Template result data1", buf.String(), "Count is 7 Material is Silk")

	buf.Reset()
	err2 := tmpl.Execute(buf, data2)
	test.AssertErrorIsNil(t, "", err2)
	test.AssertStringEquals(t, "Template result data2", buf.String(), "Count is 9 Material is Wool")
}

func TestFileTemplateNotFound(t *testing.T) {
	_, err := template.ParseFiles("../example/site/templates/simple1.template.html.x")
	test.AssertError(t, "File should NOT be found", err)
}

func TestFileTemplates(t *testing.T) {
	tmpl, err := template.ParseFiles("../example/site/templates/simple1.template.html")
	test.AssertErrorIsNil(t, "", err)
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
	test.AssertErrorIsNil(t, "", err1)
	test.AssertStringEquals(t, "Template result data1", buf.String(), "S1 - Count is 7 Material is Silk")

	buf.Reset()
	err2 := tmpl.Execute(buf, data2)
	test.AssertErrorIsNil(t, "", err2)
	test.AssertStringEquals(t, "Template result data2", buf.String(), "S1 - Count is 9 Material is Wool")
}
