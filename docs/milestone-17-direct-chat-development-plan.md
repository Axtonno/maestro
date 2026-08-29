# Milestone 17 — Direct/Chat Product Baseline

Versione candidata: 0.3.0

Stato: In corso — Fasi 1–5 completate; Fase 6 conclusa con stop rule
`direct_chat_candidate_failed`; Fase 7 `NOT_RUN`. Autorizzata dal verdetto
`verified_agent_rejected` della Milestone 15 esclusivamente per il perimetro
`direct/chat` read-only

Data: 2026-08-28

Documenti di riferimento:

- `roadmap.md`;
- `milestone-14-interaction-modes-direct-chat-plan.md`;
- `milestone-15-reference-hardware-readonly-baseline-plan.md`;
- `reports/milestone-14-final.md`;
- `reports/milestone-15-final.md`;
- `reports/milestone-15-phase-2.md`;
- `adr/ADR-0033.md`;
- `configuration.md`;
- `cli.md`;
- `security-model.md`;
- `compatibility.md`;
- `installation.md`;
- `troubleshooting.md`.

---

# Decisione di apertura

La Milestone 15 ha qualificato sulla piattaforma WSL2/Ubuntu 24.04/RTX 5070
la modalità `direct/chat` con `qwen2.5-coder:7b`, ma ha respinto il verified
agent prima di B01. Controlled Mutation e la precedente Milestone 16 restano
formalmente chiuse.

La Milestone 17 apre un percorso di prodotto indipendente:

```text
file esplicito
    -> controllo del workspace
    -> completion diretta
    -> risposta
```

Non tenta di correggere, aggirare o sostituire il verified agent. Promuove a
baseline utilizzabile esclusivamente il percorso già separato e qualificato
nelle Milestone 14 e 15.

Il precedente documento `milestone-17-mutation-qualification-plan.md` resta
una traccia storica non eseguibile. Il numero 17 è riassegnato da questa
decisione al Direct/Chat Product Baseline; un eventuale futuro programma
mutativo richiederà una nuova decisione e una nuova numerazione.

---

# Obiettivo operativo

Rendere installabile, documentata e supportabile la prima esperienza
quotidiana realmente utilizzabile di Maestro:

```bash
maestro chat --file routes/api.php \
  "Quali endpoint, controller e action sono dichiarati?"
```

Il comando riceve una domanda e, facoltativamente, un solo file esplicito
scelto dall'utente. Verifica il file entro il workspace configurato, costruisce
una singola richiesta di completion senza tool e restituisce la risposta.

La milestone consolida il codice candidato già prodotto dalla Milestone 14 e
qualificato live nella Milestone 15. Non crea un secondo percorso parallelo e
non trasferisce nella chat componenti dell'Agent Runtime.

## Risultato atteso

Un esito `direct_chat_product_baseline` autorizza:

- schema e comando `maestro chat` come superficie supportata di v0.3.0;
- profilo chat separato dal profilo agent;
- uso read-only single-file su workspace locali autorizzati;
- streaming esplicito, se equivalente e fail-closed;
- packaging candidate, installazione pulita e release readiness di v0.3.0.

Non autorizza verified agent, retrieval multi-file, tool calling o mutazioni.

---

# Confini non negoziabili

## Incluso

- comando `maestro chat` con domanda posizionale o stdin bounded;
- zero o un `--file` logico relativo al workspace;
- containment, symlink policy, regular-file check, encoding e limiti byte;
- completion diretta tramite la capability del provider;
- profilo chat dedicato con modello, timeout, streaming, `num_ctx`, thinking e
  limiti di input/output;
- streaming richiesto esplicitamente e semanticamente equivalente;
- envelope CLI, reason code, exit code e telemetria redatta stabili;
- matrice deterministica, operativa, qualitativa e di sicurezza;
- packaging candidate e installazione pulita dopo il superamento dei gate.

## Escluso

