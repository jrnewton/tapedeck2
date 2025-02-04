package database

import (
	"os"
	"tapedeck/internal/file"
	"testing"
)

const dbPath = "./unit-test.db"

func cleanup() {
	file.Delete(dbPath)
}

func TestDbCreate(t *testing.T) {
	defer cleanup()

	//setup, verify db file is not there
	_, err := os.Stat(dbPath)
	if err == nil {
		t.Fatalf("database file should not exist\n")
	}

	db := New(dbPath)
	if db.state != new {
		t.Fatalf("database not in state new\n")
	}

	db.open(true)
	if db.state != opened {
		t.Fatalf("database not in state opened\n")
	}

	_, err = os.Stat(dbPath)
	if err != nil {
		t.Fatalf("database file was not created after Open(true)\n")
	}

	//close db with file deletion
	db.close(true)
	if db.state != closed {
		t.Fatalf("database not in state closed\n")
	}

	//verify db file is gone
	_, err = os.Stat(dbPath)
	if err == nil {
		t.Fatalf("database file was not removed after Close(true)\n")
	}
}

func TestDbNoCreate(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic, but got none")
		}
	}()

	// verify db file is not there
	_, err := os.Stat(dbPath)
	if err == nil {
		t.Fatalf("database file should not exist\n")
	}

	db := New(dbPath)
	if db.state != new {
		t.Fatalf("database not in state new\n")
	}

	// should panic
	db.Open()
}
