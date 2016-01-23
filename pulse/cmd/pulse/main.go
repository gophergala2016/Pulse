package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/gophergala2016/Pulse/pulse"
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
	if len(os.Args[1:]) == 0 {
		startAPI()
	} else {
		startPulse()
	}
}

func startAPI() {
	spew.Println("API Mode")
}

func startPulse() {
	var stdIn = make(chan string)
	if def {
		pulse.Run(stdIn, printFunc)
		stdIn <- "Hello World"
		stdIn <- "Because Tesla"
		stdIn <- "Why not"

	} else {
		spew.Println("Reading files from command line")
		for _, arg := range flag.Args() {
			spew.Dump(arg)
		}
	}
}

func printFunc(value string) {
	fmt.Println(value)
}
