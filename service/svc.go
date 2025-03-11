package service

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gourav-112/kc/global"
	"github.com/gourav-112/kc/helper"
	"github.com/gourav-112/kc/dbs"
	_ "github.com/mattn/go-sqlite3"
)

// Submit Job API
func SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
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
	tx, err := dbs.Db.Begin()
	if err != nil {
		log.Println("Database Transaction Begin Error:", err)
		http.Error(w, `{ "error": "Database transaction failed" }`, http.StatusInternalServerError)
		return
	}

	// Insert job entry
	result, err := tx.Exec("INSERT INTO jobs (status) VALUES (?)", global.StatusOngoing)
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
			_, err := tx.Exec("INSERT INTO images (visit_id, image_url, status) VALUES (?, ?, ?)", visitID, imageURL, global.StatusOngoing)
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
	go helper.ProcessJob(int(jobID))

	// Respond with job ID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"job_id": int(jobID)})
}

// Get Job Status API
func GetJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	jobIDStr := r.URL.Query().Get("jobid")
	if jobIDStr == "" {
		http.Error(w, `{ "error": "Missing jobid" }`, http.StatusBadRequest)
		return
	}

	var job global.Job
	err := dbs.Db.QueryRow("SELECT id, status FROM jobs WHERE id = ?", jobIDStr).Scan(&job.ID, &job.Status)
	if err != nil {
		http.Error(w, `{ "error": "Job not found" }`, http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(job)
}
