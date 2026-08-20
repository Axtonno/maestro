# Milestone 9 — Report Fase 4

Data: 2026-08-20

Stato: **COMPLETATA — v0.1.1 qualificata come patch read-only**

## Risultato

L'osservazione sui workspace reali ha confermato un bug v0.1.x: la v0.1.0
applicava al plugin Laravel la scan policy filesystem generica e poteva
interrompere l'indicizzazione su asset generati comuni. La correzione minima
introduce una policy sorgente Laravel bounded, senza nuovi tool, permission o
support claim.

Le prove candidate hanno inoltre individuato due regressioni protocollari nel
modo in cui il loop trattava pseudo-tool-call testuali. Entrambe sono coperte
deterministicamente: JSON tool-call-shaped incorporato nella prosa non completa
più lo step, anche quando usa `parameters` o nomina un tool inesistente.

## Candidate

| Candidate | Commit | SHA-256 | Esito |
|---|---|---|---|
| `v0.1.1-rc.1` | `6fc48aa` | `8cc3c64d7218b55c4e2eb5af655524b0a4c629aa467bf813432eb76aa9974620` | Rifiutata: la policy perdeva il contesto root della fixture |
| `v0.1.1-rc.2` | `f369aa9` | `6ff787ddcb1751394e9fa1a94ab38eafa59ade36b482033f6ab200ab6a65c962` | Rifiutata: pseudo-call incorporata accettata come risposta finale |
| `v0.1.1-rc.3` | `0ceeedf` | `cd5eb21fbc41b0c6856e8aed4a32cb7811ce11cb0160875343b80788cb5539b9` | Rifiutata: pseudo-call con tool inesistente accettata |
| `v0.1.1-rc.4` | `c82a518` | `55d590cc6a3386d632c58bc38af35d4b6228259958a5b95a292df99db95a5abb` | Qualificata |

Tutti gli archive sono stati costruiti da worktree pulito. RC4 è byte-for-byte
riproducibile, incorpora il commit completo
`c82a5181e0f5a8c7fac9f811c0a4f6eaa5d2cea2`, supera checksum, inventory,
manifest, installazione pulita e `doctor` 9/9.

## Gate live RC4

| Run fixture | Terminale | Turni/tool | Durata | Decisione |
|---|---|---|---:|---|
| 1 | `completed` | 2 / 1 | 351.919 ms | PASS |
| 2 | `deadline_exceeded` | 4 / 0 | 600.000 ms | Limite modello bounded; conteggio azzerato |
| repeat diagnostico | `completed` | 2 / 1 | 73.766 ms | PASS 1/2 |
| consecutiva finale | `completed` | 2 / 1 | 40.313 ms | PASS 2/2 |

Ogni PASS respinge il primo turno testuale e accetta soltanto la successiva
tool call provider-level. La run in deadline respinge quattro pseudo-risposte e
non dichiara un falso successo.

Una regressione aggiuntiva su `real-laravel-a` indicizza il workspace ma
termina `provider_failure` dopo 565.353 ms e due turni senza tool call. Il caso
è classificato come limite modello/hardware già osservato, non come regressione
del prodotto. Stato Git e diff del workspace sono invariati prima e dopo. I
PASS reali della Fase 3 restano l'evidenza live del confine di indicizzazione.

## Gate repository

- `go test ./...`: PASS;
- `go test -race ./...`: PASS;
- `go vet ./...`: PASS;
- `git diff --check`: PASS;
- regressioni Laravel source policy e pseudo-call: PASS.

## Classificazione conclusiva

| Osservazione | Classe | Destinazione |
|---|---|---|
| Asset Laravel generati bloccano la v0.1.0 | Bug v0.1.x | Corretto in v0.1.1 |
| Contesto root fixture escluso da RC1 | Regressione candidate | Corretto prima della release |
| Pseudo-call incorporata o sconosciuta accettata | Bug protocollare candidate | Corretto prima della release |
| Tool call provider-level malformata | Limite modello | Fail-closed documentato |
| Inferenza oltre timeout/deadline | Limite modello/hardware | Terminale bounded documentato |
| Mutazione controllata | Richiesta evolutiva | Milestone 10–11 |

Non restano osservazioni non classificate né bug read-only bloccanti. La patch
finale v0.1.1 deve essere prodotta dallo stesso contenuto qualificato, ripetere
il gate artifact e mantenere il supporto ristretto a Linux amd64, Ollama e
reference agent read-only.
