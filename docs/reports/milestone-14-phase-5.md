# Milestone 14 — Phase 5 Report

Data: 2026-08-28

Stato: **COMPLETATA CON STOP RULE**

## Esito

Il preflight live si è arrestato prima di C0 perché Ollama non era in
esecuzione e l'API loopback configurata rifiutava la connessione. Come imposto
dal piano, Maestro non ha avviato il provider, scaricato o caricato modelli né
proseguito con prove non attribuibili.

## Risultati C0–C4

| Gate | Stato | Motivo |
|---|---|---|
| C0 — chat senza file | NOT_RUN | stop al preflight `provider_unavailable` |
| C1 — single-file correct | NOT_RUN | stop al preflight |
| C2 — streaming equivalence live | NOT_RUN | stop al preflight |
| C3 — operatività | NOT_RUN | stop al preflight |
| C4 — sicurezza live | NOT_RUN | stop al preflight |

`not_run` non è stato reinterpretato come PASS. Non esiste evidenza live per
accettare o rifiutare la qualità di `qwen2.5-coder:7b`.

## Classificazione

- contratto: matrice deterministica F4 valida, nessun failure osservato;
- modello: non valutato;
- provider/ambiente: provider non disponibile al preflight;
- harness: binario, profilo, fixture, domande e oracoli congelati;
- sicurezza: nessuna request generativa e nessuna mutazione del catalogo o del
  workspace.

Il candidate record redatto è in
`docs/reports/milestone-14-direct-chat-candidate.md`.

## Gate

**CHIUSO PER STOP RULE.** L'esito di milestone candidato è
`direct_chat_deferred`: la superficie deterministica è valida, ma nessun
candidato live è utilizzabile o qualificabile nell'ambiente corrente.
