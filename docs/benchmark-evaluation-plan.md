# Maestro Benchmark & Evaluation Plan

Versione: 0.1.0

Stato: Pronto all'avvio

Ultimo aggiornamento: 2026-08-08

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Obiettivo

La Milestone 3 — Benchmark & Evaluation Layer deve aiutare chi usa Maestro a
valutare una configurazione completa composta da hardware, provider, modello e
plugin.

La domanda guida è:

> Con questa macchina, questo provider e questo modello, Maestro è adatto al mio
> flusso di sviluppo?

Il benchmark misura il sistema e il suo comportamento operativo. Non costruisce
classifiche assolute tra famiglie di modelli, perché diventerebbero rapidamente
obsolete e nasconderebbero l'impatto di hardware, quantizzazione, provider,
configurazione e workload.

---

# Prerequisiti

La milestone inizia dopo la chiusura della Provider Layer. Dipende in
particolare da:

- capability introspection, per conoscere le operazioni realmente disponibili;
- error semantics, per distinguere fallimenti operativi e richieste invalide;
- provider observability, per raccogliere misure senza strumentare ogni adapter
  una seconda volta;
- contratti stabili per completion, streaming, embedding e lifecycle dei
  modelli;
- Plugin Runtime e baseline Laravel per gli scenari del Developer Benchmark.

Il Benchmark Layer consuma questi contratti e non introduce semantiche provider
nel Runtime Core.

La Provider Layer è conclusa e i prerequisiti sono soddisfatti. Il manifest
`provider-smoke-benchmark-manifest.yaml` costituisce l'handoff iniziale del
Livello 1 e definisce scenari, configurazione, cleanup e redazione.

---

# Principi

- Benchmark locali, trasparenti e riproducibili.
- Risultati associati alla configurazione completa, non soltanto al nome del
  modello.
- Fixture, numero di run, warmup e condizioni ambientali sempre dichiarati.
- Report raw machine-readable separato dalla presentazione per gli utenti.
- Metriche di qualità prudenti e verificabili.
- Nessun prompt, risposta, secret o path sensibile incluso nei report per
  default.
- Capability non supportate riportate come tali, non come benchmark falliti.

---

# Famiglie di benchmark

| Area | Cosa misura | Decisione supportata |
|---|---|---|
| Provider | Conformità operativa di Ollama, llama.cpp e provider futuri | L'adapter rispetta i contratti necessari? |
| Prestazioni | Primo token, token/sec e latenza totale | Il sistema è sufficientemente reattivo? |
| Risorse | RAM, CPU, VRAM quando disponibile e costo del load | La configurazione è sostenibile sull'hardware locale? |
| Streaming | Stabilità, completamento, cancellazione e chiusura | L'esperienza interattiva e agentica è affidabile? |
| Embedding | Latenza, dimensione, coerenza e throughput | Il recupero del contesto è praticabile? |
| Task reali | Spiegazione, refactor, test e scenari Laravel | La configurazione è utile nel lavoro quotidiano? |

---

# Livello 1 — Smoke Benchmark

## Scopo

Verificare che provider, server e modelli configurati funzionino realmente.
Questo livello assorbe gli smoke test live rinviati dalla Provider Layer.

## Scenari

- Identità e capability dell'istanza.
- Model listing e discovery.
- Completion non-streaming.
- Streaming fino al terminale e chiusura delle risorse.
- Cancellazione e deadline.
- Embedding con modello compatibile.
- Load e unload con modelli dedicati.
- Pull e remove soltanto con modelli temporanei esplicitamente configurati.
- Output strutturato e tool calling quando dichiarati disponibili.

## Comando previsto

```text
maestro bench smoke
```

## Gate

- Esito distinto per capability supportata, non supportata, saltata e fallita.
- Nessuna mutazione distruttiva senza un modello fixture dedicato.
- Cancellazione e cleanup verificati per ogni stream aperto.
- Stesso schema di risultato per Ollama e llama.cpp.

---

# Livello 2 — Runtime Benchmark

## Scopo

Misurare reattività, throughput, risorse e comportamento sotto operazioni
ripetute, senza valutare ancora la qualità semantica di un task di sviluppo.

## Scenari

- Cold start e warm request.
- Time to first token e throughput dello stream.
- Latenza totale della completion.
- Load, unload e cambio di residenza del modello.
- Cancellazione durante generazione e download.
- Latenza e throughput degli embedding.
- Memoria, CPU e VRAM quando osservabili in modo affidabile.
- Errori controllati, retry e circuit breaker.

## Comandi previsti

```text
maestro bench provider
maestro bench model
```

`bench model` misura il modello selezionato nel suo profilo completo e non
costruisce confronti o classifiche automatiche tra modelli differenti.

## Gate

- Warmup e numero di run configurabili e registrati.
- Aggregati almeno con minimo, mediana, percentile 95 e massimo quando il numero
  di campioni è sufficiente.
- Misure non disponibili rappresentate come assenti, mai come zero.
- Nessuna dipendenza obbligatoria da strumenti di telemetria esterni.

---

# Livello 3 — Developer Benchmark

## Scopo

