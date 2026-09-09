# Changelog

Le modifiche rilevanti di Maestro sono registrate in questo file. Durante la
serie `0.x` i contratti pubblici restano sperimentali e ogni breaking change
deve essere dichiarato nelle note di release.

## [0.5.0] - 2026-09-06

### Released

- Controlled Mutation host-bound productizzata con `qwen2.5-coder:14b`,
  Direct Chat separata su `qwen3.5:9b`, routing senza fallback e CLI opt-in;
- release qualificata su artifact Linux amd64 eseguito in WSL2 con reference
  RTX 5070 da 12 GB, con preview,
  approval, apply atomico, allow/deny/abstain/stale e residency verificata;
- claim limitato a un intervallo selezionato in un singolo file PHP sotto
  `app/`.

Report: `docs/reports/milestone-37-final.md`.

### Field adoption

- M38 qualifica l'asset pubblico invariato su Ubuntu 24.04.4 Linux amd64
  nativo, ThinkPad T490s CPU-only: doctor 14/14, sei apply autorizzati esatti,
  deny e stale senza scritture, correttezza aggregata 11/12, completion 12/12,
  utilità mediana 5/5 e zero effetti vietati;
- F02 conserva come limite osservato una risposta semanticamente errata alla
  domanda generica senza file. Report: `docs/reports/milestone-38-final.md`.

## [Unreleased]

### Onboarding

- aperta M41 con `maestro setup`: configurazione v4 privata e idempotente,
  verifica Ollama/modelli, consenso prima del pull e rifiuto di digest o
  workspace inattesi;
- aggiunto `maestro mutate --preview`, dry-run non interattivo con approver
  deny-only, diff completo e zero authority di scrittura;
- ridotti README, installazione e Quick Start al percorso utente; spostati
  baseline Git, fixture, doctor completo e gate in `docs/validation.md`;
- l'archive v0.5.0 resta immutabile; packaging e clean install del prossimo
  artifact sono ancora `not_run`.

### Prototype

- aperta M40 con un prototipo VS Code dependency-free che inoltra chat sul
  file attivo, mutation sulla selezione, doctor e identità alla CLI nel
  terminale integrato, senza scritture o approval nell'estensione;
- congelati perimetro e matrice M40; i controlli offline sono presenti mentre
  la prova Extension Development Host resta `not_run`, quindi il prototipo non
  amplia il claim v0.5.0 e non è autorizzato alla pubblicazione.

### Changed

- riposizionata la documentazione pubblica: Maestro è una workstation AI
  locale con nucleo operativo Direct Chat + Controlled Mutation, costruita sul
  runtime modulare;
- separati esplicitamente support claim corrente e roadmap; riscritte le
  superfici pubbliche ancora ferme a v0.3.x;
- aggiunte le guide `docs/install-and-try.md` e
  `docs/controlled-mutation-support.md`, incluse nell'allowlist di packaging;
- aggiunta `docs/current-capabilities.md` come pagina verità per capacità,
  hardware consigliato, modelli, limiti e difetto legacy dell'archive v0.5.0;
- completata M39 con verdetto
  `documentation_onboarding_public_trial_ready`: trial dell'asset pubblico al
  primo tentativo, doctor 14/14, chat corretta, mutation `201` → `202` con
  allow-once e un solo file nel diff Git;
- congelata la baseline pubblica v0.5.0 con hash dei documenti e policy di
  evoluzione in `docs/v0.5.0-public-baseline-freeze.yaml`;
- congelata l'evidenza M38 con hash e policy di immutabilità in
  `docs/milestone-38-field-adoption-freeze.yaml`.

### Added

- selezione host-bound immutabile, decoder `host-bound-mutation-decision-v1`,
  fingerprint di coordinate/span/diff e adapter interno al commit atomico;
- terminali mutativi redatti distinti per target assente/ambiguo, sorgente
  stale, target protetto e approval rifiutata;
- envelope strict `mutation-decision-v1` con astensioni tipizzate per target
  assente, ambiguo o richiesta insufficiente;
- contratto strict `mutation-proposal-v1` e compilatore deterministico per una
  singola sostituzione esatta;
- superficie opt-in `workspace.replace` con preview, digest pre/post,
  fingerprint, stale check e apply atomico;
- adapter senza fallback per tool calling nativo e structured output.

### Qualification

- M36 productization: profilo v4 con `qwen3.5:9b` per Direct Chat e
  `qwen2.5-coder:14b` per Controlled Mutation, CLI opt-in
  `workspace replace`, doctor separato/all, prompt/schema congelati e routing
  senza fallback incrociati. La residency v4 su RTX 5070 con swap WSL
  disabilitato qualifica tre cicli `chat → mutation → chat`: un solo modello
  in VRAM, latenze sotto soglia, zero fallback, offload CPU, OOM o swap. La
  pubblicazione v0.5.0 resta una decisione separata. I gate packaging,
  installazione fuori checkout, doctor e prove live allow/deny/stale/abstain
  e chat → mutation → chat sono qualificati sul candidate costruito da
  `cb2a408`;
