# Milestone 7 — Phase 5 Report

Stato: Completata

Data: 2026-08-12

---

# Risultato

La Fase 5 collega sessione, Provider Runtime e Tool Runtime in un loop
sequenziale, cancellabile e bounded. Un piano valido può ora attraversare
risposte modello, tool call autorizzate e risultato finale fino al terminale
`completed`.

# Capacità consegnate

- adapter dei descriptor tool verso `provider.Tool`;
- selezione esatta di provider, modello e tool set;
- conversazioni con ruoli distinti e correlazione call/result;
- call ID provider o fallback deterministico;
- call multiple eseguite in ordine;
- risultati tool JSON tipizzati, inclusi deny e truncation;
- deny recoverable restituito al modello e deny terminale definitivo;
- assembler streaming bounded per testo e delta tool;
- semantica terminale condivisa da Complete e Stream;
- accounting prima di ogni nuova iterazione;
- propagazione di cancellazione, deadline e failure.

# Invarianti verificati

- tool sconosciuti o non inclusi non raggiungono `Prepare`;
- arguments non-oggetto non raggiungono il tool;
- call duplicate non vengono eseguite;
- call multiple non vengono parallelizzate;
- un risultato conserva ID e nome della call originaria;
- un deny non viene convertito in allow;
- failure mid-stream chiude lo stream e non usa output parziale;
- il loop non bypassa registry, policy, permit o limiti.

# Test e gate

```text
GOCACHE=/tmp/maestro-go-build go test ./internal/agent ./pkg/agent ./internal/tool ./pkg/tool
GOCACHE=/tmp/maestro-go-build go test -race ./internal/agent ./pkg/agent ./internal/tool ./pkg/tool
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

# Gate

Il loop completo è dimostrato con provider e tool in-memory, senza rete. La
Fase 6 può aggiungere binding workspace, reference tool filesystem e freshness
del contesto senza cambiare la semantica delle call.
