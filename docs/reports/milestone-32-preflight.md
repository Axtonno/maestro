# Milestone 32 — Preflight

Data: 2026-09-04

- matrice congelata prima delle generazioni: 11 development + 7 holdout;
- holdout nuovo, non riusato da M30/M31;
- schema e decoder binari strict verificati;
- test mirati mutation/tool/harness e contract test verdi;
- Ollama 0.33.1 e `qwen3.5:9b` digest `6488c96fa5fa…` coincidenti;
- stdin TTY obbligatoria e choreography allow/deny fissata;
- nessun retry, repair, fallback o tuning post-run consentito.

Il preflight non ha consumato generazioni. La matrice è stata eseguita una
sola volta dopo questi controlli.
