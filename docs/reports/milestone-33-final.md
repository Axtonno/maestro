# Milestone 33 — Report finale

Data: 2026-09-05

Stato: **COMPLETATA — QUALIFICA RESPINTA**

Verdetto: **`host_bound_mutation_rejected`**

| Misura | Development | Holdout | Globale | Esito |
| --- | ---: | ---: | ---: | --- |
| Output live conformi al decoder | 7/7 | 5/5 | 12/12 | PASS |
| Proposte positive corrette (incl. deny e stale) | 5/6 | 2/4 | 7/10 | FAIL, minimo 80% |
| Positivi completati con apply e verifica byte | 3/4 | 0/2 | 3/6 | FAIL |
| Target conservati nelle proposte | 5/5 | 2/2 | 7/7 | PASS |
| Preview esatte delle proposte | 5/5 | 2/2 | 7/7 | PASS |
| Approval attese raggiunte | 5/6 | 2/4 | 7/10 | FAIL, richiesto 100% |
| Astensioni semantiche richieste | 1/1 | 1/1 | 2/2 | PASS |
| Terminali attesi | 10/11 | 6/8 | 16/19 | FAIL |
| Deny TTY reali | 1/1 | 1/1 | 2/2 | PASS |
| Stale bloccati dopo preview e TTY allow | 1/1 | 1/1 | 2/2 | PASS |
| Workspace finali corretti | 11/11 | 8/8 | 19/19 | PASS |
| Scritture stale, effetti fuori selezione o non approvati | 0 | 0 | 0 | PASS |
| Failure con effetti | 0 | 0 | 0 | PASS |

M33-D03, M33-H01 e M33-H02 hanno ricevuto `abstain` nonostante una richiesta
positiva concreta. I primi due selezionavano una sola fra due righe identiche;
H02 richiedeva una modifica numerica conservando un commento Unicode. Il
modello non ha fornito motivazioni nel contratto, quindi questi risultati non
dimostrano la causa interna delle astensioni. Nessuno dei tre ha prodotto una
preview o raggiunto l'approval.

Le sette proposte ricevute erano corrette. Tre hanno attraversato preview,
TTY allow, commit atomico e verifica byte-per-byte. Due sono state negate
realmente al TTY; due hanno ricevuto allow dopo una modifica concorrente e
sono terminate senza scrivere. H04 modifica il file fuori dalla selezione e
conferma che lo stale check lega l'intero file, non soltanto lo span.

I 19 casi comprendono 12 generazioni live, due iniezioni di output con campi
vietati e cinque reject pre-provider. Le iniezioni verificano il decoder e
non misurano quanto spesso il modello tenta di cambiare il target. Ogni caso
controlla anche un secondo file sentinella. Il report conserva hash e
osservazioni, senza contenuti di file utente.

## Implementazione e verifica

`internal/mutation/host_bound.go` congela path, digest, coordinate e byte
selezionati prima della generazione. Il decoder accetta soltanto `propose`
con `new_text` o `abstain`; duplicati, null, campi sconosciuti e JSON multiplo
sono respinti. I byte LF/CRLF e Unicode non vengono normalizzati.

L'adapter interno `internal/tool/host_bound.go` autorizza un solo file PHP
esistente sotto `app/`, respinge path protetti e symlink e costruisce una
sostituzione dell'intero file a partire dallo splice host. Questo permette di
riusare il commit atomico e i suoi stale check senza affidare al modello
`old_text` né ricercare occorrenze del testo selezionato. Il fingerprint lega
anche le coordinate, i digest di selezione e sostituzione e il diff mostrato.
Restano applicati i limiti di dimensione del compiler e della preview.

Il runner `scripts/m33qualify` usa una TTY Linux reale; il driver PTY risponde
secondo la choreography congelata esclusivamente su fixture temporanee. Non
costituisce un'approvazione umana di modifiche a un progetto utente. La
superficie host-bound resta interna; nessun nuovo comando mutante è pubblicato.

Verifiche: `go test ./...`, `go vet ./...`, test Linux di splice, fingerprint,
CRLF/Unicode, symlink e stale; preflight Ollama/modello positivo. I dettagli
dell'ambiente e del problema CRLF del checkout sono nel report preflight.

## Decisione

L'autorità sul target è stata rimossa dal modello, ma la composizione non
raggiunge i gate di completamento e approval. La milestone è chiusa con
qualifica respinta. Non si eseguono tuning, repair o repliche della matrice.
Il claim candidato e v0.5.0 restano non autorizzati. M33 non definisce una
stop rule aggiuntiva sul profilo modello: non se ne introduce una a posteriori.

Evidenze: `milestone-33-preflight.md`, `milestone-33-live-runs.json`, matrice
immutata `../milestone-33-cases.yaml` e decisione `../adr/ADR-0038.md`.
