package pulse

import "fmt"

var input <-chan string
var output chan<- string

//Run starts the pulse package
func Run(in <-chan string, out chan<- string) {
	input = in
	output = out

	for value := range in {
		fmt.Println(value)
	}
}
