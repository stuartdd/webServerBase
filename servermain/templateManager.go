package servermain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/stuartdd/webServerBase/logging"
)

type templateGroup struct {
	Name      string
	Templates []string
}

var groupList []templateGroup

type templateData struct {
	name         string
	file         string
	dataProvider func(*http.Request, string, interface{})
	template     *template.Template
}

/*
Templates list of templates by ID
*/
type Templates struct {
	templates map[string]*templateData
}

var logger *logging.LoggerDataReference

/*
ReasonableTemplateFileHandler - Response handler for basic template processing
*/
func ReasonableTemplateFileHandler(request *http.Request, response *Response) {
	h := NewRequestHandlerHelper(request, response)
	server := h.GetServer()
	name := h.GetNamedURLPart("site", "")
	if server.HasTemplate(name) {
		ww := h.GetResponseWriter()
		contentType := LookupContentType(name)
		if (contentType != "") && (ww.Header()[ContentTypeName] == nil) {
			ww.Header()[ContentTypeName] = []string{contentType + "; charset=" + server.contentTypeCharset}
		}
		m := h.GetQueries()
		server.TemplateWithWriter(ww, name, request, m)
		response.Close()
		if logging.IsAccess() {
			response.GetWrappedServer().GetServerLogger().LogAccessf("<<< STATUS=%d: CODE=%d: RESP-FROM-FILE=%s: TYPE=%s", response.GetCode(), response.GetSubCode(), name, contentType)
			response.GetWrappedServer().logHeaderMap(response.GetHeaders(), "<-<")
		}
	} else {
		response.SetError404(h.GetURL()+" "+server.ListTemplateNames("|"), SCTemplateNotFound)
	}
}

/*
LoadTemplates - Load the templates given the template paths. For example
	"templates\\"

Templates should be named *.template.* in order to be parsed!

The resulting template name is the name with '.template' removed
*/
func loadTemplates(templatePath string) (*Templates, error) {
	logger = logging.NewLogger("Template")
	templateList := &Templates{
		templates: make(map[string]*templateData),
	}
	walkError := filepath.Walk(templatePath, func(path string, info os.FileInfo, errIn error) error {
		if strings.Contains(path, ".template.groups.json") {
			return loadGroupOfTemplates(templatePath, path, templateList)
		}
		if strings.Contains(path, ".template.") {
			return loadSingletemplate(path, templateList)
		}
		return errIn
	})
	if walkError != nil {
		return nil, walkError
	}
	return templateList, nil
}

func loadGroupOfTemplates(templatePath string, groupFile string, templateList *Templates) error {
	fullPath, filePathErr := filepath.Abs(groupFile)
	if filePathErr == nil {
		if filePathErr != nil {
			return filePathErr
		}
		err := loadJSONGroupList(fullPath, &groupList)
		if err != nil {
			return err
		}

		for _, group := range groupList {

			for index := range group.Templates {
				pathTotemplate, err := filepath.Abs(path.Join(templatePath, group.Templates[index]))
				if err != nil {
					return err
				}
				group.Templates[index] = pathTotemplate
			}
			tmpl, err := template.ParseFiles(group.Templates...)
			if err != nil {
				return err
			}
			templateList.templates[group.Name] = &templateData{
				name:     group.Name,
				file:     fullPath,
				template: tmpl,
			}
			logger.LogDebugf("Loading: Template Group defined in file:%s", fullPath)
		}
	}
	return filePathErr
}

func loadJSONGroupList(fileName string, obj interface{}) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, obj)
	if err != nil {
		return err
	}
	return nil
}

func loadSingletemplate(path string, templateList *Templates) error {
	fullPath, filePathErr := filepath.Abs(path)
	if filePathErr == nil {
		_, tname := filepath.Split(fullPath)
		fname := strings.Replace(tname, ".template", "", 1)
		var tmpl *template.Template
		var err error
		if fname == "import1.html" {
			tmpl, err = template.ParseFiles(fullPath)
		} else {
			tmpl, err = template.ParseFiles(fullPath)
		}
		if err != nil {
			return err
		}
		templateList.templates[fname] = &templateData{
			name:     fname,
			file:     fullPath,
			template: tmpl,
		}
		logger.LogDebugf("Loading: FILE:%s NAME:%s PATH:%s", tname, fname, fullPath)
	}
	return filePathErr

}

/*
HasTemplate - return true is the template is loaded and available
*/
func (p *Templates) HasTemplate(templateName string) bool {
	if p.templates == nil {
		return false
	}
	if p.templates[templateName] == nil {
		return false
	}
	return true
}

/*
HasAnyTemplates - return true is there are any templates
*/
func (p *Templates) HasAnyTemplates() bool {
	if p.templates == nil {
		return false
	}
	if len(p.templates) == 0 {
		return false
	}
	return true
}

/*
AddDataProvider - Add a method that will provide data to a template
*/
func (p *Templates) AddDataProvider(templateName string, provider func(*http.Request, string, interface{})) {
	if p.HasTemplate(templateName) {
		p.templates[templateName].dataProvider = provider
	} else {
		panic("Add Template Provider: Template[" + templateName + "] not found")
	}
}

/*
ListTemplateNames list template names
*/
func ListTemplateNames(delim string, t map[string]*templateData) string {
	var b bytes.Buffer
	mark := 0
	for key := range t {
		b.WriteString(key)
		mark = b.Len()
		b.WriteString(delim)
	}
	b.Truncate(mark)
	return b.String()
}

/*
ExecuteWriter writes a template to a io.Writer object
*/
func (p *Templates) executeWriter(w io.Writer, templateName string, data interface{}) {
	tmpl := p.templates[templateName]
	if tmpl == nil {
		ThrowPanic("E", 400, SCTemplateNotFound, fmt.Sprintf("Template '%s' not found", templateName), "")
	}
	err := tmpl.template.Execute(w, data)
	if err != nil {
		ThrowPanic("E", 400, SCTemplateError, fmt.Sprintf("Template '%s' error", templateName), err.Error())
	}
}

/*
ExecuteString writes a template to a string using ExecuteWriter
*/
func (p *Templates) executeString(templateName string, data interface{}) string {
	var buf bytes.Buffer
	p.executeWriter(&buf, templateName, data)
	return buf.String()
}

func (p *Templates) executeDataProvider(templateName string, r *http.Request, data interface{}) {
	if p.HasTemplate(templateName) {
		if p.templates[templateName].dataProvider != nil {
			p.templates[templateName].dataProvider(r, templateName, data)
		}
	}
}