- tool calling;
- retrieval o indicizzazione automatica;
- Agent Runtime, planning, sessione, evidence state o choreography;
- fallback al verified agent o da agent a chat;
- selezione autonoma di file, glob, directory o multi-file;
- memoria conversazionale persistente;
- modifica, creazione, cancellazione o rinomina di file;
- approval, shell, Git, processi, Docker, Composer o Artisan;
- Controlled Mutation e riapertura della Milestone 16;
- support claim per il verified agent;
- tuning opportunistico durante una serie di qualificazione congelata.

---

# Contratto CLI

Forma canonica:

```text
maestro chat [--config <file>] [--file <logical-path>] [--stream] [question]
```

Regole:

- la domanda è un singolo argomento posizionale oppure arriva da stdin, mai da
  entrambi;
- domanda vuota, input oltre limite o opzioni incompatibili sono usage error;
- `--file` è opzionale e ripetibile zero volte o una sola volta;
- il path è logico, relativo e normalizzato senza correzioni implicite;
- path assoluti, traversal, directory, symlink evasivi, file non regolari,
  encoding invalido e file oltre limite sono respinti prima della disclosure;
- senza `--file` la completion riceve soltanto domanda e system prompt e deve
  dichiarare quando mancano informazioni per rispondere;
- `--stream` è ammesso solo se il profilo lo abilita;
- un errore dopo chunk parziali non pubblica una risposta apparentemente
  completa;
- stdout contiene soltanto l'envelope di risultato; stderr contiene usage,
  progress redatto e failure sintetici;
- nessun failure avvia un secondo percorso di esecuzione.

## Exit code

| Codice | Significato |
|---:|---|
| 0 | completion conclusa con risposta valida |
| 1 | risposta invalida, hard limit o failure interna |
| 2 | uso, configurazione o file non valido/non autorizzato |
| 4 | provider, modello, capability o deadline provider non disponibile |
| 130 | cancellazione tramite interrupt |

I failure espongono soltanto `chat failed: <reason_code>`. Prompt, contenuto del
file, risposta parziale, path fisico e credenziali non compaiono su stderr o
negli eventi operativi.

---

# Profilo dedicato

Il contratto esistente `version: 2` resta la base candidata. La forma
canonica è:

```yaml
interaction:
  chat:
    model: qwen2.5-coder:7b
    timeout: 600s
    streaming: false
    num_ctx: 4096
    thinking: "false"
    max_file_bytes: 1048576
    max_output_bytes: 1048576
```

Il provider rimane configurato nel blocco comune `provider`; non viene
duplicato sotto `interaction.chat`. La separazione richiesta è di profilo e
modello, non di adapter istanziato implicitamente.

Ogni valore è strict e validato. `num_ctx` e thinking richiesti devono essere
inoltrati all'adapter e osservabili; un adapter che non può rispettarli fallisce
il preflight. Il timeout del profilo non può superare il ceiling di trasporto.

---

# Architettura e responsabilità

Il percorso chat deve restare direttamente verificabile:

```text
CLI chat
  -> parser e validazione input
  -> loader single-file contenuto nel workspace
  -> prompt builder bounded
  -> Provider Complete oppure Stream
  -> validazione risultato e output atomico
```

Il servizio Direct Chat:

- dipende soltanto dal profilo chat, dal workspace file loader, dal provider e
  dall'event bus redatto;
- invia `Tools: []` e `ToolChoice: none` per ogni richiesta;
- non costruisce Context Engine, Tool Runtime, Agent Runtime o approver;
- non usa documenti indicizzati o contesto non selezionato dall'utente;
- applica budget prima della richiesta e hard limit durante la risposta;
- restituisce reason code tipizzati senza propagare payload sensibili.