Valutare l'utilità pratica di Maestro su task riproducibili di sviluppo
software. Questo è il livello più caratterizzante del progetto.

## Dataset iniziale

- Spiegare un controller Laravel.
- Individuare le dipendenze di una service class.
- Generare un test PHPUnit da requisiti e codice fixture.
- Rifattorizzare un metodo lungo senza cambiare comportamento osservabile.
- Riassumere la struttura di un piccolo progetto.
- Usare embedding per recuperare i file rilevanti a una domanda.

Le fixture sono versionate, locali, prive di dati sensibili e sufficientemente
piccole da poter essere eseguite anche su hardware non workstation.

Il primo scenario di retrieval usa direttamente gli embedding provider sulle
fixture versionate. L'integrazione con il futuro Context Engine potrà estendere
lo stesso scenario senza bloccare la Milestone 3.

## Comando previsto

```text
maestro bench laravel
```

## Gate

- Dataset minimale versionato e documentato.
- Risultato tecnico separato dalla valutazione qualitativa.
- Rubrica 0–3 applicabile manualmente o con controlli semi-automatici.
- Nessun evaluator LLM obbligatorio nella prima versione.

---

# Metriche

| Metrica | Tipo | Significato |
|---|---|---|
| `time_to_first_token_ms` | Durata | Reattività percepita dello stream |
| `tokens_per_second` | Tasso | Velocità generativa osservata |
| `total_latency_ms` | Durata | Tempo totale dell'operazione o task |
| `peak_memory_mb` | Risorsa | Massimo consumo RAM attribuibile alla configurazione |
| `cpu_avg_percent` | Risorsa | Carico CPU medio durante lo scenario |
| `cpu_peak_percent` | Risorsa | Picco CPU durante lo scenario |
| `peak_vram_mb` | Risorsa opzionale | Massima VRAM osservata |
| `stream_completed` | Booleano | Presenza del terminale previsto |
| `cancel_latency_ms` | Durata | Tempo tra cancellazione e arresto osservato |
| `embedding_latency_ms` | Durata | Costo di una richiesta embedding |
| `success_rate` | Rapporto | Run completate senza errore sul totale |
| `quality_score` | Rubrica | Utilità pratica valutata da 0 a 3 |

Token/sec viene pubblicato soltanto quando il provider espone conteggi o quando
il metodo di stima è dichiarato. Le metriche di risorsa devono indicare il
perimetro misurato: processo provider, processo Maestro oppure intero sistema.

---

# Rubrica qualitativa

| Score | Significato |
|---|---|
| 0 | Risposta inutilizzabile o task non completato |
| 1 | Risultato parzialmente utile, con lacune sostanziali |
| 2 | Risultato utile ma bisognoso di correzioni |
| 3 | Risultato direttamente utilizzabile nello scenario fixture |

Il punteggio non è una verità assoluta e non viene aggregato in una classifica
globale dei modelli. Report e rubriche devono conservare scenario, evaluator e
motivazione.

---

# Profilo della configurazione

Ogni run registra, quando disponibili:

- versione e commit di Maestro;
- sistema operativo e architettura;
- CPU e memoria totale;
- GPU, backend e VRAM;
- provider, versione del server e URL redatto;
- modello, digest, formato, quantizzazione e context length;
- plugin e relative versioni;
- parametri di generazione rilevanti;
- stato cold/warm, warmup e numero di run;
- timestamp e versione del dataset.

Secret, prompt completi, risposte complete e path utente non sono inclusi per
default.

---

# Report

Ogni comando produce:

- un report JSON versionato, destinato ad automazione e confronti ripetibili;
- un report Markdown derivato, leggibile e adatto alla documentazione;
- exit status non-zero soltanto per errori del runner o gate richiesti, non per
  capability dichiarate non supportate.

Il formato JSON distingue metadata, configurazione redatta, scenario, campioni,
aggregati, errori classificati e valutazione qualitativa.

---

# Scomposizione della Milestone 3

| Fase | Incremento | Criterio di uscita |
|---|---|---|
| 1 | Benchmark Contracts & Runner | Manifest, scenari, sample e report JSON versionati |
| 2 | Smoke Benchmark | Matrice live provider/modello e cleanup affidabile |
| 3 | Runtime Benchmark | Prestazioni e risorse riproducibili |
| 4 | Developer Benchmark | Dataset Laravel/PHP e rubrica 0–3 |
| 5 | Reporting & Hardware Profiles | Report Markdown, profili e gate documentato |

---

# Output della milestone

```text
maestro bench smoke
maestro bench provider
maestro bench model
maestro bench laravel
```

Inoltre:

- report JSON versionato;
- report Markdown leggibile;
- profili hardware documentati;
- dataset minimale di task reali;
- istruzioni per riprodurre ogni scenario.

---

# Fuori scope iniziale

- Classifica pubblica dei modelli.
- Un singolo punteggio globale di qualità.
- Evaluator LLM obbligatorio.
- Benchmark cloud come requisito della milestone.
- Raccolta automatica o invio remoto dei risultati.
- Ottimizzazioni applicate automaticamente in base ai risultati.

Il primo obiettivo è rendere le decisioni locali più informate, non dichiarare
un modello universalmente migliore.
