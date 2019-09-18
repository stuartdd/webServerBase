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
	FsPath    string
	next      *FileServerContainer
}

/*
FileServerData contains a list of fileServerContainer's
*/
type FileServerData struct {
	FileServerContainerRoot *FileServerContainer
}

/*
NewStaticFileServer create a NEW static file server given a list of URL prefixs and root directories
*/
func NewStaticFileServer(mappings map[string]string) *FileServerData {

	sfs := &FileServerData{
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
AddStaticFileServerData appends a URL prefix and a root directory
*/
func (p *FileServerData) AddStaticFileServerData(urlPrefix string, root string) {
	container := p.FileServerContainerRoot
	for container.next != nil {
		container = container.next
	}
	container.URLPrefix = urlPrefix
	container.FsPath = root
	container.next = &FileServerContainer{
		URLPrefix: "",
		FsPath:    "",
		next:      nil,
	}
}

/*
GetStaticPathForURL - Get static path for the front of a url.
	The url will have '\' added to start if not already there
*/
func (p *FileServerData) GetStaticPathForURL(url string) *FileServerContainer {
	container := p.FileServerContainerRoot
	if container == nil {
		ThrowPanic("E", 400, SCStaticFileInit, fmt.Sprintf("URL:%s Unsupported", url), "Static File Server Data - File Server List has not been defined.")
	}
	if strings.HasPrefix(url, "/") {
		return p.GetStaticPathForName(url)
	}
	return p.GetStaticPathForName("/" + url)
}

/*
GetStaticPathForName return the file server container for a given url prefix
*/
func (p *FileServerData) GetStaticPathForName(name string) *FileServerContainer {
	container := p.FileServerContainerRoot
	if container == nil {
		ThrowPanic("E", 400, SCStaticFileInit, fmt.Sprintf("Name:%s Unsupported", name), "Static File Server Data - File Server List has not been defined.")
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
	if (resp == nil) {
		ThrowPanic("W", 404, SCStaticPathNotFound, fmt.Sprintf("Entity:%s Not Found", name), fmt.Sprintf("Static File Server Data. Entity:%s is not defined", name))
	}
	return resp
}

/*
ReasonableStaticFileHandler Read a file from a static file location and return it
*/
func ReasonableStaticFileHandler(request *http.Request, response *Response) {
	url := request.URL.Path
	server := response.GetWrappedServer()
	fileServerData := server.fileServerData
	/*
		If an there is no file server data then change the response to Not Found and return
	*/
	if fileServerData == nil {
		ThrowPanic("E", 500, SCStaticFileInit, fmt.Sprintf("URL:%s Unsupported", url), "Static File Server Data has not been defined.")
	}
	/*
		If an there is no file server list then change the response to Not Found and return
	*/
	container := fileServerData.GetStaticPathForURL(url)
	if container != nil {
		/*
			Forward the response headers in to the wrapped http.ResponseWriter
		*/
		ww := response.GetWrappedWriter()
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
		filename := filepath.Join(container.FsPath, fileShort)
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
	/*
		If not matching file was found then change the response to Not Found and return
	*/
	response.SetError404(url, SCContentNotFound)
}
