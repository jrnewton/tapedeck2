package database

import (
	"fmt"
	"log"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const DatabaseFileName = "tapedeck.db"

type Database struct {
	FilePath string
	// schemaVersion is read on startup from the VERSION table in each database
	schemaVersion int64
}

func (db *Database) String() string {
	return fmt.Sprintf("database %v", db.FilePath)
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

// Close the connection and return either the current error
// or if the error that occurred during close. If both errors
// exist then return current which represents the first problem
// and should not be masked by the subsequent close failure.
//
// This function is meant to be called in a defer block in a func
// that is using a named return value for error.
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

func (db Database) UpgradeCheck(expectedVersion int64) (upgrade bool, err error) {
	log.Println("enter UpgradeCheck", db)
	defer log.Println("exit UpgradeCheck")

	var actualVersion int64 = -1
	err = db.RunQuery(Query{
		Name:           "GetVersion",
		Sql:            "SELECT NUMBER FROM VERSION;",
		PerformsUpdate: false,
		ResultFunc: func(stmt *sqlite.Stmt) error {
			actualVersion = stmt.GetInt64("NUMBER")
			return nil
		},
	})

	db.schemaVersion = actualVersion

	if actualVersion <= 0 {
		log.Panicf("db schema version (%v) <= 0", actualVersion)
	} else if actualVersion > expectedVersion {
		log.Panicf("db schema version (%v) > expected version (%v)", actualVersion, expectedVersion)
	} else if actualVersion < expectedVersion {
		upgrade = true
		log.Printf("db schema upgrade required, %v < %v\n", actualVersion, expectedVersion)
	} else {
		upgrade = false
		log.Printf("db schema upgrade not required, %v = %v\n", actualVersion, expectedVersion)
	}

	return
}

// TODO: follow the 12 steps
// https://sqlite.org/lang_altertable.html#making_other_kinds_of_table_schema_changes
func (db Database) Upgrade() {
	log.Println("enter Upgrade", db)
	defer log.Println("exit Upgrade")

	log.Panicf("upgrade not implemented yet")
}

func (db Database) RunQuery(query Query) (err error) {
	log.Println("enter RunQuery", query.Name)
	defer log.Println("exit RunQuery", query.Name)

	log.Println("open connection")
	var flags sqlite.OpenFlags
	if query.PerformsUpdate {
		flags = sqlite.OpenReadWrite
	} else {
		flags = sqlite.OpenReadOnly
	}

	conn, err := sqlite.OpenConn(db.FilePath, flags)
	if err != nil {
		log.Println("failed to open connection!", err)
		return
	}

	defer func() { err = closeConn(conn, err) }()

	log.Println("execute query", query.Sql)

	err = sqlitex.Execute(conn, query.Sql, &sqlitex.ExecOptions{
		Named:      query.Named,
		ResultFunc: query.ResultFunc,
	})

	if err != nil {
		log.Println("execute query error", err)
	}

	return
}
