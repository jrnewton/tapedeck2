package tape

import (
	"fmt"
	"log"
	"tapedeck/internal/database"

	"zombiezen.com/go/sqlite"
)

const (
	SchemaVersion = 1
)

const (
	// Tape is waiting to be processed
	StatusTodo = "todo"
	// Tape is being processed
	StatusInProgress = "inprogress"
	// Tape processing is done without errors
	StatusDone = "done"
	// Tape processing failed with an error
	StatusError = "error"
)

// Tape is the digital equivalent to a physical cassette tape.
// It holds something recorded off the radio.
type Tape struct {
	Id    int64
	Title string
	Desc  string
	// Status of the processing for the tape
	Status string
	// StatusMsg provides additional details when Status is [StatusError]
	StatusMsg   string
	DateCreated string
	DateUpdated string
}

const (
	// TypeFile indicates that the source object is a physical file.
	TypeFile = "file"
	// TypeStream indicates that the source object is a streaming resource.
	TypeStream = "stream"
)

const (
	ContentTypeMp3  = "audio/mpeg"
	ContentTypeM3u  = "audio/x-mpegurl"
	ContentTypeAac  = "application/octet-stream"
	ContentTypeM3u8 = "application/x-mpegURL"
)

type TapeSource struct {
	Id     int64
	TapeId int64
	Seq    int64
	// Type defines the overall content type which is either
	// [TypeFile] or [TypeStream]
	Type string
	Url  string
	// FileExtension is ending '.xxx' for a [TypeFile] source object,
	// such as an mp3 file.  A [TypeStream] source object which also
	// contains mp3 data will not have this field populated.
	FileExtension string
	// ContentType is the value provided in the HTTP response header.
	ContentType string
	// optional, content downloaded via url.
	Content     string
	DateCreated string
	DateUpdated string
}

func (t *Tape) String() string {
	return fmt.Sprintf("tape %v %.20v", t.Id, t.Title)
}

func tapeCreator(stmt *sqlite.Stmt) (*Tape, error) {
	return &Tape{
		Id:          stmt.GetInt64("ID"),
		Title:       stmt.GetText("TITLE"),
		Desc:        stmt.GetText("DESC"),
		Status:      stmt.GetText("STATUS"),
		StatusMsg:   stmt.GetText("STATUS_MSG"),
		DateCreated: stmt.GetText("DATE_CREATED"),
		DateUpdated: stmt.GetText("DATE_UPDATED"),
	}, nil
}

func GetAllTapes(db *database.Database) ([]*Tape, error) {
	log.Println("enter GetAllTapes")
	defer log.Println("exit GetAllTapes")

	tapes := make([]*Tape, 0)
	err := db.RunQuery(database.Query{
		Name:           "GetAllTapes",
		Sql:            "SELECT * FROM TAPE;",
		Named:          nil,
		PerformsUpdate: false,
		ResultFunc: func(stmt *sqlite.Stmt) error {
			tape, err := tapeCreator(stmt)
			if err != nil {
				tapes = append(tapes, tape)
				log.Println("tape added", tape)
			}
			return err
		},
	})

	return tapes, err
}

func GetTape(db *database.Database, id int) (*Tape, error) {
	log.Println("enter GetTape", id)
	defer log.Println("exit GetTape", id)

	var tape *Tape
	err := db.RunQuery(database.Query{
		Name:           "GetTape",
		Sql:            "SELECT * FROM TAPE WHERE ID=:id;",
		PerformsUpdate: false,
		Named:          map[string]any{":id": id},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			tape, err := tapeCreator(stmt)
			if err != nil {
				log.Println("tape returned", tape)
			}
			return err
		},
	})

	return tape, err
}

func RecordTape() {
	log.Println("enter RecordTape")
	defer log.Println("exit RecordTape")
}
