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

La versione di Fase 4 è `v0.1.0-pc.1`. `ARTIFACT-MANIFEST.txt` registra
versione, commit, piattaforma, versione Go e fixture. `maestro version` deve
restituire la stessa versione e lo stesso commit.

# Contenuto

- binario `maestro` Linux `amd64`;
- `LICENSE` Apache-2.0, `NOTICE` e licenze delle dipendenze distribuite;
- README e documentazione preliminare per installazione, configurazione, CLI e
  UX operativa;
- configurazione strict `version: 1` senza secret;
- fixture `maestro-laravel-mini@1.0.0`, priva di dipendenze installate;
- manifest dell'artifact.

# Non garanzie

Il candidate non certifica provider o modelli live, non chiude il debito
llama.cpp e non implica supporto oltre Linux `amd64`. Non include sandbox,
installer privilegiato, aggiornamento automatico o dependency download.
