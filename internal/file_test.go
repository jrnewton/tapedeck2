package tapedeck

import (
	"path/filepath"
	// avoid clash with local var 'db'
	dbpkg "tapedeck/internal/db"
	"testing"
)

var inputSchema = filepath.Join("./", dbpkg.DatabaseFileName)

func TestReadLines(t *testing.T) {
	err := ReadLines(inputSchema, func(line string) error {
		t.Log("line:", line)
		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
