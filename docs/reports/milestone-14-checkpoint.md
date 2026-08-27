# Milestone 14 — Checkpoint interno

Ultimo aggiornamento: 2026-08-27

## Stato corrente

- milestone: in corso;
- fase attiva: Fase 2 completata, commit da creare;
- ultimo commit di fase: `a93c0e7` — Fase 1 completata;
- ultimo commit verde della baseline:
  `a93c0e7`;
- worktree atteso: modifiche complete della Fase 2 non ancora committate.

## Evidenze verdi

- `go test ./...`;
- `go test -race ./...`;
- `go vet ./...`;
- `git diff --check`.

Le suite Go richiedono `GOCACHE` sotto `/tmp` nell'ambiente Codex corrente.

## Decisioni congelate

- ADR-0033 separa `chat` e `agent`;
- `run` è alias esatto e deprecato di `agent` fino ad almeno v0.4.0;
- schema v2 per profili distinti; v1 resta valido solo per `agent/run`;
- chat single-file, zero tool/retrieval/state machine/fallback;
- `num_ctx` e `thinking` espliciti devono essere onorati o rifiutati.

## Prossimo passo riproducibile

Creare il commit della Fase 2. Dopo il commit, aggiornare il checkpoint con il
nuovo hash e avviare la Fase 3 con un application service chat separato,
loader single-file confinato e comando CLI non-streaming.
