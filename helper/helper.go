package helper

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gourav-112/kc/global"
	"github.com/gourav-112/kc/dbs"
	_ "github.com/mattn/go-sqlite3"
)

func ProcessJob(jobID int) {
	rows, err := dbs.Db.Query("SELECT image_url FROM images WHERE job_id = ?", jobID)
	if err != nil {
		fmt.Println("Error fetching images:", err)
		updateJobStatus(jobID, global.StatusFailed)
		return
	}
	defer rows.Close()

	var image_url []string
	for rows.Next() {
		var imageURL string
		if err := rows.Scan(&imageURL); err != nil {
			fmt.Println("Error scanning image URL:", err)
			updateJobStatus(jobID, global.StatusFailed)
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
			status := global.StatusCompleted
			if !success {
				status = global.StatusFailed
			}
			_, err := dbs.Db.Exec("UPDATE images SET status = ? WHERE job_id = ? AND image_url = ?", status, jobID, img)
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
		updateJobStatus(jobID, global.StatusCompleted)
	} else {
		updateJobStatus(jobID, global.StatusFailed)
	}
}

// Helper function to process an image (simulated)
func processImage(imageURL string) bool {
	time.Sleep(time.Duration(100+rand.Intn(300)) * time.Millisecond) // Simulate processing
	return rand.Intn(10) > 1                                         // 90% chance of success
}

// Helper function to update job status
func updateJobStatus(jobID int, status global.JobStatus) {
	_, err := dbs.Db.Exec("UPDATE jobs SET status = ? WHERE id = ?", status, jobID)
	if err != nil {
		fmt.Println("Error updating job status:", err)
	}
}
