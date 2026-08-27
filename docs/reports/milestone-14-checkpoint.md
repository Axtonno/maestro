# Milestone 14 — Checkpoint interno

Ultimo aggiornamento: 2026-08-27

## Stato corrente

- milestone: in corso;
- fase attiva: transizione Fase 3 -> Fase 4;
- ultimo commit di fase: `6e5639a` — Fase 2 completata; il commit Fase 3 è il
  prossimo passo;
- ultimo commit verde della baseline:
  `6e5639a`;
- worktree atteso: Fase 3 completa e verificata, non ancora committata.

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

Creare il commit della Fase 3. Quindi implementare la Fase 4: matrice negativa
completa, streaming esplicito con equivalenza deterministica, terminali
deadline/cancel, immutabilità e harness anti-leak.
