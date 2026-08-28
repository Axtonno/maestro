# Milestone 14 — Checkpoint interno

Ultimo aggiornamento: 2026-08-28

## Stato corrente

- milestone: in corso;
- fase attiva: transizione Fase 5 -> Fase 6;
- ultimo commit di fase: `72de866` — Fase 4 completata; il commit Fase 5 è il
  prossimo passo;
- ultimo commit verde della baseline:
  `72de866`;
- worktree atteso: Fase 5 documentata, in attesa di validazione e commit.

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

Validare il profilo candidato e creare il commit della Fase 5. Quindi eseguire
l'audit Fase 6 e chiudere la milestone con esito `direct_chat_deferred` e
handoff ripetibile, senza trasformare C0-C4 `not_run` in PASS.