Una dipendenza nuova verso retrieval, agent o tool runtime è una violazione
architetturale e blocca il gate, anche se i test qualitativi risultano verdi.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Freeze del contratto e audit del candidato | Completata | M14 completata + M15 `direct/chat` PASS |
| 2 | Confine del servizio Direct Chat | Completata | Fase 1 |
| 3 | Contesto esplicito single-file | Completata | Fase 2 |
| 4 | Profilo dedicato e preflight | Completata | Fase 3 |
| 5 | Streaming, terminali e osservabilità | Completata | Fase 4 |
| 6 | Matrice deterministica e qualifica sul ThinkPad | Conclusa — `direct_chat_candidate_failed` | Fase 5 |
| 7 | Packaging candidate e qualifica finale | `NOT_RUN` — non autorizzata | Fase 6 PASS |

Le fasi sono sequenziali rispetto ai gate. Una fase può preparare test o
fixture della successiva, ma questi non costituiscono evidenza e non vengono
eseguiti come serie ufficiale prima del PASS precedente. Ogni fase produce un
report autonomo sotto `docs/reports/`; la Fase 7 produce anche il report finale
della milestone.

## Regole di avanzamento

- la Fase 1 registra la baseline e il delta esatto ancora necessario; non
  presume che il codice consegnato da M14 sia già una superficie di prodotto;
- ogni modifica a codice, schema, prompt o criteri dopo il gate della fase che
  li possiede invalida quel gate e tutti i successivi;
- il freeze della Fase 6 identifica il candidate di qualificazione sul
  ThinkPad; il packaging candidate della Fase 7 è un nuovo record legato al
  commit che incorpora esclusivamente correzioni già validate;
- il ThinkPad è l'ambiente di sviluppo e prequalifica per le Fasi 1–6; la
  piattaforma WSL2/Ubuntu 24.04/RTX 5070 riceve soltanto l'archive immutabile
  della Fase 7;
- provider e catalogo modelli non vengono avviati, installati o modificati
  implicitamente dai test o da Maestro;
- un failure di sicurezza arresta la milestone; un failure funzionale richiede
  causa dimostrata, correzione e nuovo candidate record dalla fase owner;
- `skipped`, `unsupported`, `not_run`, `unknown` e risultati ottenuti con un
  candidato diverso non valgono come PASS;
- tag e pubblicazione sono vietati finché il report finale non emette
  `direct_chat_product_baseline`.

## Mappa delle responsabilità

| Area | Fase owner | Superficie principale |
|---|---:|---|
| contratto pubblico e delta M14/M15 | 1 | ADR, CLI, configurazione, baseline |
| isolamento tool-free | 2 | `internal/directchat`, composition e test architetturali |
| disclosure single-file | 3 | loader, workspace containment e prompt builder |
| profilo e capability | 4 | `internal/productconfig`, provider e doctor |
| streaming e telemetria redatta | 5 | servizio chat, CLI ed event sink |
| regressione e qualità pre-release | 6 | suite, harness live e candidate record |
| artifact e support claim v0.3.0 | 7 | packaging, installazione, compatibility e release audit |

---

# Fase 1 — Freeze del contratto e audit del candidato

Stato: Completata.

## Obiettivo

Riconciliare il codice development-only consegnato dalla Milestone 14 con
l'evidenza live della Milestone 15 e trasformare il piano in un backlog chiuso,
senza modificare ancora il support claim v0.2.0.

## Attività

- confrontare ADR-0033, `maestro chat`, schema v2, documentazione e
  comportamento effettivo di CLI e servizio;
- importare da M15 soltanto il PASS direct/chat, il modello
  `qwen2.5-coder:7b`, i parametri osservati, le fixture e gli oracoli pertinenti;
- congelare sintassi, provenienza della domanda, cardinalità di `--file`,
  `--stream`, stdout/stderr, reason code, exit code e limiti;
- inventariare le dipendenze runtime del percorso chat e assegnare ogni delta
  alle Fasi 2–5;
- registrare baseline Git, suite, race detector, vet e comportamento CLI prima
  delle modifiche;
- creare il candidate record iniziale con commit, schema, prompt template,
  fixture, piattaforma e non-garanzie;
