package servermain

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"io"
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
				tmpl, err := template.ParseFiles(fullPath)
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
		return errIn
	})
	if walkError != nil {
		return nil, walkError
	}
	return templateList, nil
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
func (p *Templates) ExecuteWriter(w io.Writer, templateName string, data interface{}) (error) {
	tmpl := p.templates[templateName]
	if tmpl == nil {
		return errors.New("Template " + templateName + " not found")
	}
	var buf bytes.Buffer
	err := tmpl.template.Execute(&buf, data)
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
	err := p.ExecuteWriter(&buf,templateName, data)
	if err != nil {
		return "", errors.New("Template error: " + err.Error())
	}
	return buf.String(), nil
}
