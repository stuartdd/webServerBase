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
StaticFileServerData contains a list of fileServerContainer's
*/
type StaticFileServerData struct {
	FileServerList *fileServerContainer
}

/*
NewStaticFileServer create a NEW static file server
*/
func NewStaticFileServer(mappings map[string]string) *StaticFileServerData {

	sfs := &StaticFileServerData{
		FileServerList: &fileServerContainer{
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
func (p *StaticFileServerData) AddFileServerData(path string, root string) {
	container := p.FileServerList
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
	fileServerData := server.staticFileServerData
	/*
	If an there is no file server data then change the response to Not Found and return
	*/
	if fileServerData == nil {
		response.SetError404(url)
		return
	}
	/*
	If an there is no file server list then change the response to Not Found and return
	*/
	fileServerList := fileServerData.FileServerList
	if fileServerList == nil {
		response.SetError404(url)
		return
	}

	for fileServerList.fs != nil {
		if strings.HasPrefix(url, fileServerList.path) {
			/*
			Work out the content type from the file name extension
			*/
			contentType, _ := server.LookupContentType(url)
			if contentType != "" {
				response.GetHeaders()[contentTypeName] = []string{contentType + "; charset=" + server.contentTypeCharset}
			}
			/*
			derive the file name from the url and the path in the fileServerList
			*/
			filename := filepath.Join(fileServerList.root, url[len(fileServerList.path):])
			err := ServeContent(response.GetWrappedWriter(), request, filename)
			if err != nil {
				/*
				If an error occured change the response to an error and return
				*/
				response.ChangeResponse(400, "ServeContent", "", err)
				return
			}
			/*
			The file is being written to the response writer.

			Close the response to prevent further writes to the response writer
			*/
			response.Close()
			return
		}
		fileServerList = fileServerList.next
	}
	/*
	If not matching file was found then change the response to Not Found and return
	*/
	response.SetError404(url)
}
