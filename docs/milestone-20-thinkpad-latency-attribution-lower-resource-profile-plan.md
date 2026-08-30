# Milestone 20 — ThinkPad Latency Attribution & Lower-Resource Profile

Versione osservata: v0.3.0

Stato: **COMPLETATA — candidate non promosso**

Data: 2026-08-30

Documenti di riferimento:

- `milestone-19-post-release-adoption-lower-bound-validation-plan.md`;
- `reports/milestone-19-thinkpad-adoption.md`;
- `milestone-17-direct-chat-development-plan.md`;
- `reports/milestone-17-phase-6.md`;
- `reports/milestone-17-final.md`;
- `releases/v0.3.0.md`;
- `compatibility.md`;
- `known-issues.md`.

## Decisione di apertura

La Milestone 19 dimostra che l'asset pubblico v0.3.0 è corretto e read-only
quando completa sul ThinkPad T490s CPU-only, ma non è pratico nel ciclo
interattivo: 3/5 completion single-file, due deadline da 300 secondi e
mediana delle sole completion pari a 91,4 secondi.

Questa milestone risponde a due domande circoscritte:

1. la latenza osservata appartiene principalmente a modello/hardware oppure
   Maestro introduce overhead materiale rispetto a Ollama?
2. `qwen2.5-coder:7b` può costituire un profilo development-only Direct Chat
   più adatto al ThinkPad senza perdere qualità, sicurezza o affidabilità?

Non vengono riaperti verified agent, retrieval, multi-file, tool calling o
Controlled Mutation. Nessun esito modifica retroattivamente v0.3.0 o amplia
automaticamente il suo support claim.

## Evidenza storica da non reinterpretare

`qwen2.5-coder:7b` ha già superato un gate Direct Chat ristretto nella
Milestone 15 su RTX 5070, ma i candidate F6.1-F6.3 della Milestone 17 non hanno
raggiunto la soglia qualitativa finale: il risultato stabile era 2/5. La
Milestone 20 non cancella quell'evidenza e non presume che l'uso del modello
con Continue equivalga a una qualifica Maestro.

Un eventuale PASS su questa matrice stabilisce soltanto un candidato locale
per il perimetro congelato. Qualunque promozione a profilo distribuito richiede
una decisione separata, riconciliazione con i failure M17 e una nuova qualifica
di prodotto completa.

## Confine congelato

Sono inclusi esclusivamente:

- l'esatto asset pubblico v0.3.0 e il suo profilo ufficiale
  `qwen3.5:9b` per l'attribuzione;
- chiamate Ollama loopback e chiamate Maestro `direct/chat` semanticamente
  equivalenti;
- `qwen2.5-coder:7b` come unico candidato development-only, soltanto dopo un
  verdetto hardware/model-bound della Fase A;
- zero o un file esplicito, streaming e non-streaming;
- durata end-to-end, tempo al primo byte/chunk dove osservabile, tempo al
  terminale, usage, terminale, qualità e immutabilità;
- tre correzioni indipendenti: errori di configurazione specifici, identità
  del binario e heartbeat di progresso redatto.

Restano esclusi tuning del prompt, retry opportunistici, fallback, nuovi
provider, download impliciti, agent, tool, retrieval, multi-file, write/patch
e qualsiasi modifica del progetto esaminato. Il pull esplicito del candidato,
se non già presente, richiede una decisione operativa separata e non può
avvenire durante una serie congelata.

Le evidenze locali conservano soltanto metadata e giudizi necessari. Prompt,
risposte complete, contenuti sorgente, path fisici e remote di un eventuale
progetto reale non entrano nei report committati; gli artefatti grezzi locali
usano permessi `0600`.

## Fase A — Attribuzione della latenza

Stato: **COMPLETATA — `model_hardware_bound`**. Evidenza in
`reports/milestone-20-phase-a.md`.

### Candidate e task

Prima della prima generation vengono registrati SHA-256 di archive e binario,
versione/commit incorporati, versione Ollama, modello e digest, temperatura,
`num_ctx`, thinking, stream, timeout, `keep_alive`, stato di residenza del
modello e digest del workspace.

Si congelano due task:

- `M20-A0`: domanda senza file con oracolo di insufficienza del contesto;
- `M20-A1`: domanda single-file con risposta sintetica verificabile.

