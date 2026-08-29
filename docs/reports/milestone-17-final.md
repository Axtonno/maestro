# Milestone 17 — Final Report

Data: 2026-08-29

Verdetto: **NOT_RUN — VERDETTO TRATTENUTO**

## Stato

Le Fasi 1–6 sono completate e F6.4 ha superato la prequalifica Direct Chat con
qualità 4/5. La Fase 7 ha congelato il packaging candidate
`v0.3.0-pc.1`, commit
`70a9630203ccf82a4d8858a9e47b48f5333b9cbd`, SHA-256
`82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61`.

Doppio packaging, checksum, audit archive e installazione locale sono verdi.
La matrice finale sulla WSL2/Ubuntu 24.04/RTX 5070 non è ancora eseguita sullo
stesso archive; `not_run` non equivale a PASS.

## Decisione sospesa

`direct_chat_product_baseline` non è emesso. Tag, release candidate, artifact
finale e pubblicazione v0.3.0 non sono autorizzati. Il report verrà completato
soltanto dopo il gate live registrato in
`reports/milestone-17-phase-7.md`.

Qualunque verdetto positivo resterà limitato a Linux `amd64`, Ollama locale,
`qwen3.5:9b` con digest qualificato e Direct Chat read-only single-file.
Agent, retrieval, tool e mutation resteranno non qualificati.
