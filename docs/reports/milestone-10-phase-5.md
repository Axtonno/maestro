# Milestone 10 — Report Fase 5

Data: 2026-08-20

Stato: **COMPLETATA — lifecycle mutativo e freshness terminale verificati**

## Risultato

Il percorso Controlled Mutation rende ora osservabili, tramite sole allowlist
redatte, gli stadi `proposal`, `approval`, `apply` e `reindex`. Il risultato di
un tool mutativo può dichiarare pubblicamente se il commit non è avvenuto o se
l'effetto è applicato, oltre alla durability, senza serializzare path, diff,
contenuto, arguments o errori esterni.

## Contratto applicativo

- una run ammette un solo tentativo mutativo;
- un secondo tentativo è rifiutato prima di raggiungere `Tool Runtime.Invoke`;
- deny terminale, apply fallito e cancellazione non tornano al modello per un
  retry;
- l'inizio di `Execute` rende la sessione stale;
- dopo un commit riuscito non cancellato il loop indicizza e ricostruisce il
  bundle;
- snapshot e bundle devono dichiarare la stessa nuova generazione;
- soltanto apply successful e refresh fresh consentono un turno finale;
- `mutation_failed` e `context_refresh_failed` restano reason sintetiche e
  distinguibili nel `RunError`.

Una failure di refresh mantiene `ContextStale=true` e la generazione precedente.
Una cancellazione post-commit conserva lo stato applicato/durable nell'evento,
termina come canceled e non dichiara il workspace invariato.

## Rendering e anti-leak

Il progress renderer emette righe `mutation` per proposta, approval, apply e
reindex. I payload contengono soltanto run, stadio, stato, effetto, durability e
generazione. La diff resta visibile esclusivamente nel prompt TTY di approval
definito nella Fase 3.

## Gate

| Verifica | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |
| Apply fallito terminale senza retry | PASS |
| Secondo tentativo rifiutato prima dell'esecuzione | PASS |
| Refresh fallito conserva sessione stale | PASS |
| Cancellazione post-commit comunica effetto applicato | PASS |
| Sequenza redatta e test anti-leak | PASS |

La Fase 5 è completata. La Fase 6 può eseguire la matrice end-to-end, l'audit
pubblico e la decisione finale verso la Milestone 11.
