# Milestone 33 — Preflight

Data: 2026-09-05. Esito: PASS, senza generazioni.

- Linux amd64, WSL Ubuntu 24.04, NVIDIA GeForce RTX 5070;
- Ollama 0.33.1, `qwen3.5:9b`, digest
  `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`;
- profilo: context 4096, output 1024, temperatura 0, thinking disabilitato,
  residency 5 minuti, timeout 6 minuti;
- `go test ./...` e `go vet ./...`: PASS nella copia Linux isolata;
- decoder strict, coordinate, splice con duplicati, Unicode/CRLF, fingerprint,
  symlink, deny dei path protetti e stale check verificati;
- 11 casi development e 8 holdout nuovi, definiti dopo il prompt;
- 12 casi con generazione live, 2 output avversari iniettati e 5 reject host
  pre-provider; le iniezioni non sono evidenza di comportamento del modello;
- TTY reale tramite `pty.fork`, adapter `TerminalApprover` e choreography
  allow/deny congelata su soli workspace temporanei;
- report creato con `O_EXCL` prima della prima generazione, checkpoint dopo
  ogni caso; nessun retry, repair o fallback.

Freeze SHA-256:

| Oggetto | Digest |
| --- | --- |
| Casi | `248df368f0b7440b2a21df39859a0819e9b3fbf350b0817b19fd920866b0851b` |
| Schema (byte Git LF) | `bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d` |
| Prompt | `a96d542063eacfe4431f336813911dec0794ab9c7cea9e1f16e5e0d1dcca2c99` |

La prima suite sul checkout Windows ha rilevato CRLF nelle fixture storiche
incompatibili con i digest congelati. La verifica completa usa un archivio
Git Linux con le aggiunte M33; nessuna fixture storica è stata modificata.
