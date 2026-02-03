package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Backend2:9002] Request: %s %s", r.Method, r.URL.Path)
		response := fmt.Sprintf("Response from Backend 2 (port 9002)\nTime: %s\nPath: %s\n",
			time.Now().Format("15:04:05"), r.URL.Path)
		fmt.Fprint(w, response)
	})

	log.Println("[Backend2] Starting on :9002")
	http.ListenAndServe(":9002", nil)
}
