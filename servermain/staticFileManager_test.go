package servermain

import (
	"testing"
	"webServerBase/test"
)

func TestStaticFileForUrlPrefix(t *testing.T) {
	mapData := make(map[string]string)
	mapData["/static/bmp/"] = "site/bmp"
	mapData["/static/"] = "site/"
	sfm := NewStaticFileServer(mapData)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/name").FsPath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("static/name").FsPath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("/static/bmp/name").FsPath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("static/bmp/name").FsPath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/xmp/name").FsPath)
	test.AssertStringEquals(t, "", "site/bmp", sfm.GetStaticPathForURL("/static/bmp/").FsPath)
	test.AssertStringEquals(t, "", "site/", sfm.GetStaticPathForURL("/static/bmp").FsPath)
	test.AssertNil(t, "", sfm.GetStaticPathForURL("/statxc/xmp/name"))
	test.AssertNil(t, "", sfm.GetStaticPathForURL("statxc/xmp/name"))
	test.AssertNil(t, "", sfm.GetStaticPathForURL("/main/name"))

	sfm.AddStaticFileServerData("/main/", "local")
	test.AssertStringEquals(t, "", "local", sfm.GetStaticPathForURL("/main/name").FsPath)
	test.AssertStringEquals(t, "", "local", sfm.GetStaticPathForURL("main/name").FsPath)

	test.AssertNil(t, "", sfm.GetStaticPathForName("/statxc/xmp/name"))
	test.AssertNil(t, "", sfm.GetStaticPathForName("statxc"))
	test.AssertNil(t, "", sfm.GetStaticPathForName("static"))

	sfm.AddStaticFileServerData("static", "xxxx")
	test.AssertStringEquals(t, "", "xxxx", sfm.GetStaticPathForName("static").FsPath)

}
