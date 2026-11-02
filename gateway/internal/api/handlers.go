package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"ledger/ledger"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	tx := TransactionFromCreateRequest(req)
	if err := tx.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := ledger.AddTransaction(tx)
	if err != nil {
		if errors.Is(err, ledger.ErrBudgetExceeded) {
			writeError(w, http.StatusConflict, "budget exceeded")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, ToTransactionResponse(created))
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	items, err := ledger.ListTransactions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	out := make([]TransactionResponse, 0, len(items))
	for _, tx := range items {
		out = append(out, ToTransactionResponse(tx))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) SetBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	b := BudgetFromCreateRequest(req)
	if err := b.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := ledger.SetBudget(b); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, ToBudgetResponse(b))
}

func (h *Handler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	items, err := ledger.ListBudgets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	out := make([]BudgetResponse, 0, len(items))
	for _, b := range items {
		out = append(out, ToBudgetResponse(b))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) ReportSummary(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if err := ledger.ValidateReportRange(from, to); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	summary, err := ledger.GetReportSummary(r.Context(), from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
