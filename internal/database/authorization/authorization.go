package authorization

import (
	"fmt"
	"log"
	"tapedeck/internal/database"
	"time"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
)

const (
	SchemaVersion = 1
)

const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"
)

type User struct {
	Id          int64
	Uuid        string
	Email       string
	ExternalId  string
	Provider    string
	Status      string
	DateCreated string
}

func (u *User) String() string {
	return fmt.Sprintf("user %v", u.Email)
}

func NewUser(email string) User {
	user := User{
		Uuid:        uuid.New().String(),
		Email:       email,
		Provider:    "google",
		Status:      StatusEnabled,
		DateCreated: time.Now().Format(time.RFC3339),
	}

	return user
}

func userCreator(stmt *sqlite.Stmt) (User, error) {
	return User{
		Id:          stmt.GetInt64("ID"),
		Uuid:        stmt.GetText("UUID"),
		Email:       stmt.GetText("EMAIL"),
		Provider:    stmt.GetText("PROVIDER"),
		ExternalId:  stmt.GetText("EXTERNAL_ID"),
		Status:      stmt.GetText("STATUS"),
		DateCreated: stmt.GetText("DATE_CREATED"),
	}, nil
}

func InsertUser(db *database.Database, user User) (int, error) {
	log.Println("enter InsertUser")
	defer log.Println("exit InsertUser")

	recordCount := 0
	err := db.RunQuery(database.Query{
		Name:           "InsertUser",
		Sql:            "INSERT INTO USER (UUID, EMAIL, PROVIDER, STATUS, DATE_CREATED) VALUES(:uuid, :email, :provider, :status, :created);",
		PerformsUpdate: true,
		Named: map[string]any{
			":uuid":     user.Uuid,
			":email":    user.Email,
			":provider": user.Provider,
			":status":   user.Status,
			":created":  user.DateCreated,
		},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			log.Println("user inserted")
			recordCount++
			return nil
		},
	})

	return recordCount, err
}

func GetUserByEmail(db *database.Database, email string) (*User, error) {
	log.Println("enter GetUserByEmail")
	defer log.Println("exit GetUserByEmail")

	var user User
	err := db.RunQuery(database.Query{
		Name:           "GetUserByEmail",
		Sql:            "SELECT * FROM USER WHERE EMAIL=:email;",
		Named:          map[string]any{":email": email},
		PerformsUpdate: false,
		ResultFunc: func(stmt *sqlite.Stmt) error {
			user, err := userCreator(stmt)
			if err != nil {
				log.Println("user returned", user)
			}
			return err
		},
	})

	return &user, err
}
