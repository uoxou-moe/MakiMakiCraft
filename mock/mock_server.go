package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
    http.HandleFunc("/mock-interruption", func(w http.ResponseWriter, r *http.Request) {
        // "中断通知あり" をシミュレート
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, `{"action": "terminate", "time": "%s"}`, time.Now().Add(2*time.Minute).Format(time.RFC3339))
        log.Println("Sent mock interruption notice.")
    })
    log.Println("Mock server starting on :8080...")
    http.ListenAndServe(":8080", nil)
}
