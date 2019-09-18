package servermain

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

/*
FileServerContainer contains the url prefix and associated file path
*/
type FileServerContainer struct {
	URLPrefix string
	FilePath    string
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
		FilePath:    "",
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

/*
ReasonableStaticFileHandler Read a file from a static file location and return it
*/
func ReasonableStaticFileHandler(request *http.Request, response *Response) {
	h := NewRequestHandlerHelper(request, response)
	url := h.GetURL()
	server := h.GetServer()
	/*
		get the file path for the url and the matched url. Panics if not found with 404 so no need to check.
	*/
	container := h.GetStaticPathForURL(url)
	/*
		Forward the response headers in to the wrapped http.ResponseWriter
	*/
	ww := h.GetResponseWriter()
	for name, value := range response.GetHeaders() {
		ww.Header()[name] = value
	}
	/*
		Work out the content type from the file name extension and add it to the wrapped http.ResponseWriter
		Dont overwrite a content type that already exists!
	*/
	contentType := LookupContentType(url)
	if (contentType != "") && (ww.Header()[ContentTypeName] == nil) {
		ww.Header()[ContentTypeName] = []string{contentType + "; charset=" + server.contentTypeCharset}
	}
	/*
		derive the file name from the url and the path in the fileServerList
	*/
	fileShort := url[len(container.URLPrefix):]
	filename := filepath.Join(container.FilePath, fileShort)
	/*
	Implemented in servermain/serverInstanceData.go. This wraps the http.ServeContent to return the file contents.
	Panics 404 if file not found. Panics 500 if file cannot be read 
	*/
	ServeContent(ww, request, filename)
	/*
		The file is being written to the response writer.
		Close the response to prevent further writes to the response writer
	*/
	response.Close()
	/*
		Specific ACCESS log entry for data returned from a file. Dont echo the response as this is a stream. Just indicate the file.
		Dont log the full file name as this reveals the server file system structure and can lead to vulnerabilities.
	*/
	if response.GetWrappedServer().GetServerLogger().IsAccess() {
		response.GetWrappedServer().GetServerLogger().LogAccessf("<<< STATUS=%d: CODE=%d: RESP-FROM-FILE=%s: TYPE=%s", response.GetCode(), response.GetSubCode(), fileShort, contentType)
		response.GetWrappedServer().logHeaderMap(response.GetHeaders(), "<-<")
	}
	return
}
