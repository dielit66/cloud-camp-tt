package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	var port string
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	if port == "" {
		log.Fatal("port is required")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		fmt.Fprintf(w, "response from backend server on port: %s", port)
	})

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("server has been started on port :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
