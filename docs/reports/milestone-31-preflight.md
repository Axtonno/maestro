# Milestone 31 — Preflight

Data: 2026-09-04

Stato: **PASS**

- Linux amd64 su WSL2 Ubuntu 24.04;
- Ollama 0.33.1 e `qwen3.5:9b` digest `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`;
- RTX 5070 12227 MiB;
- schema `mutation-decision-v1`, decoder strict e compiler M31: PASS;
- terminali `target_not_found`, `target_ambiguous`, `stale_source` e `protected_target` coperti deterministicamente;
- matrice 11 development + 6 holdout nuovi congelata prima delle run.
