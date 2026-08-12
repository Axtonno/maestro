# Milestone 7 — Phase 2 Report

Stato: Completata

Data: 2026-08-12

---

# Risultato

La Fase 2 implementa il catalogo Tool trusted in-process e un execution
boundary che non può avviare `Execute` senza consumare un permit interno
verificabile.

---

# Capacità consegnate

- catalogo thread-safe con ID e nome provider univoci;
- listing deterministico dei descriptor;
- registry policy predisposto senza default permissivo;
- resolution esatta del tool;
- `Prepare` e `Execute` fuori lock;
- validazione post-Prepare di identità, versione, fingerprint ed effect;
- permission request atomica per tutte le action;
- permit privato issuer/run/permission/prepared-bound;
- consumo one-shot atomico;
- deadline derivata dal context;
- validazione di result e limiti per item e byte;
- troncamento UTF-8 esplicito;
- isolamento dei panic;
- error envelope con cause preservate.

---

# Invarianti verificati

- una collisione ID o nome non sostituisce il tool esistente;
- callback tool non vengono eseguite sotto lock del catalogo;
- un tool non può cambiare tool/call/run durante `Prepare`;
- un tool non può produrre effect non dichiarati;
- il runtime default non esegue effetti;
- deny non invoca `Execute`;
- permit di un altro issuer, altro fingerprint o già usato viene rifiutato;
- result oltre item limit fallisce;
- output oltre byte limit viene troncato senza rompere UTF-8;
- panic e cancellazione non lasciano risultati validi parziali;
- registrazioni concorrenti dello stesso ID hanno un solo vincitore.

---

# Test

La suite `internal/tool` copre catalogo, collisioni, callback bloccanti,
default-deny, deny recuperabile, permit, replay, mismatch, prepared output,
limiti, panic, cancellazione e registrazione concorrente.

Gate richiesti:

```text
GOCACHE=/tmp/maestro-go-build go test ./internal/tool ./pkg/tool
GOCACHE=/tmp/maestro-go-build go test -race ./internal/tool ./pkg/tool
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

---

# Decisioni rinviate

- policy authorizer e default-deny rules: Fase 3;
- approval flow e grant consumption: Fase 3;
- session coordinator: Fase 4;
- loop agentico: Fase 5;
- tool workspace concreti: Fase 6;
- eventi e composition root: Fase 7.

---

# Gate

Catalogo ed executor proteggono gli invarianti della Fase 2 e nessun percorso
pubblico accetta autorità fabbricabile. La Fase 3 può iniziare.
