# Milestone 8 — Fase 3 Operational Experience Report

Data: 2026-08-13

Stato: Completata

---

# Esito

Il gate della Fase 3 è superato. `maestro run` espone ora controllo
interattivo fail-closed, progresso redatto, cancellazione e un terminale finale
distinto dal risultato. L'incremento non modifica i contratti core né amplia
l'authority del runtime.

# Implementazione

- `internal/application/approval.go` implementa un Approver terminale
  cancellabile con deny, one-shot e grant exact-action per il run;
- `internal/application/progress.go` rende limiti ed eventi pubblici allowlist;
- `Application.ExecuteWithOptions` collega l'Approver alla `RunRequest` senza
  creare permit o bypassare Tool Runtime;
- `maestro run` distingue stdout stabile da stderr operativo, rileva il TTY,
  conserva stdin per approval e redige gli errori di esecuzione;
- SIGINT/SIGTERM usano il context condiviso e lo shutdown resta bounded a 30
  secondi.

# Proprietà verificate

- approve one-shot e run-scoped restituiscono scope distinti;
- deny esplicito/default, EOF, input invalido e no-TTY falliscono chiusi;
- una policy `prompt` senza TTY termina al confine CLI con exit code 3;
- cancellation e deadline interrompono un input bloccato;
- SIGINT/cancellazione del comando termina al confine CLI con exit code 130;
- una patch Laravel preparata viene mostrata come action logica, approvata e
  applicata attraversando Agent e Tool Runtime;
- disclosure fingerprint, invocation arguments e instruction non compaiono
  nell'output operativo;
- piano, step, permission, contatori, limiti e terminale sono line-oriented;
- un writer in panic non corrompe né interrompe il runtime;
- i test core preesistenti continuano a coprire fingerprint exact-match,
  replay/grant, limiti, terminali, cancellazione di model/tool call e observer
  best-effort.

# Gate eseguiti

```text
go test -count=3 ./...
go test -race ./...
go vet ./...
go test -run '^$' -bench BenchmarkAgentLoopDeterministic -benchmem ./internal/agent
git diff --check
```

Tutti i gate sono verdi. Le verifiche sono deterministiche e non richiedono
rete o provider live.

# Limiti residui

- l'approval governa l'orchestrazione ma non costituisce sandbox;
- cancellation non garantisce rollback di effetti già iniziati;
- il renderer umano non è un formato JSON pubblico;
- packaging, installazione pulita e artifact RC appartengono alla Fase 4;
- gli scenari live Ollama/llama.cpp appartengono alla Fase 5.

# Verdetto

GO alla Fase 4 — Packaging e installazione.
