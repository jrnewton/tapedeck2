package tape_test

import (
	"fmt"
	"log"
	"strings"
	tapedeck "tapedeck/internal"
	"tapedeck/internal/database"
	"tapedeck/internal/database/authorization"
	"testing"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const testEmail = "tapedeck.us@gmail.com"
const testDatabase = "./unit-test.db"

func setup(t *testing.T) *database.Database {
	var db *database.Database
	var createTable string

	err := tapedeck.ReadLines("./main.schema.sql", func(line string) error {
		foundIt := strings.HasPrefix(line, "CREATE TABLE USER ")
		if foundIt {
			db = &database.Database{FilePath: testDatabase}
			createTable = line
			return nil
		}

		return fmt.Errorf("USER table not found")
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { teardown(t, db) })

	conn, err := sqlite.OpenConn(db.FilePath, sqlite.OpenCreate, sqlite.OpenReadWrite)
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	// create table
	err = sqlitex.ExecuteTransient(conn, createTable, &sqlitex.ExecOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// insert dummy data
	user := authorization.NewUser(testEmail)

	_, err = authorization.InsertUser(db, user)

	if err != nil {
		t.Fatal(err)
	}

	// if count == 0 {
	// 	t.Fatalf("insert user produced 0 rows, expected 1 row\n")
	// }

	return db
}

func teardown(t *testing.T, db *database.Database) {
	if db != nil {
		err := tapedeck.DeleteFile(db.FilePath)
		if err != nil {
			t.Fatal(err)
			return
		}
	}
}

func TestGetUserByEmail(t *testing.T) {
	db := setup(t)

	user, err := authorization.GetUserByEmail(db, testEmail)

	if err != nil {
		t.Fatal(err)
	}

	if user == nil {
		t.Fatalf("user not found %s", testEmail)
	}

	if user.Email != testEmail {
		t.Fatalf("user record has mismatched email: expected %s, actual %s", testEmail, user.Email)
	}

	log.Println("User:", user)
}
