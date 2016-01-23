package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/gophergala2016/Pulse/pulse"
	"github.com/gophergala2016/Pulse/pulse/config"
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
		cfg, err := config.Load(cfgPath)
		if err != nil {
			panic(fmt.Errorf("Could not load the config.\n %v", err))
		}
		spew.Dump(cfg)
		pulse.Run(stdIn, printFunc)
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
