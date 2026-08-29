# Maestro Packaging and Release Artifacts

Stato: contratto di packaging v0.3.0 Direct Chat

Lo stesso percorso normalizzato produce packaging candidate, release candidate
e release finale. Lo stato è incorporato nel manifest e nella guida inclusa;
un artifact non viene rinominato o sovrascritto per cambiarne stato.

## Identità

```text
maestro-<version>-linux-amd64.tar.gz
maestro-<version>-linux-amd64.tar.gz.sha256
```

`ARTIFACT-MANIFEST.txt` registra versione, commit, piattaforma, toolchain,
fixture, stato e baseline Direct Chat. Nome archive, manifest e
`maestro version` devono coincidere.

Il candidate di Milestone 17 usa un prerelease `v0.3.0-pc.N` con stato
`packaging-candidate`. Soltanto dopo la matrice finale live sul medesimo
archive può essere costruito un artifact distinto con stato
`release-candidate`. Tag e pubblicazione richiedono il verdetto finale
`direct_chat_product_baseline` e una build separata con stato `release`.

Dal checkout sorgente, i tre stati usano build distinte:

```sh
./scripts/verify-package-candidate.sh \
  --version v0.3.0-pc.1 --status packaging-candidate
./scripts/verify-package-candidate.sh \
  --version v0.3.0-rc.1 --status release-candidate
./scripts/verify-package-candidate.sh \
  --version v0.3.0 --status release
```

Gli esempi verificano in directory temporanee doppia build, checksum,
allowlist, installazione e identità. La persistenza sotto `dist/` avviene
soltanto per il candidate approvato e lo script rifiuta ogni overwrite.

## Riproducibilità

Il packaging richiede worktree pulito, Go 1.24.5 e GNU tar. Usa commit time
come `SOURCE_DATE_EPOCH`, path rimossi, build ID vuoto, ownership e permessi
normalizzati e gzip senza timestamp. Il gate costruisce due archive e richiede
uguaglianza byte-per-byte prima di verificare checksum e contenuto.

## Contenuto v0.3.0

- binario `maestro` Linux `amd64`;
- licenza, attribution, README, changelog e security policy;
- documentazione pubblica Direct Chat e note della versione;
- `configs/maestro.chat.example.yaml`, strict v2 e chat-only;
- fixture `maestro-laravel-mini@1.0.0` senza dipendenze installate;
- manifest dell’artifact.

L’archive non include profili agentici o mutativi, benchmark, raw trace,
report di sviluppo, prompt/response di qualificazione, secret, symlink,
directory VCS, `vendor` o `node_modules`.

## Baseline qualificata

Il manifest congela:

- modello `qwen3.5:9b`;
- digest `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`;
- context 4096, thinking `false`, temperatura zero;
- streaming abilitato ma opt-in;
- file/output massimi 1 MiB;
- `workspace_mutate: deny`.

Una divergenza richiede un nuovo candidate e non può essere corretta sulla
macchina di qualifica.

## Non garanzie

Lo stato `packaging-candidate` prova riproducibilità e installabilità locale,
non il provider live. `release-candidate` registra un gate live positivo ma
non equivale a pubblicazione. Nessuno stato qualifica agent, retrieval,
tool calling, mutation, llama.cpp, sandbox o endpoint remoti.
