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
	flag.StringVar(&dbFile, "dbFile", "", "path to SQLite database file")

	var action string
	flag.StringVar(&action, "action", "", "Action to run, possible values: user-add, user-delete")

	var email string
	flag.StringVar(&email, "email", "", "User's email address")

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

	db.Open(true)

	switch action {
	case "user-add":
		{
			if email == "" {
				fmt.Println("email required")
				flag.Usage()
				return
			}

			u := user.New(email)
			err := user.Insert(db, u)
			if err != nil {
				panic(err)
			}
		}
	case "user-delete":
		{
			panic("delete-user is not implemented")
		}
	default:
		{
			fmt.Printf("Unknown action value '%s'\n", action)
			flag.Usage()
			return
		}
	}

}
