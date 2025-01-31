package database_test

import (
	"os"
	"tapedeck/internal/database"
	"testing"
)

const dbPath = "./unit-test.db"

func TestDbExists(t *testing.T) {
	//setup, verify db file is not there
	_, err := os.Stat(dbPath)
	if err == nil {
		t.Fatalf("database file should not exist\n")
	}

	//make new and open which will create the file
	db := database.New(dbPath)
	db.Open(false)

	//verify file is there
	_, err = os.Stat(dbPath)
	if err != nil {
		t.Fatalf("database file not created by Open()\n")
	}

	//close db with file deletion
	db.Close(true)

	_, err = os.Stat(dbPath)
	if err == nil {
		t.Fatalf("database file was not removed after Close()\n")
	}
}
