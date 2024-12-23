package tapedeck

import (
	"fmt"
	"log"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Database struct {
	FilePath string
}

func (db *Database) String() string {
	return fmt.Sprintf("database %v", db.FilePath)
}

type Query struct {
	Name         string
	SingleResult bool
	Sql          string
	Named        map[string]any
	ResultFunc   func(stmt *sqlite.Stmt) error
}

func (query *Query) String() string {
	var single = ""
	if query.SingleResult {
		single = "single result expected"
	} else {
		single = "multiple results expected"
	}

	return fmt.Sprintf("query %v %v (%v)", query.Name, query.Sql, single)
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

func (db Database) TestConnection() (err error) {
	log.Println("enter TestConnection", db)
	defer log.Println("exit TestConnection")

	log.Println("open connection")
	conn, err := sqlite.OpenConn(db.FilePath, sqlite.OpenReadWrite)
	if err != nil {
		log.Println("failed to open connection!", err)
		return err
	}

	defer func() { err = closeConn(conn, err) }()
	return
}

func (db Database) RunQuery(query Query) (err error) {
	log.Println("enter RunQuery", query.Name)
	defer log.Println("exit RunQuery", query.Name)

	log.Println("open connection")
	conn, err := sqlite.OpenConn(db.FilePath, sqlite.OpenReadWrite)
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
