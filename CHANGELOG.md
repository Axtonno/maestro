# Changelog

Le modifiche rilevanti di Maestro sono registrate qui. Durante la serie `0.x`
i contratti pubblici restano sperimentali e ogni breaking change viene
dichiarato nelle note di release.

## [Unreleased]

### Added

- `maestro setup` per creare una configurazione locale, verificare Ollama e
  richiedere consenso prima del download dei modelli;
- `maestro mutate --preview` per generare un diff senza autorità di scrittura;
- prototipo VS Code locale che inoltra comandi alla CLI nel terminale
  integrato; resta sperimentale e non amplia il support claim.

### Changed

- documentazione pubblica riorganizzata attorno a installazione, capacità,
  limiti, sicurezza, piattaforme e benchmark riproducibili;
- documentazione strategica, report intermedi e roadmap tattica separati dal
  repository pubblico;
- valutato e respinto il candidato single-model `qwen2.5-coder:14b` per qualità
  Direct Chat insufficiente; setup e support claim restano a due modelli.

## [0.5.0] - 2026-09-06

### Added

- Controlled Mutation host-bound su un intervallo esplicito di un file PHP
  sotto `app/`;
- preview completa, approval allow-once, stale check e apply atomico;
- profili distinti per Direct Chat e Controlled Mutation, senza fallback;
- artifact Linux `amd64` verificato anche su Linux nativo CPU-only.

### Known limitations

- nessun supporto multi-file o selezione autonoma del target;
- nessun agent autonomo, comando shell o Git eseguito dal modello;
- Windows nativo, macOS, altri provider e altri modelli non sono supportati.

I risultati aggregati sono pubblicati in [docs/benchmarks.md](docs/benchmarks.md).

## [0.4.0] - 2026-09-03

### Changed

- risposte Direct Chat limitate e focalizzate sul flusso richiesto;
- validazione più severa dei terminali del provider;
- invariati modello, profilo e support claim read-only.

## [0.3.1] - 2026-09-02

### Added

- schema chat v3 con limiti di generazione e residency espliciti;
- diagnostica di configurazione tipizzata e redatta;
- identità binaria verificabile con `version --diagnostic`;
- heartbeat redatto durante generazioni lunghe.

## [0.3.0] - 2026-08-29

### Added

- Direct Chat senza file o su un singolo file esplicito;
- schema v2 con profilo chat separato e doctor dedicato;
- streaming opt-in con pubblicazione atomica;
- packaging Linux `amd64` riproducibile.

### Security

- containment del workspace e rifiuto di path assoluti, traversal e symlink;
- nessun tool, retrieval o mutation usato come fallback;
- telemetria senza prompt, risposta, contenuto o secret.

## [0.2.0] - 2026-08-21

### Added

- contratti sperimentali per preview e risultati mutativi;
- implementazione atomica e fail-closed di `workspace.patch`, non inclusa nel
  profilo ufficiale della release;
- benchmark versionato per Controlled Mutation.

## [0.1.1] - 2026-08-15

### Fixed

- scansione Laravel limitata ai sorgenti rilevanti;
- coerenza tra quick start, versione dell’artifact e manifest;
- rifiuto delle pseudo-tool-call stampate come semplice testo.

## [0.1.0] - 2026-08-15

### Added

- CLI locale `doctor`, `models`, `agents`, `run` e `version`;
- configurazione YAML strict con target e hard limit;
- artifact Linux `amd64` con manifest e checksum SHA-256;
- reference agent Laravel read-only;
- adapter Ollama, fixture, documentazione e licenza Apache-2.0.

[0.1.0]: https://github.com/Axtonno/maestro/releases/tag/v0.1.0
[0.1.1]: https://github.com/Axtonno/maestro/releases/tag/v0.1.1
[0.2.0]: https://github.com/Axtonno/maestro/releases/tag/v0.2.0
[0.3.0]: https://github.com/Axtonno/maestro/releases/tag/v0.3.0
[0.3.1]: https://github.com/Axtonno/maestro/releases/tag/v0.3.1
[0.4.0]: https://github.com/Axtonno/maestro/releases/tag/v0.4.0
[0.5.0]: https://github.com/Axtonno/maestro/releases/tag/v0.5.0
