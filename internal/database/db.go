// avoid clash with local var 'db'
package database

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
	"tapedeck/internal/file"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitemigration"
	"zombiezen.com/go/sqlite/sqlitex"
)

const DatabaseFileName = "tapedeck.db"

// This contains the full database schema including
// additional changes for incremental upgrades.
//
//go:embed schema.sql
var sql string

type State int

const (
	new State = iota
	opened
	closed
)

type Database struct {
	filePath   string
	schemaData string
	state      State
}

func (db *Database) String() string {
	return fmt.Sprintf("database %q %v", db.filePath, db.state)
}

func New(filePath string) *Database {
	return &Database{
		filePath:   filePath,
		schemaData: sql,
		state:      new,
	}
}

func (db *Database) Open() {
	db.open(false)
}

// open will prepare the database for use by checking
// for an existing file and performing any schema upgrades.
// It will return upon success or panic when an error occurs.
func (db *Database) open(create bool) {
	log.Println("enter open", db, create)
	defer log.Println("exit open")

	if db.state != new {
		log.Panicf("database not in state new: %v\n", db)
	}

	_, err := os.Stat(db.filePath)
	if err != nil {
		if create {
			log.Printf("creating database file\n")
			_, err = file.Touch(db.filePath)
			if err != nil {
				log.Panicf("could not create database file at %v, %v\n", db.filePath, err)
			}
		} else {
			log.Panicf("database file not found: %v\n", err)
		}
	}
	db.state = opened
}

// print the internal schema table or panic
func schemaReport(conn *sqlite.Conn) {
	const listSchemaQuery = `SELECT TYPE, NAME FROM sqlite_schema ORDER BY 1, 2;`
	err := sqlitex.ExecuteTransient(conn, listSchemaQuery, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			log.Printf("  %-5s %s", stmt.ColumnText(0), stmt.ColumnText(1))
			return nil
		},
	})

	if err != nil {
		log.Panicf("sqlite_schema query failed: %v\n", err)
	}
}

// upgrade performs a database upgrade and returns on success or panic
func (db *Database) Upgrade() {
	log.Println("enter upgrade", db)
	defer log.Println("exit upgrade")

	if db.state != opened {
		log.Panicf("database not in state open: %v\n", db)
	}

	rawLines := strings.Split(db.schemaData, "\n")

	// remove SQL comments
	schemaLines := make([]string, 0)
	for _, line := range rawLines {
		if !strings.HasPrefix(line, "--") {
			schemaLines = append(schemaLines, line)
		}
	}

	log.Printf("schema lines: %d\n", len(schemaLines))

	if len(schemaLines) == 0 {
		panic("missing schema data")
	}

	schema := sqlitemigration.Schema{
		Migrations: schemaLines,
	}

	conn, err := openConn(db, true)
	if err != nil {
		log.Panicf("upgrade failed, openConn: %v\n", err)
	}

	defer func() {
		deferErr := closeConn(conn, err)
		if deferErr != nil {
			panic(deferErr)
		}
	}()

	log.Printf("schema before upgrade:\n")
	schemaReport(conn)

	err = sqlitemigration.Migrate(context.TODO(), conn, schema)
	if err != nil {
		log.Panicf("upgrade failed, migration: %v\n", err)
	}

	log.Printf("schema after upgrade:\n")
	schemaReport(conn)
}

func (db *Database) Close() {
	db.close(false)
}

func (db *Database) close(delete bool) {
	log.Println("enter close", db, delete)
	defer log.Println("exit close")

	if db.state != opened {
		log.Panicf("database not in state open: %v\n", db)
	}

	if delete {
		err := file.Delete(db.filePath)
		if err != nil {
			panic(err)
		}
	}
	db.state = closed
}

func (db *Database) RunQuery(query Query) (err error) {
	log.Println("enter RunQuery", query.Name)
	defer log.Println("exit RunQuery", query.Name)

	conn, err := openConn(db, query.PerformsUpdate)
	if err != nil {
		return
	}
	defer func() {
		err = closeConn(conn, err)
	}()

	log.Println("execute query", query.Sql, query.Named)

	err = sqlitex.Execute(conn, query.Sql, &sqlitex.ExecOptions{
		Named:      query.Named,
		ResultFunc: query.ResultFunc,
	})

	if err != nil {
		log.Println("execute query error", err)
	}

	return
}

type Query struct {
	Name           string
	Sql            string
	PerformsUpdate bool
	Named          map[string]any
	ResultFunc     func(stmt *sqlite.Stmt) error
}

func (query *Query) String() string {
	return fmt.Sprintf("query %v %v", query.Name, query.Sql)
}

func openConn(db *Database, update bool) (*sqlite.Conn, error) {
	log.Println("enter openConn")
	defer log.Println("exit openConn")

	var flags sqlite.OpenFlags
	if update {
		flags = sqlite.OpenReadWrite
	} else {
		flags = sqlite.OpenReadOnly
	}

	conn, err := sqlite.OpenConn(db.filePath, flags)
	if err != nil {
		log.Println("failed to open connection!", err)
		return nil, err
	}

	return conn, nil
}

// Close the connection and return either the current error
// or the error that occurred during close. If both errors
// exist then return current which represents the first problem
// and should not be masked by the subsequent close failure.
//
// This function is structured in this way so it can be easily
// called from a defer block in a function that is using a named
// return value for error.  That allows the error to pass back to
// caller via defer.
//
//	func doSomething() (err error) {
//	  con := . . .
//	  defer func() { err = closeConn(conn, err) }()
//	  . . .
//	  return // using named 'err'
//	}
func closeConn(conn *sqlite.Conn, currentErr error) error {
	log.Println("enter closeConn")
	defer log.Println("exit closeConn")

	closeErr := conn.Close()
	if closeErr != nil {
		log.Println("failed to close connection!", closeErr)
		if currentErr == nil {
			return closeErr
		}
	}

	return currentErr
}
