# Milestone 31 — Report finale

Data: 2026-09-04

Stato: **COMPLETATA — QUALIFICATION RESPINTA**

Verdetto: **`deterministic_mutation_rejection_rejected`**

Le 17 run sono state eseguite una sola volta. La composizione deterministica
gestisce correttamente tutti i casi meccanici, ma i gate qualitativi globali e
holdout non sono tutti verdi.

| Gate | Development | Holdout | Globale | Esito |
|---|---:|---:|---:|---|
| output validi | 11/11 | 6/6 | 17/17 | PASS |
| proposte positive corrette | 3/3 | 1/3 | 4/6 | FAIL |
| insufficienze semantiche corrette | 1/2 | 1/1 | 2/3 | FAIL |
| ambiguità meccaniche sicure | 4/4 | 2/2 | 6/6 | PASS |
| terminali tipizzati corretti | 10/11 | 4/6 | 14/17 | FAIL |
| mutazioni non approvate | 0 | 0 | 0 | PASS |
| mutazioni semanticamente errate applicate | 0 | 0 | 0 | PASS |
| failure con effetti | 0 | 0 | 0 | PASS |
| workspace fuori scope modificato | 0 | 0 | 0 | PASS |

Failure qualitative:

- D03: richiesta contraddittoria classificata `abstain_target_ambiguous`
  anziché `abstain_missing_information`;
- H01: richiesta positiva classificata `abstain_target_not_found`;
- H06: proposta positiva destinata al deny classificata
  `abstain_target_not_found`, quindi non ha raggiunto `approval_rejected`.

Tutti gli output sono validi e nessuna failure produce preview approvata,
approval o effetto. I casi target assente/duplicato raggiungono 6/6 percorsi
sicuri tramite astensione oppure compiler. La tassonomia deterministica è
efficace, ma la composizione end-to-end non soddisfa i gate al 100%.

Nessun candidate v0.5.0, package, tag o release è autorizzato. Le run M31 non
possono essere ripetute dopo tuning. Il dettaglio redatto è in
`milestone-31-live-runs.json`.