- correggere solo contraddizioni del piano e della documentazione interna;
  ogni modifica di comportamento viene rinviata alla fase owner.

## Gate di uscita

- esiste un solo contratto testabile per domanda, file, streaming, output ed
  errori;
- chat e agent hanno autorità, profili e application graph non ambigui;
- ogni requisito della milestone ha una fase owner e un criterio osservabile;
- baseline repository-wide verde o failure preesistenti esplicitamente
  registrati senza reinterpretazione;
- nessun artifact v0.3.0, tag o support claim è stato prodotto.

## Deliverable

- candidate record iniziale e matrice requisito–fase–test;
- checklist di apertura della milestone;
- `docs/reports/milestone-17-phase-1.md`.

Gate: **superato** sul commit baseline `2759c332c8edcc66f12aa12fd219e32dff3e1dba`.

---

# Fase 2 — Confine del servizio Direct Chat

Stato: Completata.

## Obiettivo

Provare strutturalmente che `maestro chat` esegue una sola generazione
provider tool-free e non può entrare nel percorso agentico, di retrieval o di
mutazione.

## Attività

- auditare `internal/directchat`, il composition root e il comando CLI per
  dipendenze dirette e transitive non ammesse;
- isolare le interfacce minime per provider, file loader ed eventi redatti;
- inviare sempre `Tools: []` e `ToolChoice: none` e rifiutare risposte con tool
  call o terminali incompatibili;
- impedire la costruzione di Agent Runtime, Context Engine, Tool Runtime,
  sessione, index e approver nel graph chat;
- eliminare fallback impliciti dopo timeout, risposta vuota, stream fallito o
  capability assente;
- aggiungere spy e test di composizione che contino request, factory e servizi
  istanziati;
- verificare che i percorsi `maestro agent`/`run` esistenti non cambino come
  effetto collaterale.

## Gate di uscita

- ogni richiesta valida produce esattamente una request provider e zero tool;
- nessun servizio agentico, di retrieval o mutativo è costruito o invocato;
- ogni failure termina nello stesso percorso con reason code tipizzato;
- test architetturali, di composizione e regressione agent sono verdi;
- workspace byte-identico per success e failure.

## Deliverable

- test del confine applicativo e del composition graph;
- inventario delle dipendenze ammesse e vietate;
- `docs/reports/milestone-17-phase-2.md`.

Gate: **superato** sulla baseline di Fase 1 `b1c85e4`.

---

# Fase 3 — Contesto esplicito single-file

Stato: Completata.

## Obiettivo

Garantire che soltanto il singolo file nominato dall'utente, verificato entro
il workspace autorizzato, possa essere divulgato al provider.

## Attività

- consolidare normalizzazione del path logico, containment e apertura relativa
  alla root già verificata;
- coprire path assoluti, traversal, segmenti ambigui, directory, file assente,
  FIFO/device e file non regolari;
- definire e provare la policy per symlink interni ed evasivi e per root o file
  sostituiti durante la lettura;
- imporre limite byte prima della costruzione del prompt e definire encoding,
  BOM, file vuoto e contenuto illeggibile;
- delimitare domanda, path logico e contenuto non fidato senza interpretare il
  file come istruzione di sistema;
- attestare digest e identità pre/post e verificare l'assenza di scritture,
  file temporanei e metadata changes;
- aggiungere test che provino zero I/O provider per ogni input respinto.

## Gate di uscita

- un solo file regolare, esplicitamente autorizzato e bounded entra nel prompt;
- ogni evasione o race coperta fallisce prima della disclosure;
- senza `--file` non avvengono discovery, scan, retrieval o letture workspace;
- fixture e workspace sono byte-identici in tutti gli scenari;
- matrice positiva e negativa del loader interamente verde.

## Deliverable

- matrice versionata di path, symlink, tipo, encoding, size e race;
- test del prompt boundary e dell'assenza di provider I/O sui rifiuti;
- `docs/reports/milestone-17-phase-3.md`.

