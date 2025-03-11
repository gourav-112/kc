package dbs

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	Db    *sql.DB
	mutex = sync.Mutex{}
)

// Initialize Database
func InitDB() {
	var err error
	Db, err = sql.Open("sqlite3", "kirana_club.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = Db.Exec(`

		CREATE TABLE jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status TEXT NOT NULL
		);

		CREATE TABLE store_visits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_id INTEGER NOT NULL,
			store_id TEXT NOT NULL,
			FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
		);

		CREATE TABLE images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			visit_id INTEGER NOT NULL,
			image_url TEXT NOT NULL,
			status TEXT NOT NULL,
			FOREIGN KEY (visit_id) REFERENCES store_visits(id) ON DELETE CASCADE
		);
	`)

	if err != nil {
		log.Fatal("Database initialization error:", err)
	}

	log.Println("Database schema updated successfully!")
}
