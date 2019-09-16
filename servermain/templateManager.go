package servermain

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"webServerBase/logging"
)

type templateGroup struct {
	Name      string
	Templates []string
}

var groupList []templateGroup

type templateData struct {
	name     string
	file     string
	template *template.Template
}

/*
Templates list of templates by ID
*/
type Templates struct {
	templates map[string]*templateData
}

var logger *logging.LoggerDataReference

/*
LoadTemplates - Load the templates given the template paths. For example
	"templates\\"

Templates should be named *.template.* in order to be parsed!

The resulting template name is the name with '.template' removed
*/
func LoadTemplates(templatePath string) (*Templates, error) {
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


func loadJSONGroupList(fileName string, obj interface{})  error {
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
	if p.templates[templateName] == nil {
		return false
	}
	return true
}

/*
HasAnyTemplates - return true is there are any templates
*/
func (p *Templates) HasAnyTemplates() bool {
	if len(p.templates) == 0 {
		return false
	}
	return true
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
func (p *Templates) ExecuteWriter(w io.Writer, templateName string, data interface{}) {
	tmpl := p.templates[templateName]
	if tmpl == nil {
		ThrowPanic("E",400,SCTemplateNotFound, fmt.Sprintf("Template '%s' not found", templateName),"")
	}
	err := tmpl.template.Execute(w, data)
	if err != nil {
		ThrowPanic("E",400,SCTemplateError, fmt.Sprintf("Template '%s' error", templateName),err.Error())
	}
}

/*
ExecuteString writes a template to a string using ExecuteWriter
*/
func (p *Templates) ExecuteString(templateName string, data interface{}) (string) {
	var buf bytes.Buffer
	p.ExecuteWriter(&buf, templateName, data)
	return buf.String()
}
