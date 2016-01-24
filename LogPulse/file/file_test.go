package file_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/gophergala2016/Pulse/LogPulse/file"
)

func TestRead(t *testing.T) {
	filename := "TestData/ReadTest.txt"
	var fileArray []string
	expectedArray := []string{"This is a line.", "This is a new line."}
	line := make(chan string)
	Read(filename, line)

	for l := range line {
		fileArray = append(fileArray, l)
	}

	if len(expectedArray) != len(fileArray) {
		t.Errorf("File does not match expected.")
		t.Logf("Expected: %d", len(expectedArray))
		t.Logf("Actual: %d", len(fileArray))
	}

	for i, j := range expectedArray {
		if j != fileArray[i] {
			t.Errorf("File does not match expected.")
			t.Logf("Expected: %v", expectedArray)
			t.Logf("Actual: %v", fileArray)
		}
	}
}

func TestUnGZip(t *testing.T) {
	filenamegz := "TestData/smallkern.log.2.gz"
	filename := filenamegz[:len(filenamegz)-3]
	UnGZip(filenamegz)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Could not find unziped file")
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Log("No file to remove")
	}
	if err := os.Remove(filename); err != nil {
		t.Log("Could not remove the file")
	}
}

func TestWrite(t *testing.T) {
	filename := "TestData/TestingFile.txt"
	string1 := "This is a line in the file"
	string2 := "This is another line in the file"

	Write(filename, string1)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Testing file was not created!")
	}

	Write(filename, string2)
	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("%s", err)
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("%s", err)
	}
	expectedString := string1 + "\n" + string2 + "\n"
	if string(fileBytes) != expectedString {
		t.Errorf("Did not append string the way we expected")
		t.Logf("Expected: %s", expectedString)
		t.Logf("Actual: %s", string(fileBytes))
	}

	file.Close()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Log("No file to remove")
	}
	if err := os.Remove(filename); err != nil {
		t.Log("Could not remove the file")
	}
}

//TODO: StreamRead