Per ciascun task si eseguono due ripetizioni direttamente contro `/api/chat`
di Ollama e due attraverso l'esatto binario Maestro v0.3.0. L'ordine è
contro-bilanciato e registrato prima della serie. Non sono ammessi rerun
selettivi; warm-up, se necessario, è identico, dichiarato ed escluso dalle
metriche.

### Equivalenza delle richieste

Il confronto non usa prompt "simili". Un relay loopback diagnostico cattura
soltanto struttura e digest del body emesso da Maestro, inoltra la richiesta a
Ollama e permette il replay diretto dello stesso body byte-identico, salvo un
eventuale identificatore diagnostico fuori dal payload. Il relay non conserva
contenuti in chiaro nel report e non modifica risposta, buffering o timeout.

Modello, messaggi, file incorporato, opzioni generative e flag `stream` devono
coincidere. Se non è possibile dimostrarlo, la coppia è `invalid_comparison` e
non produce un verdetto di attribuzione.

### Metriche

Per ogni run si registrano:

- durata da invocazione a primo byte/chunk del relay;
- durata da invocazione al terminale Ollama;
- durata fino all'output visibile di Maestro;
- load, prompt-eval ed eval duration quando restituite da Ollama;
- token in/out, terminale, exit code e reason code;
- RSS e carico CPU con campionamento redatto, senza process arguments;
- stato e digest pre/post del workspace.

La differenza primaria è il tempo appaiato
`Maestro terminale - Ollama terminale`. Il tempo al primo chunk diagnostica il
buffering ma non modifica il contratto atomico di stdout di v0.3.0.

### Regole decisionali

Con sole due ripetizioni il risultato è diagnostico, non una stima
statistica generale. Si applicano le seguenti regole predefinite:

| Segnale osservato | Decisione |
|---|---|
| entrambi i percorsi completano e l'overhead appaiato di Maestro resta entro il maggiore tra 5 s e 15% per almeno 3/4 coppie, senza asimmetria terminale | `model_hardware_bound` |
| Maestro supera il maggiore tra 5 s e 15% per almeno 3/4 coppie | `maestro_overhead_detected`; profilare adapter, relay escluso e buffering |
| rapporto max/min oltre 1,5 oppure escursione oltre 30 s in entrambi i percorsi sullo stesso task | `cpu_profile_unstable`; nessuna promozione del profilo |
| Maestro raggiunge la deadline mentre il replay Ollama appaiato completa, a parità di body e deadline | `maestro_deadline_defect`; isolare il difetto software |
| payload non equivalenti, terminali inconfrontabili o campioni mancanti | `attribution_inconclusive` |

Una differenza sotto soglia non dimostra overhead zero; dimostra soltanto che
Maestro non è la causa materiale delle latenze osservate in questa matrice.

## Fase B — Candidato lower-resource

Stato: **COMPLETATA — `thinkpad_profile_candidate`**. Evidenza in
`reports/milestone-20-phase-b.md`.

La Fase B è autorizzata soltanto da `model_hardware_bound`. Con qualunque
verdetto software o inconclusivo, la milestone resta sul percorso diagnostico
e non cambia modello.

### Candidate congelato

```yaml
interaction:
  chat:
    model: qwen2.5-coder:7b
    timeout: 5m
    streaming: true
    num_ctx: 4096
    thinking: "false"
    max_file_bytes: 1048576
    max_output_bytes: 1048576
```

Provider, prompt, temperatura, limiti, fixture e oracoli restano quelli del
candidate v0.3.0 applicabili alla matrice. Modello e digest vengono registrati
prima della prima generation. Nessun tuning è ammesso dopo il freeze.

### Matrice minima

| Gate | Requisito |
|---|---|
| no-file | 3/3 completion valide; insufficienza dichiarata e zero claim inventati |
| single-file | 5/5 completion entro 300 s sugli stessi cinque task congelati |
| stream/non-stream | almeno una coppia 2/2 semanticamente equivalente e con terminale valido |
| qualità | almeno 4/5 task single-file `correct` o `acceptable`; nessun claim materiale falso in un PASS |
| sicurezza | zero tool, retrieval, fallback e mutazioni; digest pre/post identico |
| affidabilità | nessun timeout, panic, output vuoto o terminale ambiguo |
| latenza | sugli stessi cinque task, miglioramento della mediana rispetto a `qwen3.5:9b` di almeno 30% e almeno 20 s |

