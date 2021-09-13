package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var product_release string
var product_version string

func main() {
	println("           _          _")
	println("  _ __ ___| |__   ___| |")
	println(" | '__/ _ \\  _ \\ / _ \\ |")
	println(" | | |  __/ |_) |  __/ |")
	println(" |_|  \\___|_.__/ \\___|_|")

	if product_release == "true" {
		for i := 0; i < 24-len(product_version); i++ {
			print(" ")
		}

	} else {
		log.SetFlags(log.Lshortfile)
		print(" Development")
		for i := 0; i < 12-len(product_version); i++ {
			print(" ")
		}
	}

	println(product_version)
	println()

	server = *NewHttpServer()
	loadConfig()
	server.Start()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		scanner.Text()
	}
}

func loadConfig() {
	bytes, err := ioutil.ReadFile("./rebel.cfg")

	if err == nil {
		var lines []string = strings.Split(string(bytes), "\n")

		for i := 0; i < len(lines); i++ {
			lines[i] = strings.Trim(lines[i], " ")
			if len(lines[i]) == 0 || strings.HasPrefix(lines[i], "#") {
				continue
			}

			if !strings.Contains(lines[i], "=") {
				continue
			}

			var split []string = strings.Split(lines[i], "=")
			if len(split) < 2 {
				continue
			}
			var a string = strings.TrimSpace(split[0])
			var b string = strings.TrimSpace(split[1])

			if a == "listen_http" {
				server.http_listeners = append(server.http_listeners, b)
			} else if a == "listen_https" {
				server.https_listeners = append(server.https_listeners, b)
			} else if a == "alias" {
				server.alias[b] = true
			}
		}

	} else {
		log.Println("Failed to load configuration file. Loading the default configuration.")

		server.http_listeners = append(server.http_listeners, "127.0.0.1:80")
		server.alias["localhost"] = true
	}

	for _, e := range server.http_listeners {
		server.alias[e] = true
		var split []string = strings.Split(e, ":")
		if split[1] == "80" {
			server.alias[split[0]] = true
		}
	}

	for _, e := range server.https_listeners {
		server.alias[e] = true
		var split []string = strings.Split(e, ":")
		if split[1] == "443" {
			server.alias[split[0]] = true
		}
	}
}
