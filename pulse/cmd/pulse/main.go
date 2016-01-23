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
	def  bool
	api  bool
	smtp string
)

func init() {
	flag.BoolVar(&def, "d", false, "Turn on default mode")
	flag.StringVar(&smtp, "smtp", "SMTP.toml", "Config location of SMTP.toml")
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
		smtpCfg, err := config.LoadSMTP("C:\\Users\\dixon\\Go\\src\\github.com\\gophergala2016\\Pulse\\pulse\\cmd\\pulse\\SMTP.toml")
		if err != nil {
			panic(err)
		}
		spew.Dump(smtpCfg)
		secretCfg, err := config.LoadSecret("C:\\Users\\dixon\\Go\\src\\github.com\\gophergala2016\\Pulse\\pulse\\cmd\\pulse\\secret.toml")
		if err != nil {
			panic(err)
		}
		spew.Dump(secretCfg)
		cfg, err := config.Load("C:\\Users\\dixon\\Go\\src\\github.com\\gophergala2016\\Pulse\\pulse\\cmd\\pulse\\PulseConfig.toml")
		if err != nil {
			panic(err)
		}
		spew.Dump(cfg)
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
