# Milestone 14 — Final Report

Data: 2026-08-28

Stato: **COMPLETATA**

Esito: **`direct_chat_deferred`**

## Verdetto

Maestro dispone ora di due modalità read-only separate:

```text
maestro chat   -> una completion da zero o un file esplicito, nessun tool
maestro agent  -> esplorazione verificata con runtime e tool read-only
maestro run    -> alias deprecato ed esatto di agent nella serie v0.3.x
```

Il contratto direct chat supera la matrice deterministica, negativa,
streaming e anti-leak. Non esiste però una qualifica live del candidato
`qwen2.5-coder:7b`: Ollama non era attivo al preflight sul computer corrente e
la stop rule ha lasciato C0-C4 `not_run`. Per questo l'esito è deferred, non
`direct_chat_candidate`, `direct_chat_model_rejected` o
`direct_chat_contract_rejected`.

## Risultati per fase

| Fase | Commit | Esito |
|---|---|---|
| 1 — ADR/CLI | `a93c0e7` | PASS |
| 2 — profili generativi | `6e5639a` | PASS |
| 3 — chat single-file | `818209d` | PASS |
| 4 — hardening/anti-leak | `72de866` | PASS |
| 5 — live qualification | `8a35add` | chiusa per stop rule |
| 6 — audit/handoff | commit contenente questo report | PASS |

## C0–C4

| Gate | Stato |
|---|---|
| C0 epistemica senza file | NOT_RUN |
| C1 correttezza single-file | NOT_RUN |
| C2 equivalenza live streaming | NOT_RUN |
| C3 operatività live | NOT_RUN |
| C4 sicurezza live | NOT_RUN |

La sicurezza deterministica e l'immutabilità sono PASS nella Fase 4; non sono
presentate come sostituti di C4 live.

## Superficie consegnata

- schema strict v2 con profili chat/agent separati;
- `num_ctx` e thinking tri-state provider-neutral, mappati esattamente o
  rifiutati;
- loader single-file confinato, bounded e resistente a symlink e replacement;
- completion e streaming atomico senza tool/retrieval/fallback;
- envelope CLI redatto con usage, latenza e valori richiesti/effettivi;
- candidate profile e protocollo ripetibile per nuovo hardware.

## Limiti e supporto

La superficie resta development-only. La compatibility promise v0.2.0,
l'artifact storico e il reference agent supportato non cambiano. Nessuna
release, tag o capability mutativa è stata prodotta. `unknown` per context o
thinking effettivi resta distinto da un'attestazione positiva.

## Decisione successiva

Milestone 15 deve prima qualificare provider e GPU sulla nuova piattaforma,
poi creare un nuovo candidate record con digest e template osservati usando il
profilo, fixture e oracoli M14. Solo C0-C4 verdi autorizzano il passaggio al
verified agent sintetico; nessun esito di questa milestone apre Controlled
Mutation.
