module gateway

go 1.27.1

require (
	github.com/go-chi/chi/v5 v5.2.3
	ledger v0.0.0
)

replace ledger => ../ledger
