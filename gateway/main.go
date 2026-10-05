package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	})

	addr := ":8080"
	log.Printf("Gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
