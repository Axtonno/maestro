# Milestone 6 — Phase 3 Report

Stato: Completata

Data: 2026-08-11

---

# Risultato

Gli snapshot del Context Engine possono ora includere analisi strutturate
versionate. Registry, selezione esplicita, validazione e failure atomiche sono
implementati; `context.go-ast@1` dimostra il contratto con un parser reale.

---

# Implementazione

- registry analyzer thread-safe;
- validazione nil, typed nil, ID, versione e duplicati;
- lista analyzer immutabile nel workspace;
- risoluzione deterministica per ID;
- ambiguità esplicita quando più analyzer auto-selected supportano un file;
- composizione intenzionale tramite `WorkspaceOptions.Analyzers`;
- callback `Supports` e `Analyze` fuori lock;
- recovery dei panic come failure;
- validazione dell'identità dell'output;
- esclusione dei documenti opachi;
- pubblicazione delle analysis nello snapshot atomico;
- analyzer Go basato esclusivamente sulla libreria standard.

---

# Evidenza AST Go

Le fixture verificano package, import, constant, type, struct field, method e
function. Relazioni e chunk mantengono intervalli byte nel testo normalizzato.

Una sorgente sintatticamente incompleta produce `go_parse_error` e conserva i
simboli disponibili dall'AST parziale, senza trasformare un problema locale in
failure del workspace.

---

# Concorrenza e cancellazione

Una fixture blocca `Supports` mentre un altro analyzer viene registrato: la
registrazione completa senza attendere la callback. Un analyzer cancellabile
bloccato in `Analyze` propaga `context.Canceled` e non pubblica snapshot.

Errori, panic e output con zero identity sono coperti e lasciano l'indice
invariato.

---

# Test

Coperti:

- output AST reale e intervalli;
- parse parziale con diagnostica;
- integrazione automatica durante indexing;
- ambiguità e composizione esplicita;
- errori, panic e output invalido;
- registrazione nil, typed nil e duplicata;
- callback fuori lock;
- cancellazione durante analisi;
- copie difensive della configurazione analyzer.

Comando:

```text
GOCACHE=/tmp/maestro-go-build go test ./internal/contextengine ./pkg/contextengine
```

Esito: superato.

---

# Gate

Un analyzer reale dimostra il contratto senza introdurre dipendenze esterne o
logica language-specific nel Runtime Core. Ambiguità, diagnostiche e failure
globali hanno semantiche distinte. La Fase 3 è completata; retrieval e builder
possono usare documenti, simboli e chunk nella Fase 4.