Prima del verdetto si esegue sugli stessi cinque prompt/file una baseline
`qwen3.5:9b` con identico binario, profilo, ordine logico e stato di residenza
dichiarato. Le mediane M19 — 91,4 secondi per le sole completion e 164,9
secondi includendo le cinque attese timeout-capped — restano ancore storiche,
non denominatori per confrontare task differenti. Una risposta più breve ma
qualitativamente insufficiente non conta come miglioramento.

### Verdetti

| Verdetto | Conseguenza |
|---|---|
| `thinkpad_profile_candidate` | candidato development-only dimostrato; nessuna modifica a v0.3.0 |
| `lower_resource_quality_failed` | latenza eventualmente migliore, ma modello non promuovibile |
| `lower_resource_operability_failed` | timeout, instabilità o latenza ancora impratica |
| `lower_resource_security_failed` | stop immediato e nessuna eccezione |

Se il candidato fallisce, il risultato corretto è che questa milestone non ha
trovato un profilo ThinkPad qualificabile. Non è una prova universale che ogni
uso di Maestro richieda GPU; per il profilo ufficiale osservato, la
raccomandazione operativa resta hardware accelerato.

## Fase C — Correzioni indipendenti

Stato: **COMPLETATA — `operational_corrections_ready`**. Evidenza in
`reports/milestone-20-phase-c.md`.

Questa fase non entra nel confronto A/B e non può alterarne binario, payload o
misure. Le modifiche, se implementate, usano un candidate development separato
e ripetono i gate deterministici owner.

### C1 — Errori di configurazione specifici

- distinguere almeno file/config non leggibile, YAML invalido, chiave ignota,
  campo mancante e valore non valido;
- esporre il path logico del campo, mai il valore sensibile o l'errore remoto;
- mantenere exit code 2 e reason code pubblici compatibili, aggiungendo un
  dettaglio redatto e allowlisted;
- coprire CLI e loader con test table-driven e anti-leak.

### C2 — Identità chiara del binario

- rendere osservabili percorso di invocazione risolto, versione, stato,
  commit e SHA-256 durante un comando diagnostico esplicito;
- non cambiare l'output normale di `maestro chat`;
- documentare la verifica dell'installazione quando più binari esistono nel
  `PATH`.

### C3 — Feedback di progresso redatto

- mantenere stdout atomico e riservato al risultato finale;
- emettere su stderr heartbeat bounded a cadenza fissa soltanto durante la
  generation lunga, con stato ed elapsed time allowlisted;
- non includere prompt, risposta parziale, contenuto/path del file,
  credenziali, error string remota o process arguments;
- garantire un solo terminale, cancellazione pulita e assenza di goroutine o
  ticker residui tramite test deterministici.

## Sequenza e stop rule

| Fase | Stato iniziale | Gate di uscita |
|---:|---|---|
| A | completata — `model_hardware_bound` | attribuzione esplicita o `attribution_inconclusive` |
| B | completata — `thinkpad_profile_candidate` | matrice completa e uno dei quattro verdetti lower-resource |
| C | completata — `operational_corrections_ready` | test, anti-leak e report per C1-C3 |

Un failure di sicurezza arresta immediatamente la serie interessata. Un
failure qualitativo non viene convertito in PASS attraverso reinterpretazione
post-hoc. Ogni cambio a modello, digest, prompt, parametri, file o oracolo crea
un nuovo candidate e invalida le misure successive al freeze.

## Definition of Done

La Milestone 20 è completata soltanto quando:

- il confronto Ollama/Maestro usa payload dimostrabilmente equivalenti;
- le otto run minime di Fase A sono registrate senza retry selettivi;
- esiste un verdetto di attribuzione esplicito;
- Fase B è eseguita integralmente oppure marcata `NOT_AUTHORIZED` dal verdetto
  di Fase A;
- qualità, latenza, terminali e immutabilità sono riportati senza nascondere i
  timeout;
- le correzioni C1-C3 restano separate dalle misure e hanno un esito esplicito;
- roadmap, context e report finale concordano sul support claim invariato.

## Chiusura

La milestone chiude con `model_hardware_bound`,
`thinkpad_profile_candidate` e `operational_corrections_ready`. Il modello non
viene promosso: riconciliazione M17/M20, versione Ollama, gate cold/warm,
soglie assolute e artifact qualification passano alla Milestone 21. Il report
finale è `reports/milestone-20-final.md`.
