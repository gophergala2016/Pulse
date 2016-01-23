package file

import (
	"bufio"
	"os"
)

//Read will read filename line by line and each line be returned to channel
func Read(filename string, lineOut chan<- string) {
	go func() {
		inFile, _ := os.Open(filename)
		defer func() {
			inFile.Close()
			close(lineOut)
		}()
		scanner := bufio.NewScanner(inFile)

		for scanner.Scan() {
			lineOut <- scanner.Text()
		}
	}()
}