- M35 conclusa con `mutation_specific_model_qualified`: il confronto di tre
  profili seleziona `qwen2.5-coder:14b`; qualifica development e holdout al
  100% su output, positivi, astensioni, target, preview, approval, apply e
  terminali, con zero effetti vietati. `qwen3.5:9b` resta per Direct Chat;
  aperta M36 per la productization, v0.5.0 non ancora autorizzata;
- M34 conclusa nel ramo B: `qwen3.5_9b_host_bound_mutation_profile_rejected`.
  Audit offline di prompt/payload/adapter e renderer senza conflitti residui
  rilevati; stop al tuning mutativo dello stesso profilo. Zero generazioni
  nuove o repliche M33. Aperta M35 per un modello dedicato alla mutazione;
  Direct Chat conserva `qwen3.5:9b`, v0.5.0 resta non autorizzata;
- M33: qualifica host-bound respinta; 12/12 output live validi e 7/7 target e
  preview corretti, ma 7/10 proposte positive e approval raggiunte. Tre falsi
  negativi, zero effetti non autorizzati o fuori selezione; v0.5.0 non autorizzata;
- M32: contratto binario 18/18 output validi, 8/9 positivi e 3/3 astensioni
  semantiche; qualification respinta perché quattro target assenti/duplicati
  sono stati riscritti dal modello e hanno raggiunto preview, poi negate senza
  effetti; v0.5.0 non autorizzata;
- M31: ambiguità meccaniche sicure 6/6 e zero effetti illeciti, ma proposte
  positive 4/6 e terminali 14/17; qualification respinta;
- M30: 14/14 output validi e 5/5 proposte positive corrette, ma 6/7
  astensioni; recovery respinta senza effetti o violazioni di sicurezza;
- gate deterministici, suite Linux LF, race, vet e diff check verdi;
- confronto semanticamente decisivo non eseguito per indisponibilità del
  provider target; M28 chiusa come `controlled_mutation_transport_unresolved`;
- nessun ampliamento del claim v0.4.0 e nessun candidate v0.5.0 autorizzato.

## [0.4.0] - 2026-09-03

### Changed

- risposte Direct Chat entro 450 parole, focalizzate sul flusso richiesto e
  strutturate in fatti osservati, inferenze possibili e dati non determinabili;
- inferenze, refactoring e test compaiono soltanto quando richiesti;
- `response_invalid` continua a respingere ogni terminale diverso da `stop`.

### Qualification

- holdout indipendente: 10/10 completion, 8/10 correct, zero falsità
  materiali, zero risposte invalide e zero mutazioni;
- suite, race, vet, compatibilità v2/v3, packaging riproducibile, installazione
  pulita, gate live e riscaricamento pubblico sono verdi;
- invariati modello, profilo, budget generativo e support claim read-only.

## [0.3.1] - 2026-09-02

### Added

- schema chat v3 con `num_predict` e residency espliciti;
- diagnostica di configurazione tipizzata e redatta;
- identità binaria verificabile con `version --diagnostic`;
- heartbeat redatto durante le generazioni lunghe.

### Compatibility

- invariati Ollama 0.33.1, `qwen3.5:9b`, digest, context e support claim GPU;
- il profilo distribuito usa `num_predict: 1024` e `residency: 5m`;
- nessuna promessa CPU, agentica, multi-file o mutativa viene aggiunta.

## [0.3.0] - 2026-08-29

### Added

- Direct Chat tool-free con zero o un file esplicito contained;
- schema v2 con profilo chat separato e doctor dedicato;
- streaming opt-in con aggregazione e pubblicazione atomica;
- envelope, reason code, limiti file/output e generation controls osservabili;
- packaging Linux `amd64` riproducibile con stati distinti per candidate,
  release candidate e release finale.

### Security

- path assoluti, traversal, symlink, file non regolari e race di lettura
  falliscono chiusi prima della disclosure;
- file non attendibile e domanda restano messaggi distinti, con domanda finale;
- ogni request dichiara zero tool e non costruisce agent/retrieval come fallback;
- errori e telemetria escludono prompt, response, contenuto, root e secret;
- il profilo distribuito imposta `workspace_mutate: deny` e non include agent o
  superfici mutative.

### Compatibility

- baseline qualificata: Ollama 0.33.1, `qwen3.5:9b` con digest congelato,
  context 4096, thinking false e temperatura zero;
- solo Direct Chat single-file read-only è qualificabile per v0.3.0;
- reference/verified agent, retrieval, tool e mutation restano non supportati;
- schema e CLI restano sperimentali durante la serie 0.x.

