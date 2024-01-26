package model

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type Archiver struct {
	status   string
	progress float32
	name     string
}

func NewArchiver() *Archiver {
	return &Archiver{
		status:   "Waiting",
		progress: 0,
		name:     "",
	}
}

func (a *Archiver) Status() string {
	return a.status
}

func (a *Archiver) Progress() string {
	return fmt.Sprintf("%d", int(a.progress*100))
}

func (a *Archiver) Name() string {
	return a.name
}

func (a *Archiver) Reset() {
	a.status = "Waiting"
	a.progress = 0
	a.name = ""
}

func (a *Archiver) Run(db *DBClient) {
	if a.status == "Waiting" {
		a.status = "Running"
		a.progress = 0
		go a.run(db)
	}
}

func (a *Archiver) run(db *DBClient) {
	contacts := []Contact{}
	err := db.dao.DB().Select("*").From("contacts").All(&contacts)
	if err != nil {
		log.Printf("goroutine failed to read contacts: %v\n", err)
		a.status = "Waiting"
		a.progress = 0
		return
	}

	a.progress = 0.4
	time.Sleep(time.Second)

	bs, err := json.Marshal(contacts)
	if err != nil {
		log.Printf("goroutine failed to marshal contacts: %v\n", err)
		a.status = "Waiting"
		a.progress = 0
		return
	}

	a.progress = 0.8
	time.Sleep(time.Second)

	now := time.Now()
	archiveName := fmt.Sprintf("archive/Archive-%s-%s.json", now.Format("03.04.05PM"), now.Format(time.DateOnly))
	err = os.WriteFile(archiveName, bs, 0644)
	if err != nil {
		log.Printf("goroutine failed to write output file: %v\n", err)
	}
	a.name = archiveName
	a.status = "Complete"
	a.progress = 1
}
