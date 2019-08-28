package servermain

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"webServerBase/logging"

	jsonconfig "github.com/stuartdd/tools_jsonconfig"
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

var logger = logging.NewLogger("templates")

/*
LoadTemplates - Load the templates given the template paths. For example
	"templates\\"

Templates should be named *.template.* in order to be parsed!

The resulting template name is the name with '.template' removed
*/
func LoadTemplates(templatePath string) (*Templates, error) {
	templateList := &Templates{
		templates: make(map[string]*templateData),
	}
	walkError := filepath.Walk(templatePath, func(path string, info os.FileInfo, errIn error) error {
		if strings.Contains(path, ".template.groups.") {
			return loadGroupOfTemplates(path, templateList)
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
func loadGroupOfTemplates(path string, templateList *Templates) error {
	fullPath, filePathErr := filepath.Abs(path)
	if filePathErr != nil {
		if filePathErr != nil {
			return filePathErr
		}
		err := jsonconfig.LoadJson(fullPath, &groupList)
		if err != nil {
			return err
		}
		for _, group := range groupList {
			tmpl, err := template.ParseFiles(group.Templates...)
			if err != nil {
				return err
			}
			templateList.templates[group.Name] = &templateData{
				name:     group.Name,
				file:     fullPath,
				template: tmpl,
			}
		}
	}
	return filePathErr
}

func loadSingletemplate(path string, templateList *Templates) error {
	fullPath, filePathErr := filepath.Abs(path)
	if filePathErr == nil {
		_, tname := filepath.Split(fullPath)
		fname := strings.Replace(tname, ".template", "", 1)
		var tmpl *template.Template
		var err error
		if fname == "import1.html" {
			// tmpl, err = template.ParseFiles(fullPath, "C:\\Users\\802996013\\go\\src\\webServerBase\\site\\import2.template.html", "C:\\Users\\802996013\\go\\src\\webServerBase\\site\\simple2.template.html")
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
		logger.LogInfof("Loading: FILE:%s NAME:%s PATH:%s", tname, fname, fullPath)
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
ExecuteWriter writes a template to a io.Writer object
*/
func (p *Templates) ExecuteWriter(w io.Writer, templateName string, data interface{}) error {
	tmpl := p.templates[templateName]
	if tmpl == nil {
		return errors.New("Template " + templateName + " not found")
	}
	err := tmpl.template.Execute(w, data)
	if err != nil {
		return errors.New("Template error: " + err.Error())
	}
	return nil
}

/*
ExecuteString writes a template to a string using ExecuteWriter
*/
func (p *Templates) ExecuteString(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := p.ExecuteWriter(&buf, templateName, data)
	if err != nil {
		return "", errors.New("Template error: " + err.Error())
	}
	return buf.String(), nil
}
