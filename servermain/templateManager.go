package servermain

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"webServerBase/logging"
)

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
		if strings.Contains(path, ".template.") {
			fullPath, filePathErr := filepath.Abs(path)
			if filePathErr == nil {
				_, tname := filepath.Split(fullPath)
				fname := strings.Replace(tname, ".template", "", 1)
				tmpl := template.Must(template.New(fname).ParseFiles(fullPath))
				templateList.templates[fname] = &templateData{
					name:     fname,
					file:     fullPath,
					template: tmpl,
				}

				logger.LogInfof("Loading: FILE:%s NAME:%s PATH:%s", tname, fname, fullPath)
			}
			return filePathErr
		}
		return errIn
	})
	if walkError != nil {
		return nil, walkError
	}
	return templateList, nil
}

/*
Execute a template
*/
func (p *Templates) Execute(templateName string, data interface{}) (string, error) {
	tmpl := p.templates[templateName]
	if tmpl == nil {
		return "", errors.New("Template " + templateName + " not found")
	}
	var buf bytes.Buffer
	err := tmpl.template.Execute(&buf, data)
	if err != nil {
		return "", errors.New("Template error: " + err.Error())
	}
	return buf.String(), nil
}
