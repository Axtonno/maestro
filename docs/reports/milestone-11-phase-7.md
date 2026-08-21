# Milestone 11 — Report Fase 7

Data: 2026-08-21

Stato: **COMPLETATA — verdetto `mutation_deferred`**

## Riconciliazione delle evidenze

| Evidenza | Risultato | Conseguenza |
|---|---|---|
| Profilo e harness | validi e versionati | prova riproducibile |
| Matrice deterministica | 15/15 PASS | invarianti del prodotto dimostrati |
| Preflight live | PASS sul lower bound | ambiente idoneo a eseguire il candidato |
| Gate A | FAIL, 0/1 eseguito | candidato escluso |
| Gate B | 0/2, non eseguito | stop rule Gate A |
| Gate C | 0/3, non eseguito | stop rule Gate A/B |

Il failure Gate A è classificato `model`: la read call è valida, mentre la
patch call non rispetta gli arguments esatti congelati. Non risultano failure
di harness o ambiente, differenze fisiche non spiegate, approval emesse o
effetti impliciti.

## Identità verificata

| Campo | Valore |
|---|---|
| Candidato | `v0.2.0-m11-qc.2` |
| Commit | `7e8ba62da22ad1942f3688b880922eacbec0889f` |
| Binario SHA-256 | `9870772b25f482eb4a5e539cea86e44aa19740e929c5789eab091d10c70101a3` |
| Dimensione | 9.310.464 byte |
| Profilo SHA-256 | `a64b7557ccd24f32bb4fb7cee7d64b630e16ec017c0776b2549d86bcd8480cac` |
| Provider/modello | Ollama / `ibm/granite4.1:8b` |

I due report deterministici autorevoli e i due report Gate A hanno permessi
`0600`; i rispettivi JSON usano lo schema
`mutation-qualification-report/1.0.0` e coincidono con le viste Markdown.

## Audit finale

- tutti i 15 campioni deterministici, il campione Gate A e i cinque tentativi
  live non avviati sono contabilizzati;
- il fail-fast è stato applicato senza retry o modifica del profilo;
- il workspace Gate A conserva digest file e digest aggregato iniziali;
- Gate C non ha acquisito TTY, approval o authority;
- il profilo read-only e il support claim v0.1.x restano invariati;
- nessun artifact di release è stato prodotto;
- ADR-0032 registra il rinvio e vincola la Milestone 12 al read-only.

## Verifica repository-wide

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
/tmp/maestro-v0.2.0-m11-qc.2 bench mutation \
  --profile docs/mutation-qualification-profile.yaml
git diff --check
```

Esito: **PASS**. La validazione riconosce 3 gate e 15 scenari con il digest
congelato. Tutti i JSON Milestone 11 sono parseabili, usano lo schema v1 e il
profilo atteso; la scansione dei report machine-readable non trova contenuto
della patch, root fisiche, directory temporanee, payload tool, prompt, response
o credenziali.

## Verdetto

Milestone 11: **completata**.

Controlled Mutation: **NO-GO**, esito `mutation_deferred`.

Milestone 12: **GO limitato alla productization read-only**. Una capability
mutativa futura richiede un nuovo candidato e la ripetizione completa dei gate
congelati, senza riusare questo preflight o i risultati storici come PASS.
