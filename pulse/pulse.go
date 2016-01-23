package pulse

import "fmt"

var input <-chan string
var output chan<- string

type callback func(string)

//Run starts the pulse package
func Run(in <-chan string, cb callback) {
	input = in
	go func() {
		for value := range in {
			fmt.Println(value)
			cb("haha")
		}
	}()
}
