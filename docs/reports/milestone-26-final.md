# Milestone 26 — Report finale

Data: 2026-09-03

Stato: **COMPLETATA**

Verdetto: **`field_quality_recovered`**

La milestone chiude la diagnosi dei quattro `response_invalid` M25 e valida
un candidate v0.4.0 senza ampliare capability o autorità. Non è stata
pubblicata né taggata alcuna release.

## Causa e decisione software

I quattro capture diagnostici nuovi, distinti da un replay M25, hanno tutti
HTTP 200, content non vuoto, thinking assente, zero tool call, `done=true` e
1024 output token. Ollama termina uniformemente con `done_reason=length`; il
validator propaga correttamente il fail-closed pubblico `response_invalid`.
Non è dimostrato alcun difetto di adapter, validator o contratto provider.

`num_predict` resta 1024 e il validator non viene allentato. Poiché cambia il
prompt, la linea scelta è v0.4.0. Il contratto richiede un massimo editoriale
di 450 parole, priorità al flusso richiesto e tre sezioni ordinate:
`Observed facts`, `Possible inferences`, `Information not determinable`.
Inferenze, refactoring e test compaiono soltanto quando richiesti e restano
separati dai fatti.

## Stderr e heartbeat

L'heartbeat resta limitato alla sola finestra di generation: non parte durante
il preflight, è redatto e bounded, viene fermato prima del terminale e non
continua dopo successo o failure. I failure emettono una sola riga terminale
redatta dopo gli eventuali heartbeat. I test temporali e anti-leak sono PASS.

## Selezione candidate

| Candidate | Completion | Esito |
|---|---:|---|
| v0.4.0-rc.1 | 8/11 | respinto: tre `response_invalid` |
| v0.4.0-rc.2 | 11/11 | respinto: qualità epistemica insufficiente |
| v0.4.0-rc.3 | 11/11 | respinto: qualità epistemica insufficiente |
| v0.4.0-rc.4 | 11/11 | PASS |

Ogni matrice di candidate usa una sola run per task. I fallimenti non sono
ritentati sullo stesso candidate. Domande, file, oracoli e risposte complete
restano nell'evidenza privata `0700`/`0600` fuori dal repository.

## Gate conclusivi

| Metrica | Pre-release P9 | Field Adoption P10 | Soglia |
|---|---:|---:|---:|
| completion | 11/11 | 11/11 | almeno 85% |
| correct | 10/11 | 10/11 | almeno 80% valutabili |
| partial | 1/11 | 1/11 | informativa |
| `response_invalid` | 0 | 0 | 0 |
| terminali `length` | 0 | 0 | 0 |
| falsità materiali | 0 | 0 | 0 |
| utilità mediana | 5/5 | 5/5 | almeno 4/5 |
| mutazioni workspace | 0 | 0 | 0 |

P9 produce da 220 a 590 output token; P10 da 220 a 796. Tutte le 22
completion terminano `stop`, quindi il recupero non dipende da un aumento del
budget. Il binario privato rc.4 usato nei gate ha SHA-256
`040c393f09dc85876ecb25054b52b05fa699a8db4bc162e861f94568c8452639`.

## Confine finale

Restano fuori scope CPU, altri modelli, multi-file, agent, retrieval e
Controlled Mutation. La chiusura autorizza la preparazione futura della
release v0.4.0, ma non costituisce pubblicazione e non sostituisce i normali
gate di packaging, installazione e canale pubblico.

## Verifica repository

Sono PASS i test completi di `internal/directchat`, i test CLI mirati a chat,
heartbeat e identità, `go vet ./...` e `git diff --check` sui file modificati.
La suite globale conserva failure estranei già registrati durante P2–P3:
digest congelato M21 della fixture `routes/api.php`, coerenza della fixture
mutation e due casi legacy v2 di `internal/productconfig`. Nessuno di questi
failure coinvolge il prompt Direct Chat, il validator, stderr o i test M26;
non sono stati corretti perché fuori dal perimetro della milestone.
