# Milestone 26 — Fasi 4–5: acquisizione diagnostica e causa

Data: 2026-09-02

Stato: **COMPLETATE — `model_profile_limit`**

## Protocollo

È stata eseguita una nuova generazione diagnostica per ciascuno dei quattro
input congelati `M25-C4`, `M25-C5`, `M25-C7` e `M25-B3`. Ogni input è stato
eseguito esattamente una volta. Questa serie non è e non viene descritta come
replay M25, perché i body originali non esistono.

Provider, modello, digest, context, thinking, temperatura, `num_predict` e
residency coincidono con M25. Request e response integrali sono conservati
soltanto nell'area privata fuori dal repository, con directory `0700` e file
`0600`.

## Risultati redatti

| Caso | HTTP | Body byte | Content | Thinking | Done reason | Output token | Regola |
|---|---:|---:|---:|---:|---|---:|---|
| M25-C4 | 200 | 4856 | non vuoto | assente | `length` | 1024 | `finish_reason_not_stop` |
| M25-C5 | 200 | 4722 | non vuoto | assente | `length` | 1024 | `finish_reason_not_stop` |
| M25-C7 | 200 | 4092 | non vuoto | assente | `length` | 1024 | `finish_reason_not_stop` |
| M25-B3 | 200 | 4086 | non vuoto | assente | `length` | 1024 | `finish_reason_not_stop` |

Tutti i body sono JSON e UTF-8 validi, dichiarano il modello richiesto, ruolo
`assistant`, `done=true`, zero tool call e usage non negativa. Nessun caso ha
content vuoto o whitespace e nessun caso contiene output soltanto in
`message.thinking`.

## Classificazione

La causa osservata è uniforme: la generazione raggiunge esattamente il limite
congelato di 1024 output token e Ollama termina con `done_reason=length`.
L'adapter propaga `length`; il validator Direct Chat applica correttamente la
regola fail-closed che richiede `stop`, esponendo il reason code pubblico
compatibile `response_invalid`.

L'evidenza non dimostra un difetto di validator, adapter o contratto provider.
I quattro payload sono correttamente non pubblicabili perché troncati; la
classificazione causale è **`model_profile_limit`**. Questo spiega la forma
osservata in M25 senza pretendere che le nuove risposte siano gli stessi body.

## Decisione operativa

P6 non autorizza una correzione adapter/validator: non è stato osservato
contenuto valido rifiutato. Restano vietati aumento seriale di `num_predict`,
nuovo candidate e modifica del prompt prima del successivo gate previsto.
