package db

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db    *sql.DB
	mutex = sync.Mutex{}
)

// Initialize Database
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "kirana_club.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		DROP TABLE IF EXISTS jobs;
		DROP TABLE IF EXISTS store_visits;
		DROP TABLE IF EXISTS images;

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
