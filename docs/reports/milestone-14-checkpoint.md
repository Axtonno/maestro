# Milestone 14 — Checkpoint interno

Ultimo aggiornamento: 2026-08-27

## Stato corrente

- milestone: in corso;
- fase attiva: Fase 1 completata, commit da creare;
- ultimo commit verde della baseline:
  `3dd5e62bbce77a75519e784e749c338cc9685b75`;
- worktree atteso: modifiche documentali complete della Fase 1 non ancora
  committate.

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

Verificare diff e suite, quindi creare il commit della Fase 1. Dopo il commit,
aggiornare questo checkpoint con il nuovo hash e avviare la Fase 2 dai
contratti `productconfig.Config`, `provider.GenerationOptions` e dal mapping
Ollama.
