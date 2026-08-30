# Milestone 20 — Fase A: attribuzione della latenza

Data: 2026-08-30

Stato: **COMPLETATA — `model_hardware_bound`**

## Sintesi

Sui due task congelati e a modello residente, l'esatto binario pubblico
Maestro v0.3.0 non introduce overhead materiale rispetto al replay diretto
Ollama dello stesso body byte-identico. Le quattro differenze appaiate al
terminale sono comprese tra -0,18 e +0,11 secondi, molto sotto la soglia del
maggiore tra 5 secondi e 15%.

Il profilo mostra invece costi elevati di primo uso: 57,5 secondi sul primo
task e 64,4 secondi sul primo task streaming, poi esclusi come warm-up
dichiarati. Nel task streaming il primo chunk delle run formali arriva in
0,96-0,99 secondi e il terminale in 16,7-16,9 secondi. Maestro raggiunge lo
stesso terminale senza ritardo materiale, ma rende l'output visibile soltanto
alla fine per il contratto atomico v0.3.0.

La Fase B con `qwen2.5-coder:7b` è autorizzata.

## Identità e preflight

| Campo | Valore osservato |
|---|---|
| hardware | ThinkPad T490s, Linux amd64, 8 CPU logiche, CPU-only |
| Ollama | 0.32.14, loopback |
| modello | `qwen3.5:9b` |
| digest | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| archive v0.3.0 | SHA-256 `6c8f0e883ec8f8c05571fc2e7bc1f4ecac608c2bd7e338395ae0a4253fff1aaf` |
| binario v0.3.0 | SHA-256 `378a0533083b9a00be6c0212ca52001cebc5f77b476a20038bc8e08d1fc3d42d` |
| commit incorporato | `3f4c7d4b4fd2e380644cf250ce9e8fec2311af53` |
| profilo | context 4096, thinking false, temperatura 0, timeout 5 minuti |
| relay temporaneo | SHA-256 `a2ea0bdbb6778704fdd0d3fdc4d334aee39b38ccac8fb2ff7b536122f75fb61c` |
| doctor chat | PASS 5/5 attraverso il relay |

L'archive è stato riscaricato dalla GitHub Release in `/tmp`, verificato ed
estratto senza sostituire installazioni. Il comando `maestro` già nel `PATH`
aveva invece SHA-256
`ad6949742d8eb2e389a6a5d73721efc0a6fee7dada13975343ce112b78cc6cd0` e
un output `version` privo di versione/commit release. Tutte le prove hanno
quindi usato il path assoluto del binario v0.3.0 verificato.

## Protocollo

I task usano la fixture pubblica inclusa nell'archive:

- `M20-A0`: no-file, risposta di insufficienza del contesto, non-streaming;
- `M20-A1`: `routes/api.php`, metodo/endpoint/controller/action, streaming.

Il relay loopback salva ogni request body con permesso `0600`, registra solo
digest, dimensione e tempi nel log e inoltra i chunk immediatamente. Il body
catturato dalla prima invocation Maestro viene riusato dal percorso diretto.
Per ogni task i cinque body — warm-up incluso — hanno lo stesso SHA-256:

| Task | SHA-256 body | Byte |
|---|---|---:|
| A0 | `a10402a661bd2740850e9ebec7e9aaef43cb243c859e8902307b764633e14fa6` | 2.212 |
| A1 | `797342ce140150b4adefac131ad8f176a2312cfcd79dfb6a34fc7ced61b8dcc7` | 2.688 |

A0 contiene `stream=false`; A1 contiene `stream=true`. Entrambi usano
`qwen3.5:9b`, `think=false`, `num_ctx=4096`, temperatura 0, soli ruoli
system/user e zero tool.

## Risultati formali

| Task | Ripetizione | Ollama diretto | Maestro terminale | Delta Maestro | Primo chunk diretto / Maestro |
|---|---:|---:|---:|---:|---:|
| A0 | 1 | 6,960 s | 6,780 s | -0,180 s | terminale / terminale |
| A0 | 2 | 6,897 s | 6,884 s | -0,013 s | terminale / terminale |
| A1 | 1 | 16,904 s | 16,747 s | -0,157 s | 0,987 / 0,964 s |
| A1 | 2 | 16,733 s | 16,839 s | +0,106 s | 0,963 / 0,964 s |

Tutte le run completano con HTTP 200 o exit 0, finish `stop`, risposta
corretta e nessuna asimmetria terminale. Le invocation Maestro formali hanno
RSS massimo osservato tra 8.608 e 8.992 KiB; questo dato riguarda il processo
CLI, non la memoria del runner Ollama.

## Warm-up e buffering

| Task | Primo uso escluso | Primo chunk | Terminale | Nota |
|---|---:|---:|---:|---|
| A0 | Maestro | terminale | 57,456 s | primo caricamento/uso del profilo |
| A1 | Maestro | 47,068 s | 64,355 s | primo uso del nuovo body streaming |

Il replay A0 immediatamente successivo riporta da Ollama load 0,422 s,
prompt-eval 0,488 s ed eval 6,045 s. Il replay A1 riporta load 0,455 s,
prompt-eval 0,527 s ed eval 15,917 s. I warm-up non vengono nascosti: mostrano
una coda operativa reale, ma non possono essere attribuiti all'adapter perché
le run appaiate successive coincidono.

Su A1 la finestra media tra primo chunk e terminale è circa 15,8 secondi. È il
costo percettivo del buffering atomico, non overhead al terminale. Qualunque
feedback futuro deve quindi usare stderr redatto senza pubblicare contenuto
parziale su stdout.

## Immutabilità e limiti dell'evidenza

La fixture post-run coincide byte per byte (`diff -qr` vuoto) con una seconda
estrazione dall'archive verificato. `routes/api.php` conserva SHA-256
`7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39`.
Tutte le 25 capture complessive A/B hanno permesso `0600` e la directory
permesso `0700`.

Non è stato acquisito un campionamento CPU/RSS del processo Ollama; il record
ha soltanto RSS della CLI e le duration provider. Questa omissione limita
l'analisi micro-architetturale ma non il confronto primario, che dispone di
body byte-identici, quattro coppie complete e tempi osservati al relay.

## Verdetto

Le quattro coppie soddisfano la regola di equivalenza temporale e nessuna
raggiunge la deadline. Il verdetto è **`model_hardware_bound`**: nella matrice
congelata Maestro non è la causa materiale della latenza terminale. Questo non
afferma overhead zero e non modifica il support claim v0.3.0.

