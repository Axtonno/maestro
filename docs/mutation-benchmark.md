# Maestro Controlled Mutation Benchmark

Versione: 0.1.0

Stato: Harness della Milestone 11 implementato

Ultimo aggiornamento: 2026-08-21

---

# Scopo

Il Mutation Benchmark qualifica il percorso Laravel Controlled Mutation senza
modificare il Developer Benchmark read-only. Usa il contratto congelato in
`mutation-qualification-profile.yaml` e produce un report distinto con schema:

```text
mutation-qualification-report/1.0.0
```

La presenza dell'harness e un risultato deterministico positivo non
costituiscono supporto live.

---

# Validazione del profilo

Il comando iniziale verifica profilo strict, matrice, gate, fixture embedded,
digest iniziale e digest finale proposto:

```text
maestro bench mutation \
  --profile docs/mutation-qualification-profile.yaml
```

Campi sconosciuti, gate modificati, scenari mancanti, riferimenti assoluti,
digest non validi o una fixture divergente falliscono chiusi. La validazione
materializza una copia privata della fixture e la elimina prima del ritorno.

---

# Runner fail-fast

Il runner esegue una serie per volta:

- Gate A: tre tentativi;
- Gate B: due tentativi;
- Gate C: tre tentativi.

La serie termina al primo stato diverso da `passed`. Ogni tentativo riceve una
deadline nuova di dieci minuti; il cleanup non può essere contato come PASS se
fallisce. I gate successivi vengono orchestrati soltanto dopo il PASS del gate
precedente.

La matrice deterministica per sviluppatori esegue direttamente i package che
possiedono i fault seam e produce entrambi i report:

```text
maestro bench mutation \
  --mode deterministic \
  --output mutation-deterministic.json \
  --markdown mutation-deterministic.md
```

Prima dei gate live si esegue il preflight read-only:

```text
maestro bench mutation --mode preflight
```

I gate live vengono invocati separatamente e in ordine:

```text
maestro bench mutation --mode gate-a --output gate-a.json
maestro bench mutation --mode gate-b --output gate-b.json
maestro bench mutation --mode gate-c --output gate-c.json
```

Gate C rifiuta stdin non interattivo prima dell'I/O provider. Su TTY presenta
la preview reale e richiede una nuova risposta `o`/`once` per ogni tentativo.

---

# Evidenza fisica

Ogni fixture viene materializzata dal dataset embedded
`maestro-laravel-mini@1.0.0`. Lo snapshot:

- rifiuta symlink e file non regolari;
- ordina i path logici;
- registra SHA-256 per file e digest aggregato del workspace;
- non conserva la root fisica;
- consente di dimostrare che soltanto il target ammesso è cambiato.

---

# Report

Il JSON canonico contiene esclusivamente:

- identità redatte di profilo, candidato, gate e scenario;
- versione, commit e digest del binario quando disponibili;
- hardware dichiarato;
- stato, reason code e failure class;
- contatori, terminale e lifecycle redatto;
- digest iniziale, atteso, finale e aggregato;
- approval classificata, freshness e cleanup.

Prompt, risposte, arguments, risultati tool, diff, contenuti e path fisici non
hanno campi nel contratto. Il decoder rifiuta campi sconosciuti e documenti
multipli. JSON e Markdown vengono pubblicati atomicamente con permessi `0600`.

---

# Confine di sicurezza

Approver e provider controllati sono ammessi soltanto nella matrice
deterministica. Un PASS Gate C richiede il prodotto reale, una TTY reale e una
decisione umana `allow once`. Il benchmark non avvia Ollama, non scarica
modelli e non modifica automaticamente i limiti per ottenere un PASS.
