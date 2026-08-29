# Changelog

Le modifiche rilevanti di Maestro sono registrate in questo file. Durante la
serie `0.x` i contratti pubblici restano sperimentali e ogni breaking change
deve essere dichiarato nelle note di release.

## [0.3.0] - 2026-08-29

### Added

- Direct Chat tool-free con zero o un file esplicito contained;
- schema v2 con profilo chat separato e doctor dedicato;
- streaming opt-in con aggregazione e pubblicazione atomica;
- envelope, reason code, limiti file/output e generation controls osservabili;
- packaging candidate Linux `amd64` dedicato alla baseline Direct Chat.

### Security

- path assoluti, traversal, symlink, file non regolari e race di lettura
  falliscono chiusi prima della disclosure;
- file non attendibile e domanda restano messaggi distinti, con domanda finale;
- ogni request dichiara zero tool e non costruisce agent/retrieval come fallback;
- errori e telemetria escludono prompt, response, contenuto, root e secret;
- il profilo distribuito imposta `workspace_mutate: deny` e non include agent o
  superfici mutative.

### Compatibility

- baseline candidata: Ollama 0.33.1, `qwen3.5:9b` con digest congelato,
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
