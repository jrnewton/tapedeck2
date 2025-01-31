package tape_test

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	tapedeck "tapedeck/internal"

	// avoid clash with local var 'db'
	dbpkg "tapedeck/internal/db"

	// avoid clash with local variable 'user'
	userpkg "tapedeck/internal/db/user"

	"testing"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const testEmail = "tapedeck.us@gmail.com"

var testDatabase = filepath.Join("./", "unit-test.db")
var inputSchema = filepath.Join("./", dbpkg.DatabaseFileName)

func setup(t *testing.T) *dbpkg.Database {
	var db *dbpkg.Database
	var createTable string

	err := tapedeck.ReadLines(inputSchema, func(line string) error {
		foundIt := strings.HasPrefix(line, "CREATE TABLE USER ")
		if foundIt {
			db = &dbpkg.Database{FilePath: testDatabase}
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
	user := userpkg.NewUser(testEmail)

	_, err = userpkg.InsertUser(db, user)

	if err != nil {
		t.Fatal(err)
	}

	// if count == 0 {
	// 	t.Fatalf("insert user produced 0 rows, expected 1 row\n")
	// }

	return db
}

func teardown(t *testing.T, db *dbpkg.Database) {
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

	user, err := userpkg.GetUserByEmail(db, testEmail)

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
