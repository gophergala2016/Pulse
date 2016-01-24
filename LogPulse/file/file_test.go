package file_test

import (
	"log"
	"os"
	"testing"

	. "github.com/gophergala2016/Pulse/LogPulse/file"
)

func TestRead(t *testing.T) {
	line := make(chan string)
	Read("TestData/smallkern.log.1", line)
	for l := range line {
		if l == "" {
			t.Errorf("There should not be an empty line while reading")
		}
	}
}

func TestUnGZip(t *testing.T) {
	defer func() {
		if _, err := os.Stat("TestData/smallkern.log.2"); os.IsNotExist(err) {
			log.Println("No file to remove")
		}
		if err := os.Remove("TestData/smallkern.log.2"); err != nil {
			log.Println("Could not remove the file")
		}
	}()
	UnGZip("TestData/smallkern.log.2.gz")
	if _, err := os.Stat("TestData/smallkern.log.2"); os.IsNotExist(err) {
		t.Errorf("Could not find unziped file")
	}
}
