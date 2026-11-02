package main

import (
	"log"
	"net/http"

	"gateway/internal/api"
)

func main() {
	h := api.NewHandler()
	addr := ":8080"
	log.Printf("Gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, api.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
