package file

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"mime/multipart"
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
			panic(fmt.Errorf("file.Read: %s", err))
		}
		scanner := bufio.NewScanner(inFile)
		for scanner.Scan() {
			lineOut <- scanner.Text()
		}
	}()
}

//StreamRead will read from io.Reader line by line and each line be returned to channel
func StreamRead(reader multipart.File, lineOut chan<- string) {
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			val := scanner.Text()
			fmt.Println(val)
			lineOut <- val
		}
		lineOut <- "EOF"
	}()
}

//Write will append or create filename and write the slice of strings seperated by a new line
func Write(filename string, lines []string) {
	outFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	defer outFile.Close()
	if err != nil {
		panic(fmt.Errorf("file.Write: %s", err))
	}
	longString := strings.Join(lines, "\n") + "\n"
	if _, err = outFile.WriteString(longString); err != nil {
		panic(fmt.Errorf("file.Write: %s", err))
	}
}

// UnGZip : used to uncompress files
func UnGZip(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	newfile := filename[0 : len(filename)-len(".gz")]
	out, err := os.Create(newfile)
	if err != nil {
		return err
	}

	defer out.Close()

	r, err := gzip.NewReader(file)
	io.Copy(out, r)
	r.Close()

	return nil
}
