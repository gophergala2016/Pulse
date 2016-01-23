package file

import (
	"bufio"
	"os"
	"strings"
)

//Read will read filename line by line and each line be returned to channel
func Read(filename string, lineOut chan<- string) {
	go func() {
		inFile, err := os.Open(filename)
		defer func() {
			inFile.Close()
			close(lineOut)
		}()
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(inFile)
		for scanner.Scan() {
			lineOut <- scanner.Text()
		}
	}()
}

//Write will append or create filename and write the slice of strings seperated by a new line
func Write(filename string, lines []string) {
	outFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0666)
	defer outFile.Close()
	if err != nil {
		panic(err)
	}
	longString := strings.Join(lines, "\n") + "\n"
	if _, err = outFile.WriteString(longString); err != nil {
		panic(err)
	}
}
