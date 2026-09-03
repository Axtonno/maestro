# Milestone 27 — Suite globale e compatibilità

Data: 2026-09-03

Stato: **PASS**

I failure registrati da M26 erano effetti del checkout Windows CRLF: i digest
congelati verificano byte LF e due test v2 applicano mutazioni testuali LF.
Nessun file funzionale, fixture o test è stato cambiato. Una materializzazione
Linux LF pulita del commit ha confermato:

- `go test ./...`: PASS;
- `go test -race ./...`: PASS;
- `go vet ./...`: PASS;
- `git diff --check`: PASS;
- package `internal/productconfig`, inclusi profili v2 legacy e schema v3:
  PASS.

La compatibilità è invariata: i profili v2 validi continuano a caricarsi senza
ricevere implicitamente i controlli generativi v3; lo schema v3 richiede e
propaga `num_predict` e residency. Campi sconosciuti, duplicati o riservati a
v3 restano fail-closed.
