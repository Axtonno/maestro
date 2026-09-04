# M29 — Classificazione offline degli arguments native tool calling

Data: 2026-09-04

Metodo: sola analisi degli artefatti M29 conservati e del percorso codice
Ollama → adapter → `DecodeTransport`. Nessuna nuova generazione e nessuna
ripetizione della matrice.

## Limite dell'evidenza

Il report redatto conserva terminale, finish reason, token e classe
`invalid_transport_output`, ma non gli arguments grezzi né una descrizione
strutturale dei loro campi. Gli otto output rifiutati non sono quindi
riclassificabili individualmente senza inventare evidenza.

| Classe proposta | Conclusione offline | Motivo |
|---|---|---|
| JSON malformato | esclusa al confine dell'adapter | `translateOllamaToolCalls` accetta soltanto `json.RawMessage` decodificabile come oggetto JSON; altrimenti la run avrebbe registrato `provider_error` |
| arguments serializzati come stringa | esclusa al confine dell'adapter | una stringa JSON non supera `jsonObject` |
| campi mancanti | possibile, non determinabile | il raw non è stato conservato |
| campi aggiuntivi | possibile, non determinabile | il raw non è stato conservato |
| tipo errato | possibile, non determinabile | il raw non è stato conservato |
| struttura di altro schema | possibile, non determinabile | il raw non è stato conservato |

L'adapter copia gli arguments accettati con `append(json.RawMessage(nil),
...)`; non effettua conversioni di forma. L'envelope costruito dall'harness
usa `json.RawMessage`, quindi non serializza nuovamente l'oggetto come stringa.
I test deterministici dimostrano inoltre che un oggetto `mutation-proposal-v1`
valido attraversa il decoder native e converge con structured output.

## Verdetto

`no_adapter_normalization_defect_evidenced`

Non è dimostrato che tutti gli otto output fossero semanticamente malformati,
ma l'evidenza persistita non mostra una rappresentazione valida prevista
dall'API Ollama e persa dall'adapter. Il trasporto resta respinto. Questa
analisi non autorizza nuove run M29.
