package global

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

type Visit struct {
	StoreID   string   `json:"store_id"`
	image_url []string `json:"image_url"`
}

type SubmitRequest struct {
	Count  int     `json:"count"`
	Visits []Visit `json:"visits"`
}