# Milestone 20 — Fase B: candidato lower-resource

Data: 2026-08-30

Stato: **COMPLETATA — `thinkpad_profile_candidate`**

## Sintesi

`qwen2.5-coder:7b` completa sul ThinkPad 3/3 task no-file e 5/5 task
single-file senza timeout. Quattro dei cinque task single-file rispettano
l'oracolo; una risposta è respinta per contenuto aggiunto non supportato. La
coppia complete/stream è semanticamente equivalente e il workspace resta
immutato.

Sui medesimi cinque prompt e file, la mediana del candidato è 69,0 secondi
contro 123,9 secondi di `qwen3.5:9b`: -54,9 secondi e -44,3%. Sono superate le
soglie predefinite di almeno 20 secondi e 30%.

Il risultato stabilisce un candidato development-only per questa modalità e
questo hardware. Non aggiunge `qwen2.5-coder:7b` ai modelli supportati da
v0.3.0 e non cancella i failure qualitativi 2/5 dei candidate M17.

## Candidate congelato

| Campo | Candidato | Baseline appaiata |
|---|---|---|
| modello | `qwen2.5-coder:7b` | `qwen3.5:9b` |
| digest | `dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364` | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| dimensione catalogo | 4.683.087.561 byte | 6.594.474.711 byte |
| context / thinking | 4096 / false | 4096 / false |
| temperatura / timeout | 0 / 5 minuti | 0 / 5 minuti |
| binario | v0.3.0, SHA-256 `378a0533…` | identico |
| provider | Ollama 0.32.14 loopback | identico |
| config candidato | SHA-256 `c98a81f1bc40466dc10322af477011810cc6ec73940f52d126b67daa956eb39c` | profilo release con solo base URL relay |
| doctor chat | PASS 5/5 | PASS 5/5 in Fase A |

Entrambi i modelli erano già installati. Non è stato eseguito alcun pull,
tuning, cambio di prompt interno, rebuild o retry selettivo.

## No-file

| Run | Durata | Terminale | Qualità |
|---:|---:|---|---|
| 1 | 52,779 s | completed / stop | correct; insufficienza dichiarata |
| 2 | 6,286 s | completed / stop | correct; insufficienza dichiarata |
| 3 | 6,187 s | completed / stop | correct; insufficienza dichiarata |

La prima run include il caricamento dopo il cambio modello ed è conservata
nel gate di operabilità no-file. Non sono presenti endpoint inventati.

## Matrice single-file appaiata

| Task | Oracolo principale | qwen2.5 | Qualità candidato | qwen3.5 |
|---|---|---:|---|---:|
| B1 route | `POST /orders`, `OrderController::store` | 30,848 s | correct | 73,317 s |
| B2 controller | validazione, `OrderService::create`, `data`, 201; zero route/DB | 87,134 s | correct | 123,543 s |
| B3 order service | charge, create, dispatch `OrderCreated`, return order | 62,088 s | correct | 123,941 s |
| B4 checkout | skip indisponibili, preferred match, primo fallback, null | 190,628 s | incorrect | 227,666 s |
| B5 repository | `id=42` più payload; zero effetti esterni visibili | 68,998 s | correct | 176,997 s |

B4 del candidato descrive correttamente i rami richiesti, ma aggiunge un
esempio non presente nel file con gateway concreti e costruttori non
dimostrabili. Per il gate epistemico la risposta è `incorrect`, non viene
promossa perché il nucleo era corretto. Totale qualità candidato: **4/5**.

La baseline qwen3.5 è stata eseguita dopo un no-file di residenza escluso
dalla mediana, simmetricamente al candidato. B4 contiene inoltre un errore
materiale: afferma che un gateway successivo sovrascrive il fallback, mentre
il file assegna soltanto se il fallback è `null`. B5 conclude che i dati siano
mock/fittizi, inferenza non necessaria al confronto prestazionale.

## Latenza

| Metrica sui cinque task | qwen2.5 | qwen3.5 | Miglioramento |
|---|---:|---:|---:|
| completion | 5/5 | 5/5 | nessun timeout in entrambi |
| mediana | 68,998 s | 123,941 s | 54,943 s / 44,3% |
| minimo | 30,848 s | 73,317 s | — |
| massimo | 190,628 s | 227,666 s | — |

Il candidato supera la soglia appaiata di 20 secondi e 30%. Le mediane M19
restano un riferimento storico separato perché derivano da prompt e file di
`project-a`: 91,4 secondi sulle sole completion e 164,9 secondi includendo i
cinque tentativi timeout-capped.

La coda di 190,6 secondi dimostra che il profilo non è uniformemente rapido;
il verdetto è un candidato ThinkPad, non una promessa di latenza breve per
ogni file.

## Streaming, terminali e sicurezza

B1 è stato ripetuto con `--stream`. Complete e stream restituiscono gli stessi
quattro fatti e finish `stop`. La run stream dura 7,857 secondi; il primo chunk
raggiunge il relay a 0,488 secondi. La differenza temporale rispetto alla prima
run complete non viene interpretata come vantaggio dello streaming, perché il
body era già nella cache del modello.

Tutte le run candidato hanno exit 0, modello corretto, usage non negativo,
zero tool/retrieval/fallback e nessun output vuoto o terminale ambiguo. La
fixture post-run coincide byte per byte con una nuova estrazione dell'archive;
non è stata osservata alcuna mutazione.

## Gate e verdetto

| Gate | Esito |
|---|---|
| no-file 3/3 | PASS |
| single-file 5/5 | PASS |
| complete/stream 2/2 | PASS |
| qualità almeno 4/5 | PASS — 4/5 |
| nessun timeout | PASS |
| zero mutazioni | PASS |
| mediana almeno 30% e 20 s inferiore | PASS — 44,3% e 54,9 s |

Verdetto: **`thinkpad_profile_candidate`**.

La promozione a profilo distribuito resta vietata. Richiederebbe una decisione
separata e una nuova qualifica che riconcili esplicitamente questa matrice con
i failure M17, senza estendere agent, tool calling, retrieval o mutation.

