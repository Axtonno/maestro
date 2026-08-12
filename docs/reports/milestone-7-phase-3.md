# Milestone 7 — Phase 3 Report

Stato: Completata

Data: 2026-08-12

---

# Risultato

La Fase 3 collega policy registry, authorizer, approval flow e grant al Tool
Runtime. Il percorso resta default-deny e nessun callback può fabbricare o
ricevere il permit operativo.

---

# Capacità consegnate

- `Rule` pubblica con matching exact effect/resource/workspace;
- `StaticPolicy` immutabile e senza ordine di preferenza;
- policy ID esatto e `ErrPolicyNotFound` senza fallback;
- decisione atomica per action multiple;
- deny precedence e default-deny per regole mancanti;
- prompt flow con Approver opzionale;
- deny recoverable/terminal preservato;
- allow one-shot e run-scoped;
- grant cache legata a policy/run/fingerprint;
- policy e Approver fuori lock;
- validazione e panic containment al boundary;
- authorization modello separata dalla tool invocation.

---

# Invarianti verificati

- prefix e workspace differenti non soddisfano una regola;
- matcher duplicati vengono rifiutati;
- una sola action denied nega l'intera richiesta;
- prompt senza Approver non consente l'effetto;
- approval deny non viene convertito in allow;
- grant run-scoped evita una nuova policy call soltanto per la stessa chiave;
- decisioni one-shot non vengono memorizzate;
- policy assente non seleziona il primo ID registrato;
- policy e Approver in panic non producono autorità;
- registrazione resta disponibile mentre una policy è bloccata.

---

# Test e gate

Coperti exact matching, default-deny, multi-action, duplicate rules, lookup,
prompt, approver, grant scope, callback fuori lock e panic.

```text
GOCACHE=/tmp/maestro-go-build go test ./pkg/tool ./internal/tool
GOCACHE=/tmp/maestro-go-build go test -race ./pkg/tool ./internal/tool
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

---

# Gate

Il permission model default-deny è operativo e l'executor non dispone di
percorsi alternativi all'authorizer. La Fase 4 può iniziare.
