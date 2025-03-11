package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gourav-112/kc/dbs"
	"github.com/gourav-112/kc/service"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	dbs.InitDB()

	http.HandleFunc("/api/submit", service.SubmitJobHandler)
	http.HandleFunc("/api/status", service.GetJobStatusHandler)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
