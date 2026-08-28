# Milestone 14 — Checkpoint interno

Ultimo aggiornamento: 2026-08-28

## Stato corrente

- milestone: completata con esito `direct_chat_deferred`;
- fase attiva: nessuna; Fase 6 completata;
- ultimo commit di fase: `8a35add` — Fase 5 completata;
- ultimo commit di base della suite finale: `8a35add`; il worktree Fase 6 ha
  superato suite completa, race, vet, diff check e scan anti-leak;
- worktree atteso: Fase 6 documentata, in attesa della suite finale e del
  commit conclusivo.

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

Eseguire suite finale, race, vet, diff check e scan anti-leak, quindi creare il
commit conclusivo della Fase 6. Per riprendere la qualifica live, aprire
Milestone 15 e creare un nuovo candidate ID soltanto dopo provider/GPU e
digest/template osservati; C0-C4 M14 restano definitivamente `not_run`.
