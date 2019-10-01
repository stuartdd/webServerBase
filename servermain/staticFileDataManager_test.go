package servermain

import (
	"testing"

	"github.com/stuartdd/webServerBase/test"
)

func TestStaticFileForUrlPrefix(t *testing.T) {
	sfm := createStaticFileServer()
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("/static/name").FilePath, "site/")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("static/name").FilePath, "site/")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("/static/bmp/name").FilePath, "site/bmp")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("static/bmp/name").FilePath, "site/bmp")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("/static/xmp/name").FilePath, "site/")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("/static/bmp/").FilePath, "site/bmp")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("/static/bmp").FilePath, "site/")

	notFound(t, sfm, "/statxc/xmp/name")
	notFound(t, sfm, "statxc/xmp/name")
	notFound(t, sfm, "/main/name")

	sfm.AddStaticFileServerData("/main/", "local")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("/main/name").FilePath, "local")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForURL("main/name").FilePath, "local")

	notFound(t, sfm, "/statxc/xmp/name")
	notFound(t, sfm, "statxc")
	notFound(t, sfm, "static")

	sfm.AddStaticFileServerData("static", "xxxx")
	test.AssertStringEquals(t, "", sfm.GetStaticPathForName("static").FilePath, "xxxx")

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
