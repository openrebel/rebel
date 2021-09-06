package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"strings"

	brotli "github.com/andybalholm/brotli"
)

type CacheEntry struct {
	raw         []byte
	gzip        []byte
	brotli      []byte
	contentType string
}

var cache map[string]CacheEntry = make(map[string]CacheEntry)

var CONTENT_TYPE map[string]string = map[string]string{
	/* text */
	"htm":   "text/html; charset=utf-8",
	"html":  "text/html; charset=utf-8",
	"css":   "text/css; charset=utf-8",
	"txt":   "text/plain; charset=utf-8",
	"text":  "text/plain; charset=utf-8",
	"log":   "text/plain; charset=utf-8",
	"csv":   "text/csv; charset=utf-8",
	"xml":   "text/xml; charset=utf-8",
	"vcf":   "text/vcard; charset=utf-8",
	"vcard": "text/vcard; charset=utf-8",

	/* image */
	"ico":  "image/x-icon",
	"cur":  "application/octet-stream",
	"bmp":  "image/bmp",
	"png":  "image/png",
	"gif":  "image/gif",
	"tif":  "image/tiff",
	"tiff": "image/tiff",
	"jpg":  "image/jpeg",
	"jpe":  "image/jpeg",
	"jpeg": "image/jpeg",
	"webp": "image/webp",
	"svg":  "image/svg+xml; charset=utf-8",
	"svgz": "image/svg+xml; charset=utf-8",

	/* font */
	"otf": "font/otf",
	"ttf": "font/ttf",

	/* application */
	"js":   "application/javascript; charset=utf-8",
	"json": "application/json; charset=utf-8",
	"pdf":  "application/pdf",
	"zip":  "application/zip",
}

func initializeCache() {
	log.Println("Loading front end")

	loadCacheDir("./front")

	if entry, exist := (cache["index.html"]); exist {
		cache[""] = entry
		cache["/"] = entry
	}
}

func loadCacheDir(dirname string) {
	files, err := ioutil.ReadDir(dirname)

	if err != nil {
		log.Fatalln(err)
		return
	}

	for _, f := range files {
		if f.IsDir() {
			loadCacheDir(dirname + "/" + f.Name())
		} else {
			loadCacheFile(dirname + "/" + f.Name())
		}
	}
}

func loadCacheFile(filename string) {
	bytes, err := ioutil.ReadFile(filename)

	if err == nil {
		filename = filename[8:] //remove './front/'

		var contentType string
		var dot int = strings.LastIndex(filename, ".")
		if dot > -1 {
			var extention string = strings.ToLower(filename[dot+1:])
			contentType = CONTENT_TYPE[extention]
		}
		if contentType == "" {
			contentType = "text/html; charset=utf-8"
		}

		var entry CacheEntry
		entry.raw = bytes
		entry.gzip = gZipCompress(bytes)
		if product_release == "true" {
			entry.brotli = brotliCompress(bytes, 11)
		} else {
			entry.brotli = brotliCompress(bytes, 1)
		}
		entry.contentType = contentType
		cache[filename] = entry

	} else {
		log.Fatalln("File reading error", err.Error())
		return
	}
}

func gZipCompress(raw []byte) []byte {
	var buffer bytes.Buffer
	var gz *gzip.Writer = gzip.NewWriter(&buffer)

	if _, err := gz.Write(raw); err != nil {
		return nil
	}

	if gz.Flush() != nil {
		return nil
	}

	return buffer.Bytes()
}

func brotliCompress(raw []byte, quality int) []byte {
	var buffer bytes.Buffer
	var br *brotli.Writer = brotli.NewWriterOptions(&buffer, brotli.WriterOptions{Quality: quality})

	var in *bytes.Reader = bytes.NewReader(raw)
	n, err := io.Copy(br, in)
	if err != nil {
		return nil
	}

	if int(n) != len(raw) { //size mismatch
		return nil
	}

	if err := br.Close(); err != nil {
		return nil
	}

	return buffer.Bytes()
}
