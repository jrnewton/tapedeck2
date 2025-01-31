package tapedeck

import (
	"path/filepath"
	"tapedeck/internal/database"
	"testing"
)

var inputSchema = filepath.Join("./", database.DatabaseFileName)

func TestReadLines(t *testing.T) {
	err := ReadLines(inputSchema, func(line string) error {
		t.Log("line:", line)
		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
