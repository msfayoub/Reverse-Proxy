package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Backend1:9001] Request: %s %s", r.Method, r.URL.Path)
		response := fmt.Sprintf("Response from Backend 1 (port 9001)\nTime: %s\nPath: %s\n",
			time.Now().Format("15:04:05"), r.URL.Path)
		fmt.Fprint(w, response)
	})

	log.Println("[Backend1] Starting on :9001")
	http.ListenAndServe(":9001", nil)
}
