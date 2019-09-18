package servermain

import (
	"testing"
	"webServerBase/test"
)

func TestStaticFileForUrlPrefix(t *testing.T) {
	sfm := createStaticFileServer()
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/name").FsPath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("static/name").FsPath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("/static/bmp/name").FsPath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("static/bmp/name").FsPath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/xmp/name").FsPath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("/static/bmp/").FsPath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/bmp").FsPath)

	notFound(t, sfm, "/statxc/xmp/name")
	notFound(t, sfm, "statxc/xmp/name")
	notFound(t, sfm, "/main/name")

	sfm.AddStaticFileServerData("/main/", "local")
	test.AssertStringEquals(t, "", "local", sfm.GetStaticPathForURL("/main/name").FsPath)
	test.AssertStringEquals(t, "", "local", sfm.GetStaticPathForURL("main/name").FsPath)

	notFound(t, sfm, "/statxc/xmp/name")
	notFound(t, sfm, "statxc")
	notFound(t, sfm, "static")

	sfm.AddStaticFileServerData("static", "xxxx")
	test.AssertStringEquals(t, "", "xxxx", sfm.GetStaticPathForName("static").FsPath)

}

func notFound(t *testing.T, sfm *FileServerData,  name string) {
	defer test.AssertPanicAndRecover(t, "Entity:"+name+" is not defined")
	test.AssertNil(t, "", sfm.GetStaticPathForName(name))
}

func createStaticFileServer() *FileServerData {
	mapData := make(map[string]string)
	mapData["/static/bmp/"] = "site/bmp"
	mapData["/static/"] = "site/"
	return NewStaticFileServer(mapData)
}