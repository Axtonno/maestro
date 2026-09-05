# Milestone 35 — Preflight della qualificazione

Data: 2026-09-05. Esito: PASS, senza generazioni.

- modello selezionato `qwen2.5-coder:14b`, digest verificato contro Ollama;
- 30 casi congelati e indipendenti dalla selezione e da M33;
- sviluppo e holdout contengono ciascuno 15 casi: 10 positivi, 2 astensioni
  semanticamente necessarie e 3 rifiuti host-side;
- ciascun set richiede 12 chiamate provider, 10 passaggi di approvazione e 7
  applicazioni completate;
- tutte le selezioni provider-bound e le sostituzioni attese sono state
  ricostruite byte per byte prima dell'avvio;
- un solo tentativo per caso, senza retry, repair, fallback o tuning;
- approvazioni `allow`/`deny` esercitate tramite TTY reale;
- checkpoint atomico dopo ogni caso e report live creato con `O_EXCL`;
- `go test ./scripts/m35qualify` e `go vet ./scripts/m35qualify`: PASS.

Freeze SHA-256:

| Artefatto | Digest |
| --- | --- |
| Matrice di qualificazione | `e7e75d0daf3be372ac4bf311f8d02944d634573e54a62a3c96d63ee207d5db8c` |
| Schema | `bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d` |
| Prompt | `594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732` |

Il gate viene calcolato separatamente su sviluppo e holdout, oltre che sul
totale. Ogni set deve soddisfare tutti i vincoli; un singolo effetto vietato o
un terminale inatteso respinge il profilo.
