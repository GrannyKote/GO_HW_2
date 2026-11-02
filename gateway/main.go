package main

import (
	"context"
	"log"
	"net/http"

	"gateway/internal/api"
	"ledger/ledger"
)

func main() {
	ctx := context.Background()
	if err := ledger.Init(ctx); err != nil {
		log.Fatalf("init ledger storage: %v", err)
	}
	defer ledger.Close()

	h := api.NewHandler()
	addr := ":8080"
	log.Printf("Gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, api.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
