// CLI to manage the main database file
package main

import (
	"flag"
	"fmt"
	"tapedeck/internal/database/authorization"
)

func main() {
	var action string
	flag.StringVar(&action, "action", "", "Action to run, possible values: add-user, delete-user")

	var email string
	flag.StringVar(&email, "email", "", "User's email address")

	flag.Parse()

	if action == "" {
		fmt.Println("action required")
		flag.Usage()
		return
	}

	if email == "" {
		fmt.Println("email required")
		flag.Usage()
		return
	}

	switch action {
	case "add-user":
		{
			user := authorization.NewUser(email)
			sqlInsert := fmt.Sprintf("INSERT INTO USER (DATE_CREATED, STATUS, EMAIL, UUID) VALUES('%s','%s', '%s', '%s');",
				user.DateCreated,
				user.Status,
				user.Email,
				user.Uuid)
			fmt.Println(sqlInsert)
		}
	case "delete-user":
		{
			fmt.Println("delete-user is not implemented")
		}
	default:
		{
			fmt.Printf("Unknown action value '%s'\n", action)
			flag.Usage()
			return
		}
	}

}
