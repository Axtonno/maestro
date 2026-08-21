# Milestone 11 — Report Fase 5

Data: 2026-08-21

Stato: **CONCLUSA — Gate B non eseguito per stop rule**

## Decisione

Gate B richiede come precondizione Gate A `3/3`. Il candidato
`v0.2.0-m11-qc.2` ha ottenuto 0 successi su 1 tentativo Gate A eseguito; il
primo failure ha arrestato la serie e impedisce l'avvio dei gate successivi.

Eseguire comunque le due run read-only produrrebbe un dato diagnostico fuori
protocollo, non un'evidenza della Milestone 11. Nessuna run Gate B è stata
quindi avviata.

## Contabilità

| Tentativo | Stato | Motivo |
|---:|---|---|
| 1 | non eseguito | dipendenza Gate A non soddisfatta |
| 2 | non eseguito | dipendenza Gate A non soddisfatta |

Gate B non è `passed`, `failed`, `skipped` o `unsupported`: non ha campioni
perché la sua precondizione sequenziale non è stata soddisfatta. Il report Gate
A conserva l'evidenza che ha attivato la stop rule.

## Invarianti

- zero I/O provider aggiuntivo;
- zero fixture Gate B materializzate;
- zero tool call e zero effetti attribuiti a Gate B;
- nessun timeout, prompt o criterio modificato;
- nessun risultato storico Gate B riutilizzato come evidenza corrente.

## Gate

- precondizione Gate A `3/3`: **non soddisfatta**;
- Gate B `2/2`: **non eseguito**;
- Gate C autorizzato: **no**.

La Fase 5 è contabilizzata e conclusa senza aggirare il fail-fast. La Fase 6
deve registrare la stessa stop rule per il percorso mutativo completo.
