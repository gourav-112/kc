package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type JobStatus string

const (
	StatusOngoing   JobStatus = "ongoing"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

type Job struct {
	ID     int       `json:"job_id"`
	Status JobStatus `json:"status"`
}

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

type Visit struct {
	StoreID   string   `json:"store_id"`
	image_url []string `json:"image_url"`
}

type SubmitRequest struct {
	Count  int     `json:"count"`
	Visits []Visit `json:"visits"`
}

// Submit Job API
func submitJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Count  int `json:"count"`
		Visits []struct {
			StoreID   string   `json:"store_id"`
			ImageURLs []string `json:"image_url"`
			VisitTime string   `json:"visit_time"`
		} `json:"visits"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("JSON Decode Error:", err)
		http.Error(w, `{ "error": "Invalid JSON" }`, http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Database Transaction Begin Error:", err)
		http.Error(w, `{ "error": "Database transaction failed" }`, http.StatusInternalServerError)
		return
	}

	// Insert job entry
	result, err := tx.Exec("INSERT INTO jobs (status) VALUES (?)", StatusOngoing)
	if err != nil {
		log.Println("Error inserting job:", err)
		tx.Rollback()
		http.Error(w, `{ "error": "Failed to create job" }`, http.StatusInternalServerError)
		return
	}

	jobID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error retrieving job ID:", err)
		tx.Rollback()
		http.Error(w, `{ "error": "Failed to retrieve job ID" }`, http.StatusInternalServerError)
		return
	}

	log.Printf("Created job ID: %d\n", jobID)

	// Process visits and images
	for _, visit := range req.Visits {
		result, err := tx.Exec("INSERT INTO store_visits (job_id, store_id) VALUES (?, ?)", jobID, visit.StoreID)
		if err != nil {
			log.Println("Error inserting store visit for store_id", visit.StoreID, ":", err)
			tx.Rollback()
			http.Error(w, `{ "error": "Failed to save store visit" }`, http.StatusInternalServerError)
			return
		}

		visitID, err := result.LastInsertId()
		if err != nil {
			log.Println("Error retrieving store visit ID for store_id", visit.StoreID, ":", err)
			tx.Rollback()
			http.Error(w, `{ "error": "Failed to retrieve store visit ID" }`, http.StatusInternalServerError)
			return
		}

		log.Printf("Created store visit ID: %d for store_id: %s\n", visitID, visit.StoreID)

		// Insert images linked to store visit
		for _, imageURL := range visit.ImageURLs {
			_, err := tx.Exec("INSERT INTO images (visit_id, image_url, status) VALUES (?, ?, ?)", visitID, imageURL, StatusOngoing)
			if err != nil {
				log.Println("Error inserting image for store_id", visit.StoreID, ":", err)
				tx.Rollback()
				http.Error(w, `{ "error": "Failed to save images" }`, http.StatusInternalServerError)
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Println("Database Commit Error:", err)
		http.Error(w, `{ "error": "Database commit failed" }`, http.StatusInternalServerError)
		return
	}

	// Process job asynchronously
	go processJob(int(jobID))

	// Respond with job ID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"job_id": int(jobID)})
}

// Process Job (Simulating Image Processing)
func processJob(jobID int) {
	rows, err := db.Query("SELECT image_url FROM images WHERE job_id = ?", jobID)
	if err != nil {
		fmt.Println("Error fetching images:", err)
		updateJobStatus(jobID, StatusFailed)
		return
	}
	defer rows.Close()

	var image_url []string
	for rows.Next() {
		var imageURL string
		if err := rows.Scan(&imageURL); err != nil {
			fmt.Println("Error scanning image URL:", err)
			updateJobStatus(jobID, StatusFailed)
			return
		}
		image_url = append(image_url, imageURL)
	}

	var wg sync.WaitGroup
	statusChan := make(chan bool, len(image_url))

	// Process each image concurrently
	for _, imageURL := range image_url {
		wg.Add(1)
		go func(img string) {
			defer wg.Done()

			success := processImage(img)
			statusChan <- success

			// Update image processing status
			status := StatusCompleted
			if !success {
				status = StatusFailed
			}
			_, err := db.Exec("UPDATE images SET status = ? WHERE job_id = ? AND image_url = ?", status, jobID, img)
			if err != nil {
				fmt.Println("Error updating image status:", err)
			}
		}(imageURL)
	}

	wg.Wait()
	close(statusChan)

	// Check overall job status
	allCompleted := true
	for success := range statusChan {
		if !success {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		updateJobStatus(jobID, StatusCompleted)
	} else {
		updateJobStatus(jobID, StatusFailed)
	}
}

// Helper function to process an image (simulated)
func processImage(imageURL string) bool {
	time.Sleep(time.Duration(100+rand.Intn(300)) * time.Millisecond) // Simulate processing
	return rand.Intn(10) > 1                                         // 90% chance of success
}

// Helper function to update job status
func updateJobStatus(jobID int, status JobStatus) {
	_, err := db.Exec("UPDATE jobs SET status = ? WHERE id = ?", status, jobID)
	if err != nil {
		fmt.Println("Error updating job status:", err)
	}
}

// Get Job Status API
func getJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	jobIDStr := r.URL.Query().Get("jobid")
	if jobIDStr == "" {
		http.Error(w, `{ "error": "Missing jobid" }`, http.StatusBadRequest)
		return
	}

	var job Job
	err := db.QueryRow("SELECT id, status FROM jobs WHERE id = ?", jobIDStr).Scan(&job.ID, &job.Status)
	if err != nil {
		http.Error(w, `{ "error": "Job not found" }`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(job)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	initDB()

	http.HandleFunc("/api/submit", submitJobHandler)
	http.HandleFunc("/api/status", getJobStatusHandler)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
