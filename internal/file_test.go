package tapedeck

import (
	"testing"
)

func TestReadLines(t *testing.T) {
	err := ReadLines("./main.schema.sql", func(line string) error {
		t.Log("line:", line)
		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
