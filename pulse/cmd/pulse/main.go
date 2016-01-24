package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/davecgh/go-spew/spew"
	"github.com/gophergala2016/Pulse/pulse"
	"github.com/gophergala2016/Pulse/pulse/api"
	"github.com/gophergala2016/Pulse/pulse/config"
	"github.com/gophergala2016/Pulse/pulse/email"
	"github.com/gophergala2016/Pulse/pulse/file"
)

var (
	def         bool
	outputFile  string
	buffStrings []string
	logList     []string
)

func init() {
	flag.BoolVar(&def, "d", false, "Turn on default mode")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("Could not load the config.\n %v", err))
	}

	logList = cfg.LogList
	outputFile = cfg.OutputFile
}

func main() {

	if len(flag.Args()) == 0 && !def {
		startAPI()
	} else if def {
		if len(logList) == 0 {
			panic(fmt.Errorf("Must supply a list of log files in the config."))
		}
		startPulse(logList)
	} else {
		startPulse(flag.Args())
	}
}

func startAPI() {
	api.Start()
}

func startPulse(filenames []string) {
	spew.Dump(filenames)
	checkList(filenames)
	stdIn := make(chan string)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// On keyboard interrup cleanup the program
	go func(i chan string) {
		for _ = range c {
			fmt.Println("Exiting for Keyboard Interupt")
			close(i)
			email.DumpBuffer()
			os.Exit(0)
		}
	}(stdIn)

	pulse.Run(stdIn, email.Send)
	for _, filename := range filenames {
		line := make(chan string)
		file.Read(filename, line)
		for l := range line {
			stdIn <- l
		}
	}
	close(stdIn)
}

func checkList(filenames []string) {
	for _, filename := range filenames {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			panic(err)
		}
	}
}