## [0.2.0] - 2026-08-21

### Added

- contratti sperimentali per preview content-bound, risultato mutativo ed
  eventi redatti del lifecycle apply/reindex;
- implementazione Linux di `workspace.patch` atomica e fail-closed, mantenuta
  fuori dal profilo ufficiale;
- benchmark versionato per la qualificazione Controlled Mutation, con matrice
  deterministica, gate fail-fast e report redatti.

### Security

- le istruzioni esplicite `Read ...` del reference agent richiedono una
  `workspace.read` riuscita prima di accettare il testo finale;
- gli arguments invalidi dei soli tool read-only possono essere corretti dal
  modello entro gli hard limit tramite un risultato sintetico redatto;
- prima della read richiesta dal quick start, il modello vede soltanto
  `workspace.read` e riceve regole esplicite sul solo campo `path`;
- le istruzioni `Read <logical-path> ...` eseguono una read deterministica
  attraverso Tool Runtime prima della prima inferenza, conservando policy,
  containment, hard limit e correlazione del risultato;
- la configurazione distribuita resta read-only con soli list/read/search e
  `workspace_mutate: deny`;
- packaging e quick start non includono il profilo mutante né presentano
  `workspace.patch` come supportato;
- ADR-0032 registra `mutation_deferred` dopo il failure del primo Gate A live.

### Compatibility

- CLI e schema YAML `version: 1` del percorso supportato restano invariati;
- le API Go aggiunte sono sperimentali e additive; nessuna API viene
  stabilizzata dalla release;
- Linux `amd64`, Ollama e `llama3.1:8b` restano l'unica combinazione
  qualificabile per il reference agent Laravel read-only.

## [0.1.1] - 2026-08-15

### Fixed

- il workspace Laravel usa ora una scan policy sorgente bounded invece della
  policy filesystem generica, evitando che asset generati in `public/` o dati
  runtime in `storage/` impediscano l'analisi di progetti reali;
- file sorgente Laravel fino a 2 MiB restano indicizzabili, mantenendo il
  limite complessivo di 64 MiB e gli stessi controlli su path e symlink;
- il packaging seleziona le note di release corrispondenti alla patch release
  e rende la versione del quick start coerente con artifact e manifest.
- la policy conserva `README.md` e i metadata versionati della fixture, così
  il quick start mantiene il contesto qualificato della v0.1.0.
- una pseudo-tool-call JSON incorporata in testo esplicativo viene riconosciuta
  quando presenta una forma tool-call (`name` con `arguments`, `parameters` o
  `input`), anche se nomina un tool inesistente; il loop richiede il canale
  tool reale senza eseguire o accettare la pseudo-call.

Nessun tool, permesso o support claim mutativo viene aggiunto.

## [0.1.0] - 2026-08-15

### Added

- CLI locale `doctor`, `models`, `agents`, `run` e `version`;
- configurazione YAML strict `version: 1` con target e hard limit espliciti;
- artifact Linux `amd64` riproducibile con manifest e checksum SHA-256;
- reference agent Laravel read-only con list/read/search;
- adapter Ollama e fixture `llama3.1:8b`;
- fixture embedding `embeddinggemma:latest`;
- progress redatto, cancellazione SIGINT/SIGTERM e shutdown bounded;
- correzione protocollare bounded quando un modello stampa una tool call JSON
  dichiarata invece di invocare l'interfaccia tool;
- configurazione e fixture Laravel utilizzabili direttamente dall'archive;
- documentazione di installazione, quick start, sicurezza, compatibilità,
  troubleshooting e API sperimentali;
- licenza Apache-2.0 e attribution delle dipendenze.

### Security

- il profilo ufficiale non registra tool mutanti e imposta
  `workspace_mutate: deny`;
- containment dei path workspace, rifiuto symlink e limiti di I/O;
- eventi operativi basati su allowlist senza prompt, contenuti, argomenti,
  fingerprint, root fisica o secret;
- permission model exact-action e nessun auto-approval globale.

### Known limitations

- soltanto Linux `amd64`, Ollama e `llama3.1:8b` sono qualificati per il
  reference agent supportato;
- llama.cpp, mutazioni, approval mutativa e tool/agent di terze parti sono
  sperimentali/non supportati;
- trusted in-process, nessuna sandbox, recovery o memoria persistente;
- CLI, configurazione e package Go restano sperimentali nella serie 0.x.

[0.1.0]: https://github.com/Axtonno/maestro/releases/tag/v0.1.0
[0.1.1]: https://github.com/Axtonno/maestro/releases/tag/v0.1.1
[0.2.0]: https://github.com/Axtonno/maestro/releases/tag/v0.2.0
[0.3.0]: https://github.com/Axtonno/maestro/releases/tag/v0.3.0
