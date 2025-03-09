package user_test

import (
	"log"
	"tapedeck/internal/database"
	"tapedeck/internal/database/user"
	"tapedeck/internal/file"
	"testing"
)

const testEmail = "tapedeck.us@gmail.com"

const dbPath = "./unit-test.db"

func setup(t *testing.T) *database.Database {
	file.Touch(dbPath)
	db := database.New(dbPath)
	t.Cleanup(func() { teardown() })

	db.Open()
	db.Upgrade()

	// insert dummy data
	u := user.New(testEmail)

	err := user.Insert(db, u)

	if err != nil {
		t.Fatal(err)
	}

	return db
}

func teardown() {
	file.Delete(dbPath)
}

func TestGetUserByEmail(t *testing.T) {
	db := setup(t)

	u, err := user.GetByEmail(db, testEmail)

	if err != nil {
		t.Fatal(err)
	}

	if u == nil {
		t.Fatalf("user not found %s", testEmail)
	}

	log.Println("User:", u)

	if u.Email != testEmail {
		t.Fatalf("user record has mismatched email: expected '%s', actual '%s'", testEmail, u.Email)
	}
}
