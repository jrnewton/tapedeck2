package user

import (
	"fmt"
	"log"
	"tapedeck/internal/database"

	"time"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
)

const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"
)

type User struct {
	Id         int64
	Uuid       string
	Email      string
	ExternalId string
	Provider   string
	Status     string
	Created    string
}

func (u *User) String() string {
	return fmt.Sprintf("user %q %q", u.Uuid, u.Email)
}

func New(email string) User {
	user := User{
		Uuid:     uuid.New().String(),
		Email:    email,
		Provider: "google",
		Status:   StatusEnabled,
		Created:  time.Now().Format(time.RFC3339),
	}

	return user
}

func userCreator(stmt *sqlite.Stmt) (*User, error) {
	return &User{
		Id:         stmt.GetInt64("ID"),
		Uuid:       stmt.GetText("UUID"),
		Email:      stmt.GetText("EMAIL"),
		Provider:   stmt.GetText("PROVIDER"),
		ExternalId: stmt.GetText("EXTERNAL_ID"),
		Status:     stmt.GetText("STATUS"),
		Created:    stmt.GetText("CREATED_AT"),
	}, nil
}

func Insert(db *database.Database, user User) error {
	log.Println("enter InsertUser")
	defer log.Println("exit InsertUser")

	err := db.RunQuery(database.Query{
		Name:           "InsertUser",
		Sql:            "INSERT INTO USER (UUID, EMAIL, PROVIDER, STATUS, CREATED_AT) VALUES(:uuid, :email, :provider, :status, :created);",
		PerformsUpdate: true,
		Named: map[string]any{
			":uuid":     user.Uuid,
			":email":    user.Email,
			":provider": user.Provider,
			":status":   user.Status,
			":created":  user.Created,
		},
	})

	return err
}

func GetByEmail(db *database.Database, email string) (*User, error) {
	log.Println("enter GetUserByEmail")
	defer log.Println("exit GetUserByEmail")

	var user *User
	err := db.RunQuery(database.Query{
		Name:           "GetUserByEmail",
		Sql:            "SELECT * FROM USER WHERE EMAIL=:email;",
		Named:          map[string]any{":email": email},
		PerformsUpdate: false,
		ResultFunc: func(stmt *sqlite.Stmt) error {
			u, err := userCreator(stmt)
			if err == nil {
				log.Println("user returned", u)
				user = u
			}
			return err
		},
	})

	return user, err
}
