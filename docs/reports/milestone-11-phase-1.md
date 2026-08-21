# Milestone 11 — Report Fase 1

Data: 2026-08-21

Stato: **COMPLETATA — contratto di qualificazione congelato**

## Risultato

Il profilo consegnato dalla Milestone 10 è stato trasformato in un protocollo
eseguibile. Task, fixture, digest, limiti, gate, livelli di prova, evidenze,
redazione e stop rule non sono più impliciti.

Controlled Mutation resta candidato non supportato. Nessuna prova live è stata
eseguita e nessun support claim è stato modificato.

## Contratto congelato

| Area | Decisione |
|---|---|
| Candidato | Linux `amd64`, Ollama, `ibm/granite4.1:8b` |
| Lower bound | i5-8365U, 8 CPU logiche, 15 GiB RAM, 4 GiB swap |
| Fixture | `maestro-laravel-mini@1.0.0`, nuova materializzazione per run |
| Patch | chiave JSON `data` sostituita con `order` nel controller |
| Digest | iniziale `4826abe...b1bd`, finale `509b566b...ac5d` |
| Limiti | provider 5m, run 10m, cleanup 30s, Gate A 256 token/turno |
| Gate | A `3/3`, B `2/2`, C `3/3`, consecutivi e fail-fast |
| Approval | TTY reale e `allow once`; fake esclusi dal PASS live |
| Report | JSON canonico, Markdown derivato, payload sensibili esclusi |

## Livelli di prova

La matrice fisica completa viene eseguita deterministicamente. Gate A/B/C e i
terminali che richiedono provider o TTY hanno anche prove live. La copertura
deterministica protegge gli invarianti, ma non sostituisce il Gate C live.

## Baseline

```text
GOCACHE=/tmp/maestro-go-build go test ./...
```

Esito: **PASS**.

```text
git diff --check
```

Esito: **PASS**.

## Gate

- criteri di accettazione espliciti: superato;
- comandi, ordine, stop rule e cleanup definiti: superato;
- scenario owner e livello di prova assegnati: superato;
- baseline deterministica verde: superato;
- support claim invariato: superato.

La Fase 1 è completata. La Fase 2 può implementare il Developer Benchmark
mutativo contro questo contratto.
