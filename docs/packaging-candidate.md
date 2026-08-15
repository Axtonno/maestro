# Maestro Packaging and Release Artifacts

Stato: Contratto di packaging v0.1.0

Lo stesso percorso riproducibile produce packaging candidate, release
candidate e release finale. Lo stato è sempre esplicito nel manifest e nella
guida d'installazione inclusa; un artifact non viene rinominato o sovrascritto
per cambiarne lo stato.

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

La release finale usa `--status release` e una versione senza prerelease:

```sh
./scripts/verify-package-candidate.sh --version v0.1.0 --status release
./scripts/package-candidate.sh --version v0.1.0 --status release --output dist
```

Deve essere prodotta da un worktree pulito successivo alla documentazione
pubblica e superare nuovamente installazione e quick start. Una release
candidate non viene rinominata come release.

# Contenuto

- binario `maestro` Linux `amd64`;
- `LICENSE` Apache-2.0, `NOTICE` e licenze delle dipendenze distribuite;
- README, changelog, security policy e documentazione pubblica per
  installazione, quick start, configurazione, CLI, reference agent, sicurezza,
  compatibilità, troubleshooting e release notes;
- configurazione strict `version: 1` senza secret;
- profilo ufficiale read-only senza tool mutanti e con
  `workspace_mutate: deny`;
- fixture `maestro-laravel-mini@1.0.0`, priva di dipendenze installate;
- manifest dell'artifact.

# Non garanzie

Lo stato `packaging-candidate` non certifica provider o modelli live; lo stato
`release-candidate` registra che il gate live è superato ma non rappresenta la
pubblicazione finale. Anche lo stato `release` resta limitato dalla matrice in
`compatibility.md`: llama.cpp e il reference agent mutante sono sperimentali e
non supportati. Nessun artifact include sandbox, installer privilegiato,
aggiornamento automatico o dependency download.
