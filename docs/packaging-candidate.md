# Maestro Packaging Candidate

Stato: Milestone 8, Fase 4

Il packaging candidate è un artifact tecnico installabile usato per verificare
build, contenuto, checksum e indipendenza dal checkout. Non è una release
candidate e non è candidato alla pubblicazione: la promozione richiede i gate
live della Fase 5.

# Identità

Il nome è:

```text
maestro-<version>-linux-amd64.tar.gz
```

La Fase 4 ha prodotto la prima iterazione `v0.1.0-pc.1`; le iterazioni
successive conservano lo stesso formato e incorporano gli hardening emersi nei
gate live. `ARTIFACT-MANIFEST.txt` registra versione, commit, piattaforma,
versione Go e fixture. La guida d'installazione inclusa viene renderizzata con
la versione esatta dell'archive e `maestro version` deve restituire la stessa
versione e lo stesso commit.

Lo script usa `status=packaging-candidate` per default. Soltanto dopo un gate
live positivo della Fase 5 può essere invocato esplicitamente con
`--status release-candidate`; il manifest deve riportare lo stesso stato. Un
artifact `pc.N` fallito non viene rinominato o sovrascritto.

# Contenuto

- binario `maestro` Linux `amd64`;
- `LICENSE` Apache-2.0, `NOTICE` e licenze delle dipendenze distribuite;
- README e documentazione preliminare per installazione, configurazione, CLI e
  UX operativa;
- configurazione strict `version: 1` senza secret;
- profilo ufficiale read-only senza tool mutanti e con
  `workspace_mutate: deny`;
- fixture `maestro-laravel-mini@1.0.0`, priva di dipendenze installate;
- manifest dell'artifact.

# Non garanzie

Il candidate non certifica provider o modelli live e non implica supporto oltre
Linux `amd64`. llama.cpp e il reference agent mutante sono sperimentali e non
supportati nella v0.1.0. Il candidate non include sandbox, installer
privilegiato, aggiornamento automatico o dependency download.