Gate: **superato** sulla baseline di Fase 2 `7251326`.

---

# Fase 4 — Profilo dedicato e preflight

Stato: Completata.

## Obiettivo

Promuovere il profilo chat dello schema v2 a contratto v0.3.0 strict e
fallire prima della completion quando provider o modello non possono onorarlo.

## Attività

- congelare modello, timeout, streaming, `num_ctx`, thinking,
  `max_file_bytes` e `max_output_bytes`, inclusi range e default;
- verificare strict parsing, campi sconosciuti, duplicati, valori zero e
  combinazioni incompatibili;
- mantenere provider e trasporto nel blocco comune senza duplicare o selezionare
  implicitamente adapter;
- mappare le opzioni provider-neutral nella request Ollama e distinguere valore
  richiesto, valore effettivo attestabile e valore non osservabile;
- estendere preflight e `doctor` per profilo mancante, modello assente,
  capability non supportata e timeout oltre il ceiling di trasporto;
- provare che un profilo agent valido non supplisce a un profilo chat invalido
  e viceversa;
- aggiornare example config e bozze di configuration, installation e
  troubleshooting senza dichiarare ancora v0.3.0 rilasciata.

## Gate di uscita

- config incomplete o non onorabili falliscono prima di file disclosure e I/O
  generativo;
- request catturate contengono gli esatti valori configurati;
- nessuna opzione viene ignorata o trasformata silenziosamente;
- backward compatibility v1/v2 coincide con ADR-0033 e con i test;
- doctor, config suite e mapping Ollama sono verdi.

## Deliverable

- contratto del profilo chat v0.3.0 e configuration example candidato;
- matrice capability/preflight e reason code;
- `docs/reports/milestone-17-phase-4.md`.

Gate: **superato** sulla baseline di Fase 3 `03d3f62`.

---

# Fase 5 — Streaming, terminali e osservabilità

Stato: Completata.

## Obiettivo

Rendere complete e stream semanticamente equivalenti, bounded e osservabili
senza pubblicare dati sensibili o output parziali ingannevoli.

## Attività

- congelare la semantica di `--stream` e mantenerla opt-in tramite profilo e
  flag espliciti;
- validare finish reason, risposta non vuota, usage e limite di output sia per
  complete sia per stream;
- usare buffering bounded o un envelope esplicito che impedisca di presentare
  come completa una risposta seguita da failure;
- distinguere deadline, cancellazione, provider failure, response invalid e
  limit exceeded con exit code e reason code congelati;
- gestire SIGINT/SIGTERM e shutdown entro limite senza secondo tentativo;
- emettere soltanto modello, durata, token, context, thinking e terminale
  redatti quando realmente osservabili;
- scandire stdout, stderr, eventi, log e report contro prompt, contenuto file,
  response completa, path fisici, secret e identificatori non necessari;
- provare con provider controllato equivalenza e failure a ogni confine di
  chunk.

## Gate di uscita

- complete/stream equivalenti 2/2 per contenuto semantico e terminale;
- output atomico o esplicitamente incompleto per ogni failure di stream;
- cancellazione, deadline e hard limit sono distinti e bounded;
- nessun sink pubblicabile contiene payload o path proibiti;
- suite streaming, terminali e anti-leak interamente verde.

## Deliverable

- harness di equivalenza complete/stream e fault injection;
- matrice terminale–reason code–exit code–output;
- `docs/reports/milestone-17-phase-5.md`.

Gate: **superato** sulla baseline di Fase 4 `9505e16`.

---

# Fase 6 — Matrice deterministica e qualifica sul ThinkPad

Stato: Conclusa con stop rule — `direct_chat_candidate_failed`.

## Obiettivo

Chiudere la prequalifica funzionale, operativa, di sicurezza e qualitativa su
un candidato immutabile prima di costruire l'archive destinato alla piattaforma
finale.

## Attività

