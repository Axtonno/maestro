# Milestone 13 — Model Candidate Qualification: `qwen3.5:9b`

Data: 2026-08-26

Stato: completata; `candidate_rejected`

## Ambito e baseline congelato

Questa qualificazione è separata dalla Field Validation ufficiale della
Milestone 13. Non modifica, sostituisce o ricalcola alcuna run precedente.

- Batch 1 resta immutabile a 5/22 run ufficiali;
- il Batch 2 Pilot è concluso e non viene riaperto;
- il progressive choreography experiment è concluso;
- `llama3.1:8b` è escluso da ulteriori esperimenti multi-file;
- un ulteriore aumento del timeout è escluso come soluzione;
- il retrieval della campagna originale non è ancora valutato;
- `v0.2.1-dev.m13.1` resta soltanto un candidate diagnostico;
- choreography e prompt non sono stati modificati durante questa fase.

Le prove dirette e Maestro di questo report restano fuori dal denominatore
ufficiale di 22.

## Identità del candidato e host

| Campo | Valore congelato |
|---|---|
| Modello | `qwen3.5:9b` ufficiale Ollama |
| Digest | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| Dimensione | 6.594.474.711 byte |
| Formato / quantizzazione | GGUF / `Q4_K_M` |
| Famiglia | `qwen35` |
| Parametri | 9.653.104.368 (`9.7B`) |
| Capability dichiarate | `completion`, `vision`, `tools`, `thinking` |
| Ollama | `0.32.14`, pacchetto Snap `v0.32.14` rev. 131 |
| Candidate Maestro | `v0.2.1-0.20260826174807-969f8e2b3ac7`, build tag `maestro_development` |
| Host | Linux `amd64`, Intel Core i5-8365U, 8 CPU logiche |
| RAM / swap | 16.403.841.024 / 4.294.963.200 byte |

Prima del caricamento erano disponibili 12.439.650.304 byte di RAM e la swap
usata era 34.492.416 byte. Dopo i gate diretti, con il modello residente a
contesto 8192, erano disponibili 4.680.482.816 byte e la swap usata era
291.901.440 byte. Non è stato osservato OOM.

I parametri incorporati nel modello sono `temperature=1`, `top_k=20`,
`top_p=0.95` e `presence_penalty=1.5`. Le richieste dirette non hanno
sovrascritto questi valori.

## Configurazione diretta congelata

| Campo | Valore |
|---|---|
| Endpoint | Ollama native `POST /api/chat` |
| `num_ctx` | 8192 |
| Temperatura | default del modello; campo omesso |
| Thinking | default del modello; campo omesso |
| Streaming | `false` |
| Timeout per chiamata provider | 600 s |
| Limite per singola run | 900 s |
| Modelfile di terze parti | nessuno |

Le ripetizioni dello stesso gate hanno usato conversazioni iniziali e tool
schema identici. Thinking e contenuto non sono stati persistiti come evidenza
grezza nel repository.

## Gate A — Tool calling elementare

Un solo tool `lookup_city_temperature` richiedeva l'argomento esatto
`{"city":"Berlin"}`. Una risposta era valida soltanto con una call nativa,
una sola call, argomenti esatti e contenuto testuale vuoto.

| Run | Esito | Durata ms | Token in/out | Osservazione |
|---|---|---:|---:|---|
| A-1 | PASS | 65.844 | 302/84 | call nativa esatta; load 11.423 ms |
| A-2 | PASS | 30.163 | 302/96 | call nativa esatta |
| A-3 | PASS | 28.337 | 302/87 | call nativa esatta |

Verdetto: **3/3**, gate superato. Non sono comparse pseudo-call testuali.

## Gate B — Continuazione dopo il risultato

Ogni run richiedeva la catena
`get_order -> risultato -> notify_customer -> risultato -> risposta finale`.
La seconda call doveva usare `customer_id=C-7` e il messaggio esatto derivati
dal primo risultato. Una risposta testuale era ammessa soltanto nel terzo
turno.

| Run | Esito | Durata ms | Token in/out | Sequenza |
|---|---|---:|---:|---|
| B-1 | PASS | 144.931 | 1.618/282 | call, call, finale |
| B-2 | PASS | 137.074 | 1.714/394 | call, call, finale |

Verdetto: **2/2**, gate superato. Ordine, nomi e argomenti sono corretti e non
sono comparse pseudo-call testuali.

## Gate C — Correzione dopo finalizzazione respinta

Il primo turno produceva la finalizzazione testuale `FINAL: ready`. Il turno
successivo comunicava `FINALIZATION_REJECTED` per evidenza mancante e
richiedeva la call nativa
`inspect_release_evidence({"release":"v0.2.1"})`, senza testo.

| Run | Esito | Durata ms | Token in/out | Osservazione |
|---|---|---:|---:|---|
| C-1 | PASS | 245.879 | 357/608 | finalizzazione testuale, poi call nativa esatta |
| C-2 | PASS | 146.448 | 357/427 | finalizzazione testuale, poi call nativa esatta |

Verdetto: **2/2**, gate superato. La variabilità deriva soprattutto dal
thinking predefinito: il primo turno delle due run ha generato rispettivamente
498 e 295 token per produrre la stessa risposta breve.

