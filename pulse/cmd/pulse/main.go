package main

import (
	"flag"
	"os"
)

var (
	debug bool
	api   bool
)

func init() {
	flag.BoolVar(&debug, "d", false, "Turn on debug mode")
	// so that flags are last and everything else (filenames are first)
	// that is how the flags parse
	permutateArgs(os.Args)
	flag.Parse()
}

func main() {
	if len(flag.Args()) == 0 {
		startAPI()
	} else {
		startPulse()
	}
}

func startAPI() {

}

func startPulse() {

}

func permutateArgs(args []string) int {
	args = args[1:]
	optind := 0

	for i := range args {
		if args[i][0] == '-' {
			tmp := args[i]
			args[i] = args[optind]
			args[optind] = tmp
			optind++
		}
	}

	return optind + 1
}
