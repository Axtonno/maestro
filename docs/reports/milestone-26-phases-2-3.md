# Milestone 26 — Fasi 2–3: replay sintetico e cattura raw

Data: 2026-09-02

Stato: **COMPLETATE — nessuna generazione eseguita**

## Replay offline del validator

Il replay usa soltanto fixture sintetiche normalizzate e non tenta di
ricostruire i quattro body M25 mancanti. Un controllo valido è accettato; undici
fixture negative esercitano separatamente ruolo, tool call, finish reason,
content vuoto, content whitespace, UTF-8, NUL, model mismatch, usage input,
usage output e durata. Tutte le esecuzioni usano un provider in memoria: zero
chiamate HTTP e zero generazioni.

Gli asset sono il test `TestMilestone26OfflineValidatorFixtureReplay`, il test
per UTF-8/durata e il dataset JSON in `internal/directchat/testdata`. Non è
stato modificato alcun ramo del validator.

## Harness privato di cattura

L'harness locale è stato creato fuori dal checkout e non è un asset di
prodotto o di candidate. Riceve un request JSON privato e invia una singola
POST a `/api/chat`; conserva status HTTP, header e body grezzi, byte count,
SHA-256, exit code e timestamp. Non copia il request e non stampa payload.
Il digest dell'eseguibile installato è
`6493cfdf58857ddc158e6c55acae82068e229951446af2cb84e31daa9cfce38c`.

Il processo usa `umask 077`, directory `0700` e file `0600`, rifiuta una
destinazione appartenente a qualunque worktree Git e non sovrascrive un capture
ID esistente. La forma d'uso privata è:

```text
capture-raw.sh --request request.json --output /private/captures --id M26-C4
```

La verifica sintattica Bash è PASS. Non è stata effettuata alcuna chiamata
live: P4 resta separata e richiede gli input privati congelati.

I package interessati (`internal/directchat` e `internal/provider/ollama`)
superano i test. La suite globale non è verde nel checkout Windows/WSL: restano
failure estranei a M26 sui digest congelati M21/mutation e su due sostituzioni
testuali legacy di `internal/productconfig`; nessuno dei file coinvolti è stato
modificato in queste fasi.

## Vincoli confermati

- codice prodotto, validator, adapter, profilo, candidate e prompt: invariati;
- fixture: esclusivamente sintetiche;
- raw M25: ancora indisponibili e non dichiarati replayabili;
- nuova generazione: zero;
- prossimo gate: P4, matrice diagnostica una tantum sui quattro input M25.
