package servermain

import (
	"testing"
	"github.com/stuartdd/webServerBase/test"
)

func TestStaticFileForUrlPrefix(t *testing.T) {
	sfm := createStaticFileServer()
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/name").FilePath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("static/name").FilePath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("/static/bmp/name").FilePath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("static/bmp/name").FilePath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/xmp/name").FilePath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("/static/bmp/").FilePath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/bmp").FilePath)

	notFound(t, sfm, "/statxc/xmp/name")
	notFound(t, sfm, "statxc/xmp/name")
	notFound(t, sfm, "/main/name")

	sfm.AddStaticFileServerData("/main/", "local")
	test.AssertStringEquals(t, "", "local", sfm.GetStaticPathForURL("/main/name").FilePath)
	test.AssertStringEquals(t, "", "local", sfm.GetStaticPathForURL("main/name").FilePath)

	notFound(t, sfm, "/statxc/xmp/name")
	notFound(t, sfm, "statxc")
	notFound(t, sfm, "static")

	sfm.AddStaticFileServerData("static", "xxxx")
	test.AssertStringEquals(t, "", "xxxx", sfm.GetStaticPathForName("static").FilePath)

}

func notFound(t *testing.T, sfm *StaticFileServerData, name string) {
	defer test.AssertPanicAndRecover(t, "Entity:"+name+" is not defined")
	test.AssertNil(t, "", sfm.GetStaticPathForName(name))
}

func createStaticFileServer() *StaticFileServerData {
	mapData := make(map[string]string)
	mapData["/static/bmp/"] = "site/bmp"
	mapData["/static/"] = "site/"
	return NewStaticFileServerData(mapData)
}
