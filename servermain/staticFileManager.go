package servermain

import (
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
StaticFileServer contains a list of fileServerContainer's
*/
type StaticFileServer struct {
	serverInstance *ServerInstanceData
	fileServerList *fileServerContainer
}

/*
NewStaticFileManager create a NEW static file server
*/
func NewStaticFileServer(mappings map[string]string, server *ServerInstanceData) *StaticFileServer {

	sfs := &StaticFileServer{
		serverInstance: server,
		fileServerList: &fileServerContainer{
			path: "",
			root: "",
			fs:   nil,
			next: nil,
		},
	}
	for key, value := range mappings {
		sfs.AddFileServerData(key, value)
	}
	return sfs
}

/*
AddFileServerData creates a file server for a path and a root directory
*/
func (p *StaticFileServer) AddFileServerData(path string, root string) {
	container := p.fileServerList
	for container.next != nil {
		container = container.next
	}
	container.path = "/" + path + "/"
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
	fileServer := server.staticFileServer
	if fileServer == nil {
		response.SetError404(url)
		return
	}
	fileServerList := fileServer.fileServerList
	if fileServer == nil {
		response.SetError404(url)
		return
	}

	for fileServerList.fs != nil {
		if strings.HasPrefix(url, fileServerList.path) {
			contentType, _ := server.LookupContentType(url)
			if contentType != "" {
				response.GetHeaders()[contentTypeName] = []string{contentType + "; charset=" + server.contentTypeCharset}
			}
			filename := filepath.Join(fileServerList.root, url[len(fileServerList.path):])
			err := ServeContent(response.GetWrappedWriter(), request, filename)
			if err != nil {
				response.ChangeResponse(400, "ServeContent", "", err)
				return
			}
			response.Close()
			return
		}
		fileServerList = fileServerList.next
	}
	response.SetError404(url)
}
