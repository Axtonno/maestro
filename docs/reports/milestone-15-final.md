# Milestone 15 — Report finale

Data: 2026-08-28

Esito: **`verified_agent_rejected`**

La piattaforma WSL2/Ubuntu 24.04/RTX 5070 è qualificata per provider, GPU e
direct/chat. `qwen2.5-coder:7b` supera C0 3/3, C1 3/3 e due coppie
streaming/non-streaming, con workspace invariato e offload GPU completo.

Il verified agent `qwen3.5:9b` non supera la prima progressione live: dopo la
route verificata termina con `tool_failure`. Il failure non dipende da OOM,
CPU fallback, timeout o mutazione del workspace.

## Conseguenze

- B01 è `NOT_RUN` per stop rule;
- il baseline read-only multi-file non è qualificato;
- v0.3.0 non viene costruita, taggata o pubblicata;
- Controlled Mutation resta non supportata;
- Milestone 16 resta chiusa;
- v0.2.0 conserva soltanto il proprio perimetro storico.

La milestone è conclusa con un esito ammesso dal piano, non con una release.

