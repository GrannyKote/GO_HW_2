package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Post("/transactions", h.CreateTransaction)
		r.Get("/transactions", h.ListTransactions)
		r.Post("/budgets", h.SetBudget)
		r.Get("/budgets", h.ListBudgets)
	})
	return LoggingMiddleware(r)
}
