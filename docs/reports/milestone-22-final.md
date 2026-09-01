# Milestone 22 — Audit finale e decisione

Data: 2026-09-01

Stato: **COMPLETATA — candidate qualificato**

Verdetto: `v0.3.1_operational_hardening_qualified`

## Decisione

Le correzioni operative sviluppate nelle Milestone 20 e 21 sono productizzate
nel contratto v0.3.1 senza ampliare il support claim. Il profilo pubblico v3
conserva Ollama 0.33.1, `qwen3.5:9b`, digest, context 4096, thinking false e
temperatura zero; aggiunge `num_predict: 512` e residency 5 minuti.

Restano esclusi CPU, nuovi modelli/provider, agent, retrieval, multi-file,
tool e mutazioni. Il verdetto qualifica il sorgente e un release candidate;
non dichiara pubblicati tag o asset v0.3.1.

## Gate deterministici

| Gate | Esito |
|---|---|
| `go test ./...` su blob Git LF | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| sintassi script Bash | PASS |
| doppio packaging byte-identico | PASS |
| checksum, archive audit e installazione pulita | PASS |
| diagnostica, identità, heartbeat e redazione | PASS |

Il checkout Windows converte le fixture congelate in CRLF; eseguire i test
direttamente da `/mnt/c` produce digest diversi senza modifica Git. I gate
Linux sono stati quindi eseguiti su una materializzazione LF dei blob Git,
coerente con gli artifact. Non sono stati aggiornati digest o oracoli storici.

Il verifier ha costruito due volte `v0.3.1-rc.1`, stato
`release-candidate`, profilo `release`, ottenendo archive byte-identici e
SHA-256
`1750de78f94636ddd10ee7790e77e3140e9ac75fd77efd05c83f58b896db7ff6`.
Il commit incorporato appartiene al repository temporaneo di qualifica e non
costituisce un candidate pubblicabile dal checkout principale.

## Gate live RTX 5070

Ambiente osservato:

- WSL2 / Ubuntu 24.04, Linux `amd64`;
- NVIDIA RTX 5070 12 GB;
- Ollama 0.33.1;
- `qwen3.5:9b`, digest
  `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`.

| Prova | Esito |
|---|---|
| identità e doctor chat | PASS, 5/5 |
| no-file epistemico | PASS |
| single-file complete | PASS |
| single-file stream | PASS |
| equivalenza semantica | PASS — POST `/orders`, `OrderController`, `store` |
| envelope v3 | PASS — 512, 5m, `truncated=false`, `stop` |
| traversal | PASS — `file_not_allowed`, exit 2 |
| fixture immutata | PASS |
| stderr allowlist | PASS |

Complete e stream sono terminate sotto 15 secondi e hanno quindi prodotto
zero heartbeat, come previsto dal contratto. Cadenza, limite, redazione e stop
del ticker sono coperti dalla suite deterministica e dal race detector.

## Output consegnato

- schema e profilo pubblico v3;
- packaging release con parametri operativi nel manifest;
- verifier aggiornato e script ripetibile del gate live;
- README, CLI, configurazione, compatibility, troubleshooting, installazione,
  changelog e note v0.3.1;
- piano e report finale della milestone.

## Chiusura

La Milestone 22 è completa con candidate qualificato. Una pubblicazione remota
richiede un commit release pulito, artifact finale con identità reale, tag
annotato, GitHub Release e verifica post-download; nessuna di queste azioni
esterne è stata implicitamente eseguita.

