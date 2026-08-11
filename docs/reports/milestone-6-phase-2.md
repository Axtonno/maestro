# Milestone 6 — Phase 2 Report

Stato: Completata

Data: 2026-08-11

---

# Risultato

Il Context Engine possiede ora una source filesystem sicura e un indice
in-memory capace di pubblicare snapshot generazionali all-or-nothing.

---

# Implementazione

Package:

```text
internal/contextengine
pkg/contextengine
```

Sono stati implementati:

- engine thread-safe con source registry;
- source built-in `context.filesystem`;
- scansione cancellabile senza lock globali;
- path relativi normalizzati;
- include, exclude e supporto ricorsivo `/**`;
- limiti per file, conteggio e byte totali;
- esclusione default di hidden, `.git`, `vendor` e `node_modules`;
- symlink non seguiti;
- verifica di stabilità dei file letti;
- normalizzazione BOM e line ending;
- classificazione deterministica di media type e linguaggio;
- contenuti binari esclusi oppure opachi su richiesta;
- snapshot ordinati e generazioni monotone;
- conservazione dell'ultimo snapshot su failure.

---

# Concorrenza

La source viene risolta sotto read lock e invocata senza lock dell'engine. Una
fixture bloccante dimostra che nuove source possono essere registrate mentre
una scansione è sospesa.

Il commit dello snapshot è breve e serializzato. Venti indicizzazioni
concorrenti dello stesso workspace producono venti pubblicazioni e una
generazione finale pari a 20 senza race o snapshot parziali.

Non viene introdotto singleflight: richieste distinte mantengono context e
risultati distinti.

---

# Failure atomiche

Sono verificate:

- source che restituisce errore;
- output con documenti duplicati;
- limiti superati;
- context già cancellato;
- source nil e typed nil;
- root o policy invalidi.

In tutti i casi nessun nuovo snapshot viene pubblicato. Le cause restano
ispezionabili insieme ai sentinel di dominio.

---

# Test

Coperti con directory temporanee:

- file Go e PHP annidati;
- ordine deterministico;
- normalizzazione CRLF;
- hidden, dipendenze escluse, symlink e binari;
- include per basename;
- limiti per numero, dimensione e totale;
- refresh riuscito e fallito;
- output source invalido;
- callback bloccante fuori lock;
- generazioni concorrenti;
- cancellazione e typed nil.

Comando dedicato:

```text
GOCACHE=/tmp/maestro-go-build go test ./internal/contextengine ./pkg/contextengine
```

Esito: superato.

---

# Documentazione

La semantica operativa è descritta in `docs/context-engine-indexing.md`.

---

# Gate

Snapshot deterministici, containment, limiti, pubblicazione atomica e callback
fuori lock sono coperti. La Fase 2 è completata; la Fase 3 può introdurre
registry e analyzer AST senza modificare la source.
