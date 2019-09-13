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
type DataError struct {
	count  int
	MadeOf string
}

var templ *Templates
var err1 error

func TestLoadTemplatesNested(t *testing.T) {
	if templ == nil {
		logging.CreateTestLogger("TestTemplates")
		templ, err1 = LoadTemplates("../site")
		test.AssertErrorIsNil(t, "", err1)
	}
	mapData := make(map[string]string)
	mapData["Type"] = "Water"
	mapData["Count"] = "4"
	mapData["Extra"] = "More"
	mapData["Material"] = "Silk"
	test.AssertFalse(t, "", templ.HasTemplate("import1.html"))
	test.AssertFalse(t, "", templ.HasTemplate("part1.html"))
	test.AssertFalse(t, "", templ.HasTemplate("part2.html"))
	test.AssertTrue(t, "", templ.HasTemplate("composite1.html"))
	test.AssertTrue(t, "", templ.HasTemplate("simple1.html"))
	test.AssertTrue(t, "", templ.HasTemplate("simple2.html"))
	txt1 := templ.ExecuteString("composite1.html", mapData)
	test.AssertEqualString(t, "Template result data1", "Imp 1.0 - Count is 4 Material is More Silk --PART 1 Silk-- --PART 2 is 4 of Silk -- --PART 3 wants More--", txt1)

}

func TestFileGroupTemplates(t *testing.T) {
	tmpl, err := template.ParseFiles("../site/import1.html", "../site/part1.html", "../site/part2.html")
	test.AssertErrorIsNil(t, "", err)
	buf := new(bytes.Buffer)
	mapData := make(map[string]string)
	mapData["Type"] = "Water"
	mapData["Count"] = "4"
	mapData["Extra"] = "More"
	mapData["Material"] = "Silk"
	err1 := tmpl.Execute(buf, mapData)
	test.AssertErrorIsNil(t, "", err1)
	test.AssertEqualString(t, "Template result data1", "Imp 1.0 - Count is 4 Material is More Silk --PART 1 Silk-- --PART 2 is 4 of Silk -- --PART 3 wants More--", buf.String())

}
func TestWriterNotFoundPanic(t *testing.T) {
	if templ == nil {
		templ, err1 = LoadTemplates("../site")
	}
	test.AssertErrorIsNil(t, "", err1)
	defer test.AssertPanicThrown(t, "Template 'simple4.html' not found")
	templ.ExecuteString("simple4.html", make(map[string]string))
}

func TestWriterPanic(t *testing.T) {
	if templ == nil {
		templ, err1 = LoadTemplates("../site")
	}
	test.AssertErrorIsNil(t, "", err1)
	data1 := DataError{
		count:  2,
		MadeOf: "Metal",
	}
	defer test.AssertPanicThrown(t, "can't evaluate field Count")
	templ.ExecuteString("simple1.html", data1)
}

func TestLoadTemplatesFromPath(t *testing.T) {
	if templ == nil {
		templ, err1 = LoadTemplates("../site")
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
	test.AssertTrue(t, "", templ.HasTemplate("simple1.html"))
	txt1 := templ.ExecuteString("simple1.html", data1)
	test.AssertEqualString(t, "Template result data1", "S1 - Count is 2 Material is Metal", txt1)

	test.AssertTrue(t, "", templ.HasTemplate("simple2.html"))
	txt2 := templ.ExecuteString("simple2.html", data2)
	test.AssertEqualString(t, "Template result data2", "S2 - Material is Zinc Count is 4", txt2)

	test.AssertFalse(t, "simpleXXX.html", templ.HasTemplate("simpleXXX.html"))

	test.AssertTrue(t, "", templ.HasTemplate("simple3.html"))
	txt4 := templ.ExecuteString("simple3.html", map[string]string{"Type": "Fire", "Count": "SIX"})
	test.AssertEqualString(t, "S3 - Template result mapped", "S3 - Material is Fire Count is SIX", txt4)

	mapData := make(map[string]string)
	mapData["Type"] = "Water"
	mapData["Count"] = "4"
	mapData["Extra"] = "More"
	txt5 := templ.ExecuteString("simple3.html", mapData)
	test.AssertEqualString(t, "S3 - Template result mapped", "S3 - Material is Water Count is 4", txt5)
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
	test.AssertEqualString(t, "Template result data1", "Count is 7 Material is Silk", buf.String())

	buf.Reset()
	err2 := tmpl.Execute(buf, data2)
	test.AssertErrorIsNil(t, "", err2)
	test.AssertEqualString(t, "Template result data2", "Count is 9 Material is Wool", buf.String())
}

func TestFileTemplateNotFound(t *testing.T) {
	_, err := template.ParseFiles("../site/simple1.template.html.x")
	test.AssertIsError(t, "File should NOT be found", err)
}

func TestFileTemplates(t *testing.T) {
	tmpl, err := template.ParseFiles("../site/simple1.template.html")
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
	test.AssertEqualString(t, "Template result data1", "S1 - Count is 7 Material is Silk", buf.String())

	buf.Reset()
	err2 := tmpl.Execute(buf, data2)
	test.AssertErrorIsNil(t, "", err2)
	test.AssertEqualString(t, "Template result data2", "S1 - Count is 9 Material is Wool", buf.String())
}
