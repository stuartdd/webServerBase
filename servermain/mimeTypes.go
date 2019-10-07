package servermain

import "strings"

var contentTypesMap = makeContentTypesMap()

/*
from : https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Complete_list_of_MIME_types
*/
func makeContentTypesMap() map[string]string {
	mime := make(map[string]string)
	mime["aac"] = "audio/aac"
	mime["abw"] = "application/x-abiword"
	mime["arc"] = "application/x-freearc"
	mime["avi"] = "video/x-msvideo"
	mime["azw"] = "application/vnd.amazon.ebook"
	mime["bin"] = "application/octet-stream"
	mime["bmp"] = "image/bmp"
	mime["bz"] = "application/x-bzip"
	mime["bz2"] = "application/x-bzip2"
	mime["csh"] = "application/x-csh"
	mime["css"] = "text/css"
	mime["csv"] = "text/csv"
	mime["doc"] = "application/msword"
	mime["docx"] = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	mime["eot"] = "application/vnd.ms-fontobject"
	mime["epub"] = "application/epub+zip"
	mime["gif"] = "image/gif"
	mime["htm"] = "text/html"
	mime["html"] = "text/html"
	mime["ico"] = "image/vnd.microsoft.icon" // Some browsers use image/x-icon. Add to config data to override!
	mime["ics"] = "text/calendar"
	mime["jar"] = "application/java-archive"
	mime["jpeg"] = "image/jpeg"
	mime["jpg"] = "image/jpeg"
	mime["js"] = "text/javascript"
	mime["json"] = "application/json"
	mime["jsonld"] = "application/ld+json"
	mime["mid"] = "audio/midi audio/x-midi"
	mime["midi"] = "audio/midi audio/x-midi"
	mime["mjs"] = "text/javascript"
	mime["mp3"] = "audio/mpeg"
	mime["mpeg"] = "video/mpeg"
	mime["mpkg"] = "application/vnd.apple.installer+xml"
	mime["odp"] = "application/vnd.oasis.opendocument.presentation"
	mime["ods"] = "application/vnd.oasis.opendocument.spreadsheet"
	mime["odt"] = "application/vnd.oasis.opendocument.text"
	mime["oga"] = "audio/ogg"
	mime["ogv"] = "video/ogg"
	mime["ogx"] = "application/ogg"
	mime["otf"] = "font/otf"
	mime["png"] = "image/png"
	mime["pdf"] = "application/pdf"
	mime["ppt"] = "application/vnd.ms-powerpoint"
	mime["pptx"] = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	mime["rar"] = "application/x-rar-compressed"
	mime["rtf"] = "application/rtf"
	mime["sh"] = "application/x-sh"
	mime["svg"] = "image/svg+xml"
	mime["swf"] = "application/x-shockwave-flash"
	mime["tar"] = "application/x-tar"
	mime["tif"] = "image/tiff"
	mime["tiff"] = "image/tiff"
	mime["ts"] = "video/mp2t"
	mime["ttf"] = "font/ttf"
	mime["txt"] = "text/plain"
	mime["vsd"] = "application/vnd.visio"
	mime["wav"] = "audio/wav"
	mime["weba"] = "audio/webm"
	mime["webm"] = "video/webm"
	mime["webp"] = "image/webp"
	mime["woff"] = "font/woff"
	mime["woff2"] = "font/woff2"
	mime["xhtml"] = "application/xhtml+xml"
	mime["xls"] = "application/vnd.ms-excel"
	mime["xlsx"] = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	mime["xml"] = "application/xml"
	mime["xul"] = "application/vnd.mozilla.xul+xml"
	mime["zip"] = "application/zip"
	mime["7z"] = "application/x-7z-compressed"
	return mime
}

/*
AddNewContentTypeToMap for a given url return the content type based on the .ext
*/
func AddNewContentTypeToMap(ext string, mime string) {
	contentTypesMap[ext] = mime
}

/*
LookupContentType for a given url return the content type based on the .ext
*/
func LookupContentType(url string) string {
	ext := url
	pos := strings.LastIndex(url, ".")
	if pos > 0 {
		ext = url[pos+1:]
	}
	mapping, found := contentTypesMap[ext]
	if found {
		return mapping
	}
	return ""
}