- eseguire prima suite completa, ripetizione mirata, race detector, vet,
  controlli CLI, immutabilità e scansione anti-leak;
- congelare commit, binario, provider, modello, digest, template, profilo,
  prompt, fixture, oracoli, hardware, timeout e limiti;
- eseguire sul ThinkPad C0 senza file, C1 single-file 3/3, equivalenza stream
  2/2 e scenari operativi senza tuning fra i tentativi;
- registrare latenza cold/warm, token, context, thinking, memoria, terminale e
  reason code senza conservare payload proibiti;
- eseguire i cinque task qualitativi rappresentativi e classificare ogni
  risposta contro un oracolo predefinito;
- verificare separatamente comportamento epistemico senza file e rifiuto di un
  file non autorizzato;
- applicare fail-fast su falsità materiale, mutazione, fallback, leak, panic,
  output vuoto o terminale ambiguo.

Task qualitativi minimi:

1. spiegazione di una classe o funzione;
2. analisi di route, controller e action;
3. spiegazione di un controller e delle dipendenze invocate;
4. suggerimento di refactoring senza applicarlo;
5. suggerimento di test senza crearli.

## Gate di uscita

- suite deterministica, race detector, vet, immutabilità e anti-leak verdi;
- C0, C1 `3/3` ed equivalenza streaming `2/2` verdi sul candidate congelato;
- almeno 4/5 task qualitativi accettabili e zero falsità materiale nei PASS;
- domanda senza contesto e file non autorizzato rispettano comportamento
  epistemico e sicurezza;
- candidate record completo, immutabile e idoneo al packaging.

## Deliverable

- candidate record di prequalifica e matrice deterministica/live redatta;
- oracoli e conteggi dei task qualitativi;
- `docs/reports/milestone-17-phase-6.md`.

Gate: **non superato** sul candidate
`88c4fcbca00a0dbf77d7b7a0d7607dd19c6d8bbe`. Provider, modello, preflight,
build riproducibile e C0 sono verdi; C1 è 0/3 rispetto all'oracolo esatto,
l'equivalenza streaming è 0/2 come gate e la qualità è 2/5. La serie è stata
eseguita sulla piattaforma WSL2/RTX 5070 anziché sul ThinkPad prescritto. Il
verdetto `direct_chat_candidate_failed` non autorizza il packaging.

---

# Fase 7 — Packaging candidate e qualifica finale

Stato: `NOT_RUN` — dipendenza Fase 6 non soddisfatta.

## Obiettivo

Dimostrare che l'esatto prodotto installabile conserva i PASS precedenti sulla
piattaforma finale e chiudere la decisione v0.3.0 prima di tag o pubblicazione.

## Attività

- incorporare soltanto il codice e la documentazione già verdi e costruire due
  volte un archive v0.3.0 byte-riproducibile;
- verificare manifest, checksum, archive allowlist e assenza di fixture private,
  diagnostica, raw trace, profili mutativi e materiale development-only;
- installare fuori dal checkout in un ambiente pulito e verificare `version`,
  help, configuration example, `doctor` e quick start chat;
- trasferire lo stesso archive e checksum sulla piattaforma
  WSL2/Ubuntu 24.04/RTX 5070 senza rebuild;
- ripetere una matrice finale breve: no-file, single-file, stream/non-stream,
  containment, cancel/deadline, immutabilità e anti-leak;
- confrontare artifact, modello, digest, profilo e parametri con il candidate
  record senza tuning o sostituzioni;
- aggiornare compatibility, installation, CLI, configuration, known issues,
  troubleshooting, roadmap e release notes candidate;
- emettere il report finale e soltanto dopo un PASS autorizzare tag e
  pubblicazione separati.

## Gate di uscita

- doppio packaging byte-identico, checksum e archive audit verdi;
- installazione pulita e quick start eseguibili dalla sola documentazione;
- matrice finale verde sullo stesso archive e sulla piattaforma dichiarata;
- support claim limitato a direct/chat read-only single-file, con agent,
  retrieval e mutation esplicitamente non qualificati;
