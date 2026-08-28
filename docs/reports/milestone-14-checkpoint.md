# Milestone 14 — Checkpoint interno

Ultimo aggiornamento: 2026-08-28

## Stato corrente

- milestone: in corso;
- fase attiva: transizione Fase 4 -> Fase 5;
- ultimo commit di fase: `818209d` — Fase 3 completata; il commit Fase 4 è il
  prossimo passo;
- ultimo commit verde della baseline:
  `818209d`;
- worktree atteso: Fase 4 completa, in attesa della suite finale e del commit.

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

Rieseguire suite, race, vet e diff check e creare il commit della Fase 4.
Quindi congelare il candidate record e avviare il solo preflight read-only
della Fase 5, senza avviare Ollama o scaricare modelli.
