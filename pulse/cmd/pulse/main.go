package main

import (
	"flag"
	"os"

	"github.com/davecgh/go-spew/spew"
)

var (
	def bool
	api bool
)

func init() {
	flag.BoolVar(&def, "d", false, "Turn on default mode")
	flag.Parse()
}

func main() {
	var stdIn = make(chan string)
	var stdOut = make(chan string)
	if len(os.Args[1:]) == 0 {
		startAPI(stdIn, stdOut)
	} else {
		startPulse(stdIn, stdOut)
	}
}

func startAPI(stdIn, stdOut chan string) {
	spew.Println("API Mode")
}

func startPulse(stdIn, stdOut chan string) {
	if def {
		spew.Println("Defalut Mode")
	} else {
		spew.Println("Reading files from command line")
		for _, arg := range flag.Args() {
			spew.Dump(arg)
		}
	}
}
