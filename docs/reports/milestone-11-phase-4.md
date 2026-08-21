# Milestone 11 — Report Fase 4

Data: 2026-08-21

Stato: **CONCLUSA — FAIL Gate A, stop fail-fast**

## Candidato

| Campo | Valore |
|---|---|
| Versione | `v0.2.0-m11-qc.2` |
| Commit | `7e8ba62da22ad1942f3688b880922eacbec0889f` |
| Binario SHA-256 | `9870772b25f482eb4a5e539cea86e44aa19740e929c5789eab091d10c70101a3` |
| Profilo SHA-256 | `a64b7557ccd24f32bb4fb7cee7d64b630e16ec017c0776b2549d86bcd8480cac` |
| Provider/modello | Ollama / `ibm/granite4.1:8b` |

## Preflight

Il preflight sandboxed non poteva raggiungere il servizio host. La ripetizione
fuori sandbox, autorizzata e senza avviare servizi o scaricare modelli, ha
superato tutti i controlli:

- Linux `amd64`;
- Intel Core i5-8365U e 8 CPU logiche;
- 15.643 MiB RAM e 4.095 MiB swap;
- configurazione, workspace, composition, agent, tool e policy;
- Ollama con 11 capability di istanza;
- capability richieste disponibili su `ibm/granite4.1:8b`;
- detection Laravel.

Il lower bound è quindi soddisfatto. Il preflight non è stato promosso a prova
di Gate A.

## Gate A

Il Gate A richiedeva tre sequenze consecutive
`read -> result -> patch`, a temperatura zero e massimo 256 token per turno,
senza invocare il Tool Runtime.

| Tentativo | Read call | Patch call | Turni/call | Durata | Esito |
|---:|---|---|---:|---:|---|
| 1 | valida | presente ma arguments non esatti | 2 / 2 | 120.119 ms | FAIL `patch_tool_call_invalid` |
| 2 | non eseguito | non eseguito | — | — | stop fail-fast |
| 3 | non eseguito | non eseguito | — | — | stop fail-fast |

Gate A: **0 successi su 1 tentativo eseguito**. Il failure è classificato come
limite del modello sul protocollo congelato, non come errore ambientale o
mutazione del prodotto.

## Integrità

- Tool Runtime non invocato;
- nessuna approval presentata o emessa;
- zero tentativi mutativi;
- SHA-256 finale del controller identico a quello iniziale
  `4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`;
- digest aggregato workspace invariato
  `9cea0e1691a6a65d572c5a0e3b7059340e0df588482359e84c20c6c384010030`;
- cleanup dei temporanei pulito;
- report privo di prompt, response, arguments, risultati tool e path fisici.

## Evidenze

- `reports/milestone-11-gate-a.json`;
- `reports/milestone-11-gate-a.md`.

Entrambi hanno permessi `0600`.

## Gate e stop rule

- preflight interamente superato: sì;
- Gate A `3/3`: **no**;
- fixture invariata: sì;
- Gate B autorizzato: **no**;
- Gate C autorizzato: **no**.

La Fase 4 è conclusa con FAIL riproducibile. Le Fasi 5 e 6 devono contabilizzare
la mancata esecuzione senza aggirare il fail-fast.
