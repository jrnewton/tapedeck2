package tapedeck

import (
	"fmt"
	"log"

	"zombiezen.com/go/sqlite"
)

type Tape struct {
	Id     int
	Title  string
	Desc   string
	Status string
}

func (ts *Tape) String() string {
	return fmt.Sprintf("tape %v %.20v", ts.Id, ts.Title)
}

func GetAllTapes(db *Database) ([]*Tape, error) {
	log.Println("enter GetAllTapes")
	defer log.Println("exit GetAllTapes")

	tapes := make([]*Tape, 0)
	err := db.RunQuery(Query{
		Name:         "GetAllTapes",
		SingleResult: true,
		Sql:          "SELECT * FROM tape;",
		Named:        nil,
		ResultFunc: func(stmt *sqlite.Stmt) error {
			tape := &Tape{
				Id:     stmt.ColumnInt(0),
				Title:  stmt.ColumnText(1),
				Desc:   stmt.ColumnText(2),
				Status: stmt.ColumnText(3),
			}
			tapes = append(tapes, tape)
			log.Println("tape added", tape)
			return nil
		},
	})

	return tapes, err
}

func GetTape(db *Database, id int) (*Tape, error) {
	log.Println("enter GetTape", id)
	defer log.Println("exit GetTape", id)

	var tape Tape
	err := db.RunQuery(Query{
		Name:         "GetTape",
		SingleResult: false,
		Sql:          "SELECT * FROM tape WHERE id=:id;",
		Named:        map[string]any{":id": id},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			tape = Tape{
				Id:     stmt.ColumnInt(0),
				Title:  stmt.ColumnText(1),
				Desc:   stmt.ColumnText(2),
				Status: stmt.ColumnText(3),
			}
			log.Println("tape returned", tape)
			return nil
		},
	})

	return &tape, err
}

func RecordTape() {
	log.Println("enter RecordTape")
	defer log.Println("exit RecordTape")
}
