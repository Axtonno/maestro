# Changelog

Le modifiche rilevanti di Maestro sono registrate in questo file. Durante la
serie `0.x` i contratti pubblici restano sperimentali e ogni breaking change
deve essere dichiarato nelle note di release.

## [0.1.0] - 2026-08-15

### Added

- CLI locale `doctor`, `models`, `agents`, `run` e `version`;
- configurazione YAML strict `version: 1` con target e hard limit espliciti;
- artifact Linux `amd64` riproducibile con manifest e checksum SHA-256;
- reference agent Laravel read-only con list/read/search;
- adapter Ollama e fixture `llama3.1:8b`;
- fixture embedding `embeddinggemma:latest`;
- progress redatto, cancellazione SIGINT/SIGTERM e shutdown bounded;
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
