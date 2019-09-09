package servermain

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

type fileServerContainer struct {
	path string
	root string
	fs   http.Handler
	next *fileServerContainer
}

/*
StaticFileServerData contains a list of fileServerContainer's
*/
type FileServerData struct {
	FileServerList *fileServerContainer
}

/*
NewStaticFileServer create a NEW static file server given a list of URL prefixs and root directories
*/
func NewStaticFileServer(mappings map[string]string) *FileServerData {

	sfs := &FileServerData{
		FileServerList: &fileServerContainer{
			path: "",
			root: "",
			fs:   nil,
			next: nil,
		},
	}
	for urlPrefix, root := range mappings {
		sfs.AddStaticFileServerData(urlPrefix, root)
	}
	return sfs
}

/*
AddFileServerData appends a URL prefix and a root directory
*/
func (p *FileServerData) AddStaticFileServerData(urlPrefix string, root string) {
	container := p.FileServerList
	for container.next != nil {
		container = container.next
	}
	container.path = "/" + urlPrefix + "/"
	container.root = root
	container.fs = http.FileServer(http.Dir(root))
	container.next = &fileServerContainer{
		path: "",
		root: "",
		fs:   nil,
		next: nil,
	}
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
	fileServerList := fileServerData.FileServerList
	if fileServerList == nil {
		ThrowPanic("E", 400, SCStaticFileInit, fmt.Sprintf("URL:%s Unsupported", url), "Static File Server Data - File Server List has not been defined.")
	}

	for fileServerList.fs != nil {
		if strings.HasPrefix(url, fileServerList.path) {
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
			fileShort := url[len(fileServerList.path):]
			filename := filepath.Join(fileServerList.root, fileShort)
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
		fileServerList = fileServerList.next
	}
	/*
		If not matching file was found then change the response to Not Found and return
	*/
	response.SetError404(url, SCContentNotFound)
}
