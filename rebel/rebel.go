package main

import (
	"io/ioutil"
	"log"
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

	loadConfig()
	initializeCache()
	initializeHttpListener()
}

func loadConfig() {
	bytes, err := ioutil.ReadFile("./rebel.cfg")

	if err == nil {
		var lines []string = strings.Split(string(bytes), " ")
		var i int
		for i = 0; i < len(lines); i++ {
			lines[i] = strings.Trim(lines[i], " ")
			if len(lines[i]) > 0 && lines[i][1] == '#' {
				continue
			}
			//TODO:
		}

	} else {
		log.Println("Failed to load configuration file. Loading the default configuration.")
	}
}
