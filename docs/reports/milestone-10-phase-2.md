# Milestone 10 — Report Fase 2

Data: 2026-08-20

Stato: **COMPLETATA — proposta patch e preview content-bound**

## Risultato

`workspace.patch` prepara ora la modifica completa prima dell'autorizzazione.
La `PreparedInvocation` conserva una preview strutturata e bounded inclusa nel
proprio fingerprint; l'esecuzione consuma il contenuto risultante preparato e
non ricostruisce l'intento dal prompt o da un nuovo output del modello.

## Contratti introdotti

- `pkg/tool.Preview` e `PreviewField`, immutabili e con defensive copy;
- `NewPreparedInvocationWithPreview`, compatibile con il costruttore esistente;
- preview inclusa nel fingerprint SHA-256 della prepared invocation;
- summary, metadata monoriga, media type e body con limiti espliciti;
- body massimo 256 KiB e rifiuto di UTF-8 invalido, NUL e campi non sicuri.

## Proposta `workspace.patch`

Durante `Prepare` il tool:

1. normalizza e valida gli arguments;
2. accetta soltanto path logici sotto `app/` con suffisso `.php` e senza
   componenti hidden;
3. riapre la root con `os.Root` e rifiuta symlink e file non regolari;
4. legge il file entro la `ScanPolicy`;
5. verifica digest, singola occorrenza e modifica non vuota;
6. calcola il contenuto risultante;
7. produce una diff unified deterministica con tre righe di contesto;
8. costruisce una preview con tool, path, digest e precondizione;
9. lega arguments preparati, action e preview allo stesso fingerprint.

La preparazione non scrive il workspace e non richiede approval. Una preview
che eccede il limite fallisce chiusa.

## Copertura

- immutabilità e defensive copy della preview;
- variazione della preview che cambia il fingerprint;
- rifiuto di summary invalida;
- diff concreta per una patch PHP;
- assenza della root fisica in summary e diff;
- file invariato dopo `Prepare`;
- rifiuto di file non PHP, digest stale e occorrenza ambigua;
- scenario agentico end-to-end aggiornato al confine `app/**/*.php`.

## Gate

| Verifica | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |
| Scritture durante prepare | Nessuna |
| Preview non content-bound | Rifiutata dal fingerprint |

La Fase 2 è completata. La Fase 3 può integrare questa preview nell'approver,
vietare i grant mutativi run-scoped e aggiungere il profilo opt-in separato.