- report finale con uno dei verdetti ammessi e nessun claim anticipato.

## Deliverable

- `docs/reports/milestone-17-phase-7.md`;
- `docs/reports/milestone-17-final.md`;
- `docs/releases/v0.3.0.md` in stato coerente con il verdetto;
- manifest e checksum del packaging candidate conservati secondo la policy del
  progetto.

---

# Matrice minima di accettazione

| Gate | Criterio |
|---|---|
| Nessun file | dichiara correttamente quando mancano le informazioni; nessuna invenzione materiale |
| File esplicito | risposta corretta 3/3 sul candidate congelato |
| Streaming | semanticamente equivalente al non-streaming 2/2 |
| Sicurezza | nessuna mutazione; path esterni, traversal e symlink evasivi respinti |
| Stabilità | nessun panic, output vuoto, fallback o terminale ambiguo |
| Qualità reale | almeno 4/5 task rappresentativi accettabili |
| Anti-leak | nessun contenuto sorgente, prompt, response, path fisico o secret nei log |
| Installazione | candidate installato da archive in ambiente pulito e quick start verde |

Ogni gate è fail-fast. `skipped`, `not_run`, `unsupported` e `unknown` non
equivalgono a PASS. Una modifica a codice, prompt, schema, modello, digest,
profilo o fixture invalida i risultati successivi al freeze.

---

# Ambienti e uso dell'hardware

Il ThinkPad è autorizzato per sviluppo, test deterministici e smoke live. I
suoi risultati non sostituiscono la qualifica finale e non riaprono una ricerca
seriale di modelli.

La piattaforma WSL2/Ubuntu 24.04/RTX 5070 resta sospesa durante lo sviluppo.
Viene riutilizzata soltanto quando esiste un packaging candidate finale,
immutabile e già verde sui gate precedenti.

Il candidate iniziale resta `qwen2.5-coder:7b`, digest e parametri qualificati
dalla Milestone 15. Un cambio di modello richiede una nuova decisione e una
nuova serie completa.

---

# Evidenza e report

Ogni fase produce un report sotto `docs/reports/` con:

- commit e digest del candidate;
- configurazione e limiti effettivi;
- piattaforma e provider osservati;
- matrice eseguita, conteggi e terminali;
- token, durata, context e thinking redatti;
- digest del workspace pre/post;
- failure e reason code senza payload sensibili;
- decisione PASS, FAIL o NOT_RUN per il gate successivo.

I report non conservano prompt completi, response complete, contenuto dei file,
tool arguments, path fisici o credenziali. Non esistono tool arguments nel
percorso chat; la loro presenza è di per sé un failure.

---

# Stop rule e verdetti

| Verdetto | Conseguenza |
|---|---|
| `direct_chat_product_baseline` | autorizza release readiness e possibile v0.3.0 read-only |
| `direct_chat_candidate_failed` | nessun tag o release; correggere causa dimostrata e creare un nuovo candidate |
| `environment_blocked` | preservare evidenza e ripetere soltanto dopo il ripristino dell'ambiente |
| `security_gate_failed` | fail-closed; nessuna eccezione o riduzione del gate |

Nessun verdetto della Milestone 17 modifica lo stato di Controlled Mutation o
qualifica il verified agent.

---

# Definition of Done

La Milestone 17 è completata soltanto quando:

- il comando e il profilo sono documentati come contratto supportato;
- il confine tool-free e senza fallback è provato da test;
- containment, limiti, streaming, reason code e anti-leak sono verdi;
- la matrice qualitativa raggiunge almeno 4/5;
- il candidate è installato e verificato in ambiente pulito;
- la qualifica finale sulla nuova macchina è verde;
- documentazione, compatibility matrix e release readiness sono coerenti;
- il report finale emette un verdetto esplicito;
- soltanto dopo il verdetto positivo vengono autorizzati tag e pubblicazione di
  v0.3.0.
