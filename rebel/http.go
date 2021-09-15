package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	websocket "github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var server HttpServer

type HttpServer struct {
	http_listeners  []string
	https_listeners []string
	alias           map[string]bool
	cache           Cache
}

func NewHttpServer() *HttpServer {
	var s *HttpServer = new(HttpServer)

	http.Handle("/", http.HandlerFunc(serve))
	http.Handle("/ws/keepalive", http.HandlerFunc(wsHandler))
	http.Handle("/ws/ping", http.HandlerFunc(WsPing))

	s.alias = make(map[string]bool)
	s.cache = *NewCache("./front")

	return s
}

func (s *HttpServer) Start() {
	for _, e := range s.http_listeners {
		var addr string = e
		go func() {
			log.Println("Initializing http listener on", addr)
			var err error = http.ListenAndServe(addr, nil)
			if err != nil {
				log.Fatalln(err)
				return
			}
		}()
	}

	for _, e := range s.https_listeners {
		var addr string = e
		go func() {
			log.Println("Initializing https listener on", addr)
			var err error = http.ListenAndServeTLS(addr, "./tls/server.crt", "./tls/server.key", nil)
			if err != nil {
				log.Fatalln(err)
				return
			}
		}()
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
	var acceptEncoding string = strings.ToLower(r.Header.Get("accept-encoding"))
	var acceptBr bool = strings.Contains(acceptEncoding, "br")
	var acceptGzip bool = strings.Contains(acceptEncoding, "gzip")

	//println(r.RemoteAddr, r.URL.String())

	switch r.RequestURI {

	case "/version":
		var version string = "{\"string\":\"" + product_version + "\"}"
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(version))
		//w.Header().Set("content-type", CONTENT_TYPE["json"])

	default: //check cache
		var file string = r.RequestURI
		if file[0] == '/' {
			file = file[1:]
		}

		var buffer *[]byte
		var status int

		if entry, exist := (server.cache.hash[file]); exist {
			status = http.StatusOK

			if acceptBr {
				buffer = &entry.brotli
				w.Header().Set("content-encoding", "br")

			} else if acceptGzip {
				buffer = &entry.gzip
				w.Header().Set("content-encoding", "gzip")

			} else {
				buffer = &entry.raw
			}

			w.Header().Set("content-type", entry.contentType)

		} else {
			status = http.StatusNotFound
		}

		if buffer == nil {
			w.WriteHeader(status)
		} else {
			w.Header().Set("content-length", strconv.Itoa(len(*buffer)))
			w.WriteHeader(status)
			w.Write(*buffer)
		}
	}
}

func CheckOrigin(r *http.Request) bool {
	var origin string = r.Header.Get("Origin")

	origin = strings.TrimPrefix(origin, "https://")
	origin = strings.TrimPrefix(origin, "http://")

	if _, exist := server.alias[origin]; exist {
		return true
	}

	return false
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	wsUpgrader.CheckOrigin = func(r *http.Request) bool {
		return CheckOrigin(r)
	}

	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	log.Print("ws connected")

	for {
		messageType, bytes, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		//log.Println(string(bytes))

		if err := ws.WriteMessage(messageType, bytes); err != nil {
			log.Println(err)
			return
		}
	}
}
