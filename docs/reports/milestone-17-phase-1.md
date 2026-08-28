# Milestone 17 — Fase 1: freeze del contratto e audit del candidato

Data: 2026-08-28

Stato: **COMPLETATA — PASS**

Commit baseline: `2759c332c8edcc66f12aa12fd219e32dff3e1dba`

## Obiettivo verificato

La fase ha riconciliato il percorso Direct Chat development-only consegnato
dalla Milestone 14 con il PASS live della Milestone 15. Il risultato è un
contratto unico, un candidate record iniziale e un backlog chiuso per le Fasi
2–7. Il support claim pubblico resta v0.2.0 e non è stato prodotto alcun
artifact v0.3.0.

## Handoff importato

Sono stati importati esclusivamente gli elementi direct/chat qualificati:

| Campo | Valore congelato |
|---|---|
| esito M15 pertinente | `direct_chat_candidate` |
| piattaforma finale già osservata | WSL2 / Ubuntu 24.04 / RTX 5070 12 GB |
| provider | Ollama 0.33.1 dentro WSL2 |
| modello | `qwen2.5-coder:7b` |
| digest catalogo | `dae161e27b0e` |
| context / thinking | 4096 / disabilitato |
| timeout | 5 minuti |
| C0 senza file | PASS 3/3 |
| C1 single-file | PASS 3/3 |
| stream/non-stream | PASS 2/2 |
| sicurezza | workspace invariato; zero tool, retrieval e fallback |

Il failure del verified agent, B01 `not_run` e Controlled Mutation non entrano
nel candidate direct/chat e non vengono reinterpretati come capability
supportate.

## Candidate record iniziale

| Elemento | Identità |
|---|---|
| sorgente | commit `2759c332c8edcc66f12aa12fd219e32dff3e1dba` |
| ADR-0033 | SHA-256 `3cd965c8d27218ab9f155ce25e63eded7cdfc0b96abb786029c281bfae5aa6f6` |
| profilo M15 | SHA-256 `fe471d519749315da13b76f5e788d49f96150d5ce3f672f170810229c48f5dbd` |
| configuration example v2 | SHA-256 `9c341a3d879e5f8ad2fde2c9d92e7ac6f171c2c5d6b7c4f19f741aa1c0094593` |
| fixture `routes/api.php` | SHA-256 `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |
| prompt template e servizio | blob Git `d6afb1737e4e5118bae22d2375d0adf2504a91ee` |
| comando CLI | blob Git `d6acdd3f800deb1209ead62bc9b35c4ad54f7aa4` |
| loader single-file | blob Git `06fd7e17ada2c6eedd774091e1229e96db30bad7` |
| schema/config | blob Git `b35b6fb4d9b5d138a399288226940fa8918306f6` |
| binario di audit `-trimpath` | SHA-256 `d6aa37122a50525c28f3f61549213377507611ae598ee8d7e17a95e2e85eab3b` |

Il binario di audit è development-only, è stato scritto fuori dal repository e
non è un packaging candidate. La sua versione incorporata è
`v0.2.1-0.20260828205742-2759c332c8ed`.

## Contratto congelato

- `maestro chat [--config path] [--file logical-path] [--stream] [question]`;
- domanda bounded da positional oppure stdin, mai da entrambe le sorgenti;
- zero o un solo file logico esplicito sotto il workspace;
- una sola completion provider con zero tool e `tool_choice: none`;
- nessun retrieval, index, sessione, approver, runtime agentico o fallback;
- schema v2 con profili chat e agent distinti;
- `num_ctx` e thinking richiesti devono essere mappati o respinti dal preflight;
- risposta validata e bounded prima della pubblicazione dell'envelope;
- failure sintetici, reason code ed exit code definiti da ADR-0033;
- verified agent, multi-file automatico e qualsiasi mutazione fuori scope.

## Baseline eseguita

Ambiente di sviluppo osservato: Linux amd64, Ubuntu 24.04.4, kernel
`7.0.0-30-generic`, Go 1.24.5. La cache Go è stata collocata sotto `/tmp`
perché la cache predefinita dell'ambiente era read-only; questo non modifica
sorgenti, dipendenze o output di test.

| Comando | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| test mirati chat/config/Ollama `-count=3` | PASS |
| `go build -trimpath ./cmd/maestro` | PASS |
| `git diff --check` prima della fase | PASS |

## Audit e fase owner

| Delta dimostrato | Fase owner |
|---|---:|
| prova strutturale del graph tool-free e assenza di fallback | 2 |
| policy esplicita per file vuoto, BOM, confini byte e race residue | 3 |
| validazione chat ancora accoppiata ai requisiti agent e doctor solo agentico | 4 |
| canonicalizzazione errori flag, fault stream e chiusura stream | 5 |
| harness M17, oracoli qualitativi e nuova serie live congelata | 6 |
| packaging v0.3.0, config installabile e documentazione pubblica | 7 |

L'accoppiamento di configurazione è il delta principale: il documento v2 è
strict, ma il caricamento e `ValidateExecutionProfile` validano anche agent,
tool e budget non necessari a una completion diretta; inoltre `doctor` prova
soltanto il modello agentico. La Fase 4 deve risolverlo senza rendere implicita
l'autorità e senza degradare la validazione del comando agent.

## Gate di uscita

- contratto unico ADR/CLI/configurazione: **PASS**;
- distinzione chat/agent e autorità: **PASS**;
- requisito–fase–test assegnato: **PASS**;
- baseline repository-wide: **PASS**;
- artifact, tag o support claim anticipati: **ASSENTI**.

Verdetto della fase: **PASS**. La Fase 2 può iniziare.
