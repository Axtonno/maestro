# Milestone 35 — Preflight della selezione

Data: 2026-09-05. Esito: PASS, senza generazioni.

- Linux `amd64`, Ubuntu 24.04 su WSL, NVIDIA GeForce RTX 5070 con 12.227 MiB;
- Ollama 0.33.1; tre modelli locali con digest, template, licenza e dimensione
  registrati in `milestone-35-model-metadata.json`;
- 12 casi congelati: 9 positivi e 3 astensioni necessarie, nessun caso M33;
- un tentativo per coppia modello/caso, ordine candidato e caso fisso;
- stesso prompt, schema e profilo di generazione; nessun tool o thinking;
- nessun retry, repair, fallback o tuning dopo l'avvio;
- `go test ./...` e `go vet ./...`: PASS nella copia Linux con fixture Git LF;
- report live assente e creato con `O_EXCL` prima della prima generazione.

Freeze SHA-256:

| Artefatto | Digest |
| --- | --- |
| Matrice di selezione | `2ae5fe27fd6bbcaee6d2f59acad1b3a5362822c737af642dcfbe4ba4e92073da` |
| Schema | `bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d` |
| Prompt | `594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732` |

L'eleggibilità richiede 12/12 output conformi, 9/9 positivi corretti e 3/3
astensioni corrette. La latenza aggregata decide soltanto tra profili eleggibili.
