package file_test

import (
	"tapedeck/internal/file"
	"testing"
)

const inputFile = "./testdata/input.txt"

func TestReadLines(t *testing.T) {
	err := file.ReadLines(inputFile, func(line string) error {
		t.Log("line:", line)
		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
