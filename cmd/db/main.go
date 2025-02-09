// CLI to manage the main database file
package main

import (
	"flag"
	"fmt"
	"tapedeck/internal/database"
	"tapedeck/internal/database/user"
)

func main() {

	var dbFile string
	flag.StringVar(&dbFile, "dbFile", "", "Path to SQLite database file")

	var action string
	flag.StringVar(&action, "action", "", "Action to run, possible values: user-add, user-delete, upgrade")

	var email string
	flag.StringVar(&email, "email", "", "User's email address for user-xxx actions")

	flag.Parse()

	if action == "" {
		fmt.Println("action required")
		flag.Usage()
		return
	}

	if dbFile == "" {
		fmt.Println("dbFile required")
		flag.Usage()
		return
	}

	db := database.New(dbFile)

	if action == "upgrade" {
		db.Open()
		db.Upgrade()
		db.Close()
	} else if action == "user-add" {
		if email == "" {
			fmt.Println("email required")
			flag.Usage()
			return
		}

		db.Open()
		u := user.New(email)
		err := user.Insert(db, u)
		if err != nil {
			panic(err)
		}
		db.Close()
	} else if action == "user-delete" {
		panic("delete-user is not implemented")
	} else {
		fmt.Printf("Unknown action value %q\n", action)
		flag.Usage()
	}
}
