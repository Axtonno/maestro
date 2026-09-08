# Maestro Packaging and Release Artifacts

Stato: contratto di packaging v0.5.0

Lo stesso percorso normalizzato produce packaging candidate, release candidate
e release. Ogni stato è una build distinta, incorporata nel manifest e nel
binario; rinominare un file non ne cambia l'identità.

## Artifact

```text
maestro-<version>-linux-amd64.tar.gz
maestro-<version>-linux-amd64.tar.gz.sha256
```

`ARTIFACT-MANIFEST.txt` registra versione, commit, piattaforma, toolchain,
fixture, profilo, modelli, digest, parametri generativi e stato. Nome archive,
manifest e `maestro version --diagnostic` devono coincidere.

## Profilo v0.5.0

```sh
./scripts/verify-package-candidate.sh \
  --version v0.5.0-pc.1 \
  --status packaging-candidate \
  --profile mutation-productization
```

Per release candidate e release cambiano versione e stato; non si riutilizza
né sovrascrive l'archive precedente.

Il profilo `mutation-productization` include:

- binario Linux `amd64`;
- configurazione strict v4
  `configs/maestro.v0.5.0-candidate.yaml`;
- `qwen3.5:9b` per Direct Chat e relativo digest;
- `qwen2.5-coder:14b` per Controlled Mutation e relativo digest;
- prompt e schema mutativi congelati;
- fixture Laravel senza dipendenze installate;
- documentazione pubblica, licenza, attribution e note di release.

Il nome `candidate` del file YAML è conservato per compatibilità con il
profilo qualificato. Lo stato autorevole dell'artifact è nel manifest e nel
binario.

## Documentazione inclusa

L'archive contiene, fra le altre:

- `docs/install-and-try.md`;
- `docs/controlled-mutation-support.md`;
- installazione, quick start, CLI e configurazione;
- compatibility matrix, security model, known issues e troubleshooting.
- freeze, report finale, ambiente ed evidenza strutturata M38.

Non include raw trace, gli altri report interni di qualificazione, secret, directory
VCS, `vendor`, `node_modules` o profili agentici.

## Riproducibilità

Il packaging richiede worktree pulito, Go 1.24.5 e GNU tar. Usa il commit time
come `SOURCE_DATE_EPOCH`, rimuove i path di build, azzera il build ID,
normalizza ownership e permessi e usa gzip senza timestamp.

Il gate:

1. costruisce due archive indipendenti;
2. richiede uguaglianza byte-per-byte;
3. verifica checksum e allowlist del contenuto;
4. installa in una directory temporanea;
5. verifica identità di binario e manifest;
6. controlla configurazione, prompt e schema.

Il packaging non esegue automaticamente i gate live del provider: quelli
appartengono alla qualifica e alla release decision.

## Profili storici

Le varianti `release` chat-only e `cpu-qualification` restano disponibili
per riprodurre artifact e prove precedenti. Non rappresentano il percorso
pubblico corrente v0.5.0 e non ricevono authority mutativa implicita.

## Non garanzie

Packaging e manifest non qualificano nuovi provider, modelli, hardware,
linguaggi, agent, retrieval, tool calling, sandbox o endpoint remoti. Il
support claim rimane quello della [Compatibility Matrix](compatibility.md).
