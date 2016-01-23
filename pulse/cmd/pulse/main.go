package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/gophergala2016/Pulse/pulse"
	"github.com/gophergala2016/Pulse/pulse/config"
	"github.com/gophergala2016/Pulse/pulse/file"
)

var (
	def      bool
	smtpPath string
	cfgPath  string
)

func init() {
	flag.BoolVar(&def, "d", false, "Turn on default mode")
	flag.StringVar(&smtpPath, "smtp", "SMTP.toml", "Config location of SMTP.toml")
	flag.StringVar(&cfgPath, "cfg", "PulseConfig.toml", "Config locaton of PulseConfig.toml")
	flag.Parse()

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		panic(fmt.Errorf("Could not find %s", cfgPath))
	}
}

func main() {
	if len(flag.Args()) == 0 && !def {
		startAPI()
	} else if def {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			panic(fmt.Errorf("Could not load the config.\n %v", err))
		}
		if len(cfg.LogList) == 0 {
			panic(fmt.Errorf("Must supply a list of log files in the config."))
		}
		startPulse(cfg.LogList)
	} else {
		startPulse(flag.Args())
	}
}

func startAPI() {
	spew.Println("API Mode")
}

func startPulse(filenames []string) {
	stdIn := make(chan string)
	pulse.Run(stdIn, printString)
	checkList(filenames)
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
	errChan := make(chan error)
	go func(errChan chan error) {
		for _, filename := range filenames {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				errChan <- fmt.Errorf("Could not find %s", filename)
			}
		}
		close(errChan)
	}(errChan)

	for err := range errChan {
		panic(err)
	}
}

func printString(value string) {
	fmt.Println(value)
}
