package main

import (
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	})

	log.Println("Auth service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
