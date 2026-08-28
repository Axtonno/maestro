# Milestone 17 — Direct/Chat Product Baseline

Versione candidata: 0.3.0

Stato: Pianificata — autorizzata dal verdetto `verified_agent_rejected` della
Milestone 15 esclusivamente per il perimetro `direct/chat` read-only

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

# Fasi di sviluppo

## Fase 1 — Freeze del contratto e audit del candidato

- confrontare implementazione M14, ADR-0033, CLI e configurazione;
- congelare sintassi, reason code, exit code, envelope e limiti;
- eliminare divergenze documentali senza cambiare semantica durante i gate;
- registrare commit, binario, schema, prompt e fixture del candidate.

Gate: contratto unico e testabile; nessuna ambiguità fra chat e agent.

## Fase 2 — Confine del servizio Direct Chat

- verificare o isolare il servizio di completion diretta;
- provare l'assenza di Tool Runtime, Context Engine e Agent Runtime;
- rendere impossibile il fallback implicito;
- aggiungere test architetturali e di composizione per il confine.

Gate: una richiesta chat genera una sola completion provider tool-free.

## Fase 3 — Contesto esplicito single-file

- consolidare containment e risoluzione del path logico;
- verificare symlink, traversal, directory, file speciali, race e dimensione;
- fissare encoding e comportamento per file vuoto o illeggibile;
- provare digest pre/post e assenza di effetti.

Gate: soltanto il file esplicitamente autorizzato può entrare nel prompt.

## Fase 4 — Profilo e preflight

- promuovere il profilo v2 candidato a contratto v0.3.0;
- verificare modello, timeout, streaming, `num_ctx`, thinking e limiti;
- distinguere valori richiesti, osservati e non osservabili;
- aggiornare example config, doctor, installazione e troubleshooting.

Gate: configurazioni incomplete o non onorabili falliscono prima della
completion.

## Fase 5 — Streaming e osservabilità

- mantenere equivalenza semantica fra complete e stream;
- validare finish reason, risposta non vuota e limiti di output;
- registrare modello, token, durata, context e terminazione;
- provare anti-leak per prompt, sorgente, risposta, path e secret;
- mantenere output atomico sui failure di stream.

Gate: equivalenza 2/2 e nessun leak nei sink pubblicabili.

## Fase 6 — Matrice operativa e qualità reale

Eseguire prima test deterministici, poi smoke live sul ThinkPad. Congelare il
candidate prima della serie e non modificare prompt o profilo fra i tentativi.

Task rappresentativi minimi:

1. spiegazione di una classe o funzione;
2. analisi di route, controller e action;
3. spiegazione di un controller e delle dipendenze invocate;
4. suggerimento di refactoring senza applicarlo;
5. suggerimento di test senza crearli;
6. domanda senza file e contesto insufficiente;
7. file non autorizzato o fuori workspace.

Gate: almeno 4/5 dei primi cinque task sono accettabili; gli scenari 6 e 7
devono entrambi rispettare il comportamento epistemico e di sicurezza.

## Fase 7 — Productization e qualifica finale

Solo dopo tutti i gate precedenti:

- costruire il packaging candidate v0.3.0;
- verificare archive allowlist, checksum e assenza di diagnostica/raw trace;
- eseguire installazione pulita e quick start;
- trasferire il candidate byte-identico sulla nuova macchina;
- ripetere una qualifica finale breve senza tuning;
- aggiornare compatibility, installation, CLI, configuration, known issues,
  troubleshooting, roadmap e release notes;
- emettere il verdetto finale prima di tag o pubblicazione.

Gate: installazione pulita e matrice finale verdi sulla piattaforma di
qualificazione. Un failure produce un report e nessuna release.

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