Nel complesso i gate diretti hanno richiesto 798.676 ms e hanno registrato
4.952/1.978 token cumulativi. I conteggi multi-turn includono il contesto
rivalutato da Ollama a ogni chiamata.

## Smoke Maestro e normalizzazione

Il provider pubblico Maestro è stato istanziato con timeout 10 minuti e ogni
prova aveva un limite esterno di 15 minuti. Non è stato modificato codice
Maestro.

| Prova | Esito | Durata ms | Token in/out | Evidenza |
|---|---|---:|---:|---|
| completion non-streaming | PASS | 62.374 | 16/151 | contenuto `Maestro`, terminale `stop` |
| completion streaming | PASS | 60.915 | 16/193 | 178 chunk, contenuto ricomposto `Maestro`, terminale `stop` |
| tool call non-streaming | PASS | parte di 112.695 | 277/81 | call completa e terminale normalizzato `tool_calls` |
| tool call streaming | PASS | parte di 112.695 | non raccolti separatamente | 37 chunk, JSON ricomposto e terminale normalizzato `tool_calls` |

Nome e argomenti normalizzati erano in entrambi i casi
`echo_message({"message":"Maestro smoke"})`. Il candidato passa quindi lo
smoke provider e non espone un'incompatibilità nella traduzione delle tool
call o nella normalizzazione del terminale streaming.

## Progressive choreography sintetica

Il candidate development-only congelato è stato eseguito senza modificare
choreography o prompt. Il fixture sintetico conteneva due stati applicativi
positivi:

1. route verso un controller/action;
2. action referenziata dal controller.

Poiché il protocollo esistente richiede anche `events_jobs_services`, il terzo
stato doveva essere chiuso come `unavailable` dopo una ricerca vuota. Il
profilo usava streaming, soli tool list/read/search, policy mutativa `deny`,
provider timeout 10 minuti e limite run 15 minuti. Il preflight ha superato
9/9 check.

| Campo | Risultato |
|---|---|
| Exit / terminale | `4` / `provider_failure` (`provider_unavailable` nella sintesi CLI) |
| Durata | 643.069 ms |
| Turni / tool call | 7 / 6, bootstrap incluso |
| Token in/out | 12.306/1.122 |
| Route | `covered` tramite bootstrap read |
| `controller_action` | read riuscita, dichiarazione `covered` accettata |
| `referenced_symbols` | rimasto aperto; due search vuoti e ulteriore discovery |
| `events_jobs_services` | non raggiunto |
| Risposta finale | assente |
| Workspace | invariato |

Il digest fisico pre/post del fixture coincide:
`0e607b31520636024af9786cdcad0a6d6ce5d9b96b83c06db758924738fc2805`.

Ollama è rimasto disponibile, non ha registrato OOM e tutte le richieste della
run sono terminate HTTP 200. Al settimo turno i contatori sono aumentati di
97 output token, ma la risposta normalizzata non conteneva né testo né tool
call. Il loop Maestro classifica esplicitamente questa condizione come
`provider_failure`; `provider_unavailable` è il reason code CLI comune per
questo terminale e non prova un guasto di trasporto.

La migliore diagnosi supportata è quindi un terminale thinking-only del
modello/template mentre uno stato era aperto, preceduto da discovery non
convergente. Il risultato non dimostra un difetto di normalizzazione
dell'adapter. I search vuoti non bastano invece a isolare il retrieval, perché
la telemetria redatta non conserva gli argomenti scelti dal modello.

### Scostamento di contesto osservato

L'adapter Maestro non espone `num_ctx` nelle generation options. Alla prima
richiesta Maestro, Ollama ha pertanto ricaricato il runner con il default 4096,
mentre i gate diretti erano congelati a 8192. I log indicano però 2.115 token
nel turno terminale e `truncated=0`: lo scostamento va corretto in una futura
qualificazione, ma non spiega il failure osservato.

## B01 ×2

Non eseguito. La stop rule richiedeva il superamento della choreography
sintetica prima di investire sul task Laravel multi-file. Non esiste quindi
una prima risposta B01 da valutare e la seconda ripetizione non è applicabile.

## Decisione

Decisione: **`candidate_rejected`** per il percorso Maestro congelato.

`qwen3.5:9b` supera i gate diretti A-C, gli smoke Maestro e la normalizzazione
delle tool call, ma non completa il gate agentico immediatamente precedente a
B01. Ha dimostrato un miglioramento sostanziale rispetto a `llama3.1:8b` —
call native, continuazione multi-turno, reazione al rifiuto e chiusura del
primo stato — senza però sostenere una progressione sintetica completa e
ripetibile.

Questa decisione riguarda la combinazione osservata di modello ufficiale,
template Ollama, default thinking, hardware di riferimento e choreography
Maestro. Non afferma che la famiglia Qwen 3.5 sia genericamente inadatta e non
riqualifica retroattivamente la campagna v0.2.0.

Non viene prodotto un nuovo artifact, non viene avviata una nuova Field
Validation e la matrice ufficiale resta storicamente valida e sospesa a 5/22.
