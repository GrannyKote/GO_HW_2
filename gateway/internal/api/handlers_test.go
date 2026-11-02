package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ledger/ledger"
)

const jsonContentTypeHeader = "application/json; charset=utf-8"

func newTestServer() *httptest.Server {
	h := NewHandler()
	return httptest.NewServer(NewRouter(h))
}

func assertJSONContentType(t *testing.T, resp *http.Response) {
	t.Helper()
	if ct := resp.Header.Get("Content-Type"); ct != jsonContentTypeHeader {
		t.Fatalf("Content-Type: want %q, got %q", jsonContentTypeHeader, ct)
	}
}

func decodeError(t *testing.T, body io.Reader) string {
	t.Helper()
	var payload errorBody
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		t.Fatalf("decode error JSON: %v", err)
	}
	return payload.Error
}

func TestBudgetAPI(t *testing.T) {
	t.Run("POST /api/budgets", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			ledger.Reset()
			t.Cleanup(ledger.Reset)

			srv := newTestServer()
			t.Cleanup(srv.Close)

			body := `{"category":"food","limit":5000}`
			resp, err := http.Post(srv.URL+"/api/budgets", jsonContentTypeHeader, strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assertJSONContentType(t, resp)
			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("status: want 201, got %d", resp.StatusCode)
			}

			var got BudgetResponse
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Category != "food" || got.Limit != 5000 {
				t.Fatalf("unexpected budget: %+v", got)
			}

			listResp, err := http.Get(srv.URL + "/api/budgets")
			if err != nil {
				t.Fatal(err)
			}
			defer listResp.Body.Close()
			assertJSONContentType(t, listResp)

			var budgets []BudgetResponse
			if err := json.NewDecoder(listResp.Body).Decode(&budgets); err != nil {
				t.Fatalf("decode list: %v", err)
			}
			found := false
			for _, b := range budgets {
				if b.Category == "food" && b.Limit == 5000 {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("budget not in list: %+v", budgets)
			}
		})

		t.Run("bad_limit", func(t *testing.T) {
			ledger.Reset()
			t.Cleanup(ledger.Reset)

			srv := newTestServer()
			t.Cleanup(srv.Close)

			body := `{"category":"food","limit":0}`
			resp, err := http.Post(srv.URL+"/api/budgets", jsonContentTypeHeader, strings.NewReader(body))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assertJSONContentType(t, resp)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status: want 400, got %d", resp.StatusCode)
			}
			if msg := decodeError(t, resp.Body); msg == "" {
				t.Fatal("expected non-empty error message")
			}
		})

		t.Run("bad_json", func(t *testing.T) {
			ledger.Reset()
			t.Cleanup(ledger.Reset)

			srv := newTestServer()
			t.Cleanup(srv.Close)

			resp, err := http.Post(srv.URL+"/api/budgets", jsonContentTypeHeader, strings.NewReader("{not json"))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assertJSONContentType(t, resp)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status: want 400, got %d", resp.StatusCode)
			}
			if got := decodeError(t, resp.Body); got != "invalid JSON body" {
				t.Fatalf("error: want %q, got %q", "invalid JSON body", got)
			}
		})
	})
}

func TestTransactionAPI(t *testing.T) {
	t.Run("flow", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			ledger.Reset()
			t.Cleanup(ledger.Reset)

			srv := newTestServer()
			t.Cleanup(srv.Close)

			budgetBody := `{"category":"food","limit":5000}`
			budgetResp, err := http.Post(srv.URL+"/api/budgets", jsonContentTypeHeader, strings.NewReader(budgetBody))
			if err != nil {
				t.Fatal(err)
			}
			budgetResp.Body.Close()
			if budgetResp.StatusCode != http.StatusCreated {
				t.Fatalf("create budget: want 201, got %d", budgetResp.StatusCode)
			}

			txBody := `{"amount":450,"category":"food","description":"lunch","date":"2025-09-10"}`
			txResp, err := http.Post(srv.URL+"/api/transactions", jsonContentTypeHeader, strings.NewReader(txBody))
			if err != nil {
				t.Fatal(err)
			}
			defer txResp.Body.Close()

			assertJSONContentType(t, txResp)
			if txResp.StatusCode != http.StatusCreated {
				t.Fatalf("status: want 201, got %d", txResp.StatusCode)
			}

			var created TransactionResponse
			if err := json.NewDecoder(txResp.Body).Decode(&created); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if created.ID != 1 || created.Amount != 450 || created.Category != "food" {
				t.Fatalf("unexpected transaction: %+v", created)
			}
			if created.Date != "2025-09-10" {
				t.Fatalf("date: want 2025-09-10, got %q", created.Date)
			}

			listResp, err := http.Get(srv.URL + "/api/transactions")
			if err != nil {
				t.Fatal(err)
			}
			defer listResp.Body.Close()

			assertJSONContentType(t, listResp)
			if listResp.StatusCode != http.StatusOK {
				t.Fatalf("list status: want 200, got %d", listResp.StatusCode)
			}

			var list []TransactionResponse
			if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
				t.Fatalf("decode list: %v", err)
			}
			if len(list) != 1 || list[0].ID != created.ID {
				t.Fatalf("unexpected list: %+v", list)
			}
		})

		t.Run("exceeded", func(t *testing.T) {
			ledger.Reset()
			t.Cleanup(ledger.Reset)

			srv := newTestServer()
			t.Cleanup(srv.Close)

			_, _ = http.Post(srv.URL+"/api/budgets", jsonContentTypeHeader,
				strings.NewReader(`{"category":"food","limit":500}`))
			_, _ = http.Post(srv.URL+"/api/transactions", jsonContentTypeHeader,
				strings.NewReader(`{"amount":400,"category":"food","description":"first","date":"2025-09-10"}`))

			overBody := `{"amount":200,"category":"food","description":"over","date":"2025-09-11"}`
			resp, err := http.Post(srv.URL+"/api/transactions", jsonContentTypeHeader, strings.NewReader(overBody))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assertJSONContentType(t, resp)
			if resp.StatusCode != http.StatusConflict {
				t.Fatalf("status: want 409, got %d", resp.StatusCode)
			}
			if got := decodeError(t, resp.Body); got != "budget exceeded" {
				t.Fatalf("error: want %q, got %q", "budget exceeded", got)
			}

			list := ledger.ListTransactions()
			if len(list) != 1 {
				t.Fatalf("transaction count after reject: want 1, got %d", len(list))
			}
		})

		t.Run("bad_json", func(t *testing.T) {
			ledger.Reset()
			t.Cleanup(ledger.Reset)

			srv := newTestServer()
			t.Cleanup(srv.Close)

			req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader([]byte("{")))
			req.Header.Set("Content-Type", jsonContentTypeHeader)
			rec := httptest.NewRecorder()
			NewHandler().CreateTransaction(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			assertJSONContentType(t, res)
			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("status: want 400, got %d", res.StatusCode)
			}
			if got := decodeError(t, res.Body); got != "invalid JSON body" {
				t.Fatalf("error: want %q, got %q", "invalid JSON body", got)
			}
		})
	})
}
