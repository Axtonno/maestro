# Milestone 9 — Report finale

Data: 2026-08-20

Stato: **COMPLETATA**

## Esito

La Milestone 9 chiude l'osservazione post-release v0.1.0, rilascia la patch
read-only v0.1.1, completa formalmente la Milestone 3 e dà GO controllato alla
Milestone 10.

| Fase | Risultato |
|---|---|
| 1 — Baseline | Contratto, artifact, profilo e classificazione congelati |
| 2 — Preflight | v0.1.0 installata fuori dal checkout; due quick start PASS |
| 3 — Workspace reali | Bug di scan riprodotto; resilienza e immutabilità verificate |
| 4 — Stabilizzazione | v0.1.1 qualificata dopo RC1–RC3 rifiutate e RC4 verde |
| 5 — Benchmark closure | Milestone 3 completata; llama.cpp non supportato |
| 6 — Audit | Release v0.1.1 verde; GO alla Milestone 10 |

## Patch v0.1.1

La patch sostituisce la scan policy Laravel generica con una policy sorgente
bounded, preserva fixture e documenti root qualificati e impedisce che
pseudo-tool-call JSON incorporate nella prosa — anche verso tool inesistenti —
completino uno step. Non aggiunge autorità o compatibilità mutativa.

L'artifact finale Linux `amd64` incorpora
`ba938abc6553bc87a89088eb6763a3e255aba4f8`, ha SHA-256
`d894568cd65c261a75212274d7ab8a45eafa950660594b6c22cc777eb8ab9cf1` e
tag annotato `v0.1.1`. È riproducibile, installabile fuori dal checkout,
supera `doctor` 9/9 e due quick start consecutivi con una vera read ciascuno.

## Decisioni permanenti

- supporto v0.1.x: Linux `amd64`, Ollama, `llama3.1:8b`, reference agent
  Laravel read-only, list/read/search;
- llama.cpp: adapter sperimentale/non supportato; nessuna matrice valida sul
  profilo corrente e nessuno skip trasformato in PASS;
- mutazioni: fuori dalla compatibility v0.1.x;
- Milestone 3: completata tramite ADR-0030;
- Milestone 10: autorizzata a iniziare, entro il confine ristretto del piano
  v0.2.0;
- Milestone 11: responsabile della qualificazione live mutativa su un profilo
  modello/hardware congelato.

Non restano attività obbligatorie della Milestone 9.
