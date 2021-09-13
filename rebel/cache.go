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

type Cache struct {
	path string
	hash map[string]CacheEntry
}

type CacheEntry struct {
	raw         []byte
	gzip        []byte
	brotli      []byte
	contentType string
}

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

func NewCache(path string) *Cache {
	var c *Cache = new(Cache)
	c.path = path
	c.hash = make(map[string]CacheEntry)

	log.Printf("Loading front end: %s", path)

	loadCacheDir(c, c.path)

	if entry, exist := c.hash["index.html"]; exist {
		c.hash[""] = entry
		c.hash["/"] = entry
	}

	return c
}

func loadCacheDir(c *Cache, dirname string) {
	files, err := ioutil.ReadDir(dirname)

	if err != nil {
		log.Fatalln(err)
		return
	}

	for _, f := range files {
		if f.IsDir() {
			loadCacheDir(c, dirname+"/"+f.Name())
		} else {
			loadCacheFile(c, dirname+"/"+f.Name())
		}
	}
}

func loadCacheFile(c *Cache, filename string) {
	bytes, err := ioutil.ReadFile(filename)

	if err == nil {
		filename = filename[len(c.path)+1:] //remove './front/'

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
		c.hash[filename] = entry

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
