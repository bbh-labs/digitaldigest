package store

import (
	"flag"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/bbhmakerlab/debug"
)

type Entry struct {
	ID int
	Title string
	Description string
	Filepath string
	UpdatedAt time.Time
	CreatedAt time.Time
}

var db gorm.DB
var dataSource = flag.String("datasource", "user=bbh dbname=dd sslmode=disable password=Lion@123", "SQL data source")

func Init() {
	var err error

	db, err = gorm.Open("postgres", *dataSource)
	if err != nil {
		debug.Fatal(err)
	}

	db.CreateTable(&Entry{})
}
