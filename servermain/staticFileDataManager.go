package servermain

import (
	"fmt"
	"strings"
)

/*
FileServerContainer contains the url prefix and associated file path
*/
type FileServerContainer struct {
	URLPrefix string
	FilePath  string
	next      *FileServerContainer
}

/*
StaticFileServerData contains a list of fileServerContainer's
*/
type StaticFileServerData struct {
	FileServerContainerRoot *FileServerContainer
}

/*
NewStaticFileServerData create a NEW static file server given a list of URL prefixs and root directories
*/
func NewStaticFileServerData(mappings map[string]string) *StaticFileServerData {
	sfs := &StaticFileServerData{
		FileServerContainerRoot: &FileServerContainer{
			URLPrefix: "",
			next:      nil,
		},
	}
	for urlPrefix, root := range mappings {
		sfs.AddStaticFileServerData(urlPrefix, root)
	}
	return sfs
}

/*
AddStaticFileServerData appends a URL prefix (or name) and a path to file system directory
*/
func (p *StaticFileServerData) AddStaticFileServerData(urlPrefix string, filePath string) {
	container := p.FileServerContainerRoot
	for container.next != nil {
		container = container.next
	}
	container.URLPrefix = urlPrefix
	container.FilePath = filePath
	container.next = &FileServerContainer{
		URLPrefix: "",
		FilePath:  "",
		next:      nil,
	}
}

/*
GetStaticPathForURL - return the file server container (path and url) for a given url. Matches from the start of the url
See tests for examples of matches
*/
func (p *StaticFileServerData) GetStaticPathForURL(url string) *FileServerContainer {
	container := p.FileServerContainerRoot
	if container == nil {
		ThrowPanic("E", 500, SCStaticFileInit, fmt.Sprintf("URL:%s Unsupported", url), "Static File Server Data - File Server List has not been defined.")
	}
	if strings.HasPrefix(url, "/") {
		return p.GetStaticPathForName(url)
	}
	return p.GetStaticPathForName("/" + url)
}

/*
GetStaticPathForName return the file server container (path and url) for a given name
See tests for examples of matches
*/
func (p *StaticFileServerData) GetStaticPathForName(name string) *FileServerContainer {
	container := p.FileServerContainerRoot
	if container == nil {
		ThrowPanic("E", 500, SCStaticFileInit, fmt.Sprintf("Name:%s Unsupported", name), "Static File Server Data - File Server List has not been defined.")
	}
	var resp *FileServerContainer
	var l = 0
	for container.next != nil {
		if strings.HasPrefix(name, container.URLPrefix) {
			if l <= len(container.URLPrefix) {
				l = len(container.URLPrefix)
				resp = container
			}
		}
		container = container.next
	}
	if resp == nil {
		ThrowPanic("W", 404, SCStaticPathNotFound, fmt.Sprintf("Entity:%s Not Found", name), fmt.Sprintf("Static File Server Data. Entity:%s is not defined", name))
	}
	return resp
}

