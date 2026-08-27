# Milestone 14 — Interaction Modes & Direct Chat Plan

Versione: 0.2.0

Stato: In corso — Fase 1 completata, Fase 2 pronta

Data: 2026-08-27

Documenti di riferimento:

- `roadmap.md`;
- `reports/milestone-13-field-validation.md`;
- `reports/milestone-13-direct-chat-diagnostic.md`;
- `compatibility.md`;
- `known-issues.md`;
- `configuration.md`;
- `cli.md`;
- `security-model.md`.

---

# Obiettivo operativo

Separare formalmente due modalità di interazione read-only:

```text
maestro chat   -> contesto esplicito, nessun tool
maestro agent  -> esplorazione verificata, tool read-only e state machine
```

La milestone non costruisce un secondo agente e non rende più permissivo il
reference agent. Introduce una superficie `direct/chat` per domande
circoscritte quando l'utente ha già selezionato il contesto, mantenendo il
percorso agentico separato per discovery e sintesi multi-file.

Il primo modello candidato è `qwen2.5-coder:7b`, già configurato localmente in
Continue. La sua presenza non costituisce evidenza di qualifica: deve superare
gate Maestro propri, senza ereditarne risultati o impressioni d'uso.

---

# Contratto candidato

L'interfaccia iniziale è:

```text
maestro chat \
  --file routes/api.php \
  "Quali endpoint, controller e action sono dichiarati?"
```

Il comando `maestro agent` rappresenta la modalità verificata. L'eventuale
relazione con l'attuale `maestro run` viene decisa in un ADR: alias e
deprecation non devono cambiare implicitamente semantica o autorità.

## `maestro chat`

- riceve la domanda e almeno zero o un file esplicitamente selezionato;
- risolve il path soltanto dentro il workspace configurato;
- applica containment, symlink policy, limiti byte ed encoding prima della
  disclosure al modello;
- costruisce una singola completion senza dichiarare tool;
- non interroga retrieval, index o state machine;
- non effettua fallback al verified agent;
- usa prompt e profilo modello dedicati;
- registra `num_ctx` e `thinking` richiesti ed effettivi;
- espone durata, token, terminale e reason code redatti;
- quando il file manca o non è stato fornito, non inventa contenuto e dichiara
  che la risposta non è determinabile.

## `maestro agent`

- conserva list/read/search, permission runtime, evidence binding e stop rule;
- resta l'unica modalità autorizzata a esplorare il workspace;
- non eredita completion o contesto dal comando chat;
- non degrada silenziosamente a chat quando choreography o retrieval falliscono.

---

# Confine della milestone

## Incluso

- ADR delle modalità e dei nomi CLI;
- configurazione separata dei profili chat e agent;
- supporto provider-neutral per `num_ctx` e `thinking`, con mapping Ollama
  osservabile;
- caricamento bounded di un solo file esplicito;
- completion non-streaming iniziale e streaming soltanto dopo equivalenza
  verificata;
- telemetria redatta di latenza, token, context e thinking;
- gate deterministici, negativi e live sul computer attuale;
- candidate record read-only da consegnare alla Milestone 15.

## Escluso

- retrieval implicito, tool calling o state machine in chat;
- multi-file automatico, glob, directory o selezione autonoma di contesto;
- session memory persistente;
- mutazioni, approval, shell, Git, processi, Docker o Composer;
- nuovi modelli scaricati in serie senza shortlist e stop rule;
- support claim, tag o release v0.3.0 prima della Milestone 15.

---

# Configurazione candidata

I profili devono essere distinti e completi, per esempio:

```yaml
interaction:
  chat:
    model: qwen2.5-coder:7b
    num_ctx: 4096
    thinking: false
    timeout: 5m
  agent:
    model: qwen3.5:9b
    num_ctx: 8192
    thinking: default
    timeout: 10m
```

La sintassi è illustrativa fino all'ADR e non modifica retroattivamente lo
schema v0.2.0. `thinking` deve distinguere almeno `default`, `true` e `false`
quando il provider li supporta; un provider che non può onorare il valore
richiesto fallisce il preflight invece di ignorarlo.

`num_ctx` richiesto, context effettivo del runner e truncation devono essere
osservabili. Una differenza non dichiarata invalida il candidate record.

---

# Gate deterministici e negativi

Prima delle prove live devono essere verdi almeno:

- parsing strict della configurazione per profili distinti;
- rifiuto di path assoluto, traversal, symlink evasivo e file fuori workspace;
- rifiuto di directory, file troppo grande, encoding invalido e contenuto
  cambiato durante la lettura;
- zero tool nel request provider di chat;
- zero accesso a retrieval, index e state machine;
- nessun fallback agentico dopo timeout o risposta vuota;
- redazione di path fisici, contenuto, prompt e response dai log operativi;
- cancellazione, provider timeout e hard limit fail-closed;
- workspace byte-identico in ogni terminale.

---

# Qualificazione live sul computer attuale

Ogni candidate record congela modello, digest, template, provider, hardware,
prompt, file, domanda, `num_ctx`, `thinking`, timeout e limiti. La prima serie
usa `qwen2.5-coder:7b`; un failure non autorizza prompt tuning durante la
serie.

| Gate | Prova | Criterio |
|---|---|---|
| C0 | chat senza file | 3/3 risposte epistemicamente corrette; zero contenuto inventato |
| C1 | stesso file e stessa domanda | 3/3 `correct`; zero falsità materiale |
| C2 | streaming/non-streaming | contenuto e terminale semanticamente equivalenti 2/2 |
| C3 | operatività | timeout, latenza, token, context e thinking entro limiti congelati |
| C4 | sicurezza | workspace invariato, zero tool/retrieval/fallback e zero leak |

Continue può essere usato come confronto qualitativo separato soltanto con lo
stesso modello, file e domanda. Non sostituisce nessun gate Maestro e il suo
template deve essere dichiarato quando osservabile.

La milestone si arresta dopo il primo gate fallito. Un secondo modello può
essere valutato soltanto se la shortlist e l'ordine erano congelati prima di
C0; non viene aperta una ricerca seriale indefinita.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | ADR interaction modes e CLI | Completata | Milestone 13 |
| 2 | Profili `num_ctx`/`thinking` osservabili | Non avviata | Fase 1 |
| 3 | Implementazione `maestro chat` single-file | Non avviata | Fase 2 |
| 4 | Matrice deterministica, negativa e anti-leak | Non avviata | Fase 3 |
| 5 | Qualificazione live sul computer attuale | Non avviata | Fase 4 |
| 6 | Audit e handoff read-only | Non avviata | Fase 5 |

Le fasi sono sequenziali rispetto ai gate. Codice preparatorio per una fase
successiva può essere sviluppato nello stesso branch, ma non viene promosso né
usato come evidenza finché il gate precedente non è chiuso. Ogni fase produce
un report sotto `docs/reports/`; la Fase 6 produce anche
`docs/reports/milestone-14-final.md`.

---

# Regole trasversali

- `chat` e `agent` sono modalità diverse, non livelli di fallback della stessa
  esecuzione;
- la modalità chat non costruisce l'application graph agentico e non riceve
  registry di tool, retrieval, index, sessione o approver;
- il file è selezionato esclusivamente dall'utente e resta l'unico contenuto
  workspace disclosure ammesso nella prima iterazione;
- configurazione, modello, prompt, limiti, timeout e criteri vengono congelati
  prima della prima run live ufficiale;
- le prove live non avviano Ollama, non scaricano modelli e non modificano il
  catalogo del provider;
- `skipped`, `unsupported`, `failed` e `not_run` non vengono reinterpretati
  come PASS;
- prompt istanziati, response complete, contenuti dei file, secret e path
  fisici non entrano nei log operativi o nei report committati;
- ogni scenario registra lo stato fisico pre/post del workspace; qualunque
  differenza inattesa arresta la fase;
- un cambiamento a candidato, prompt, template, modello o profilo dopo C0
  invalida la serie e richiede un nuovo candidate record;
- la milestone resta development-only: nessuna fase modifica il support claim
  v0.2.0 o produce release, tag o artifact pubblico.

---

# Mappa delle responsabilità

| Area | Fase owner | Superficie principale |
|---|---:|---|
| semantica `chat`/`agent`/`run` | 1 | ADR, CLI e compatibility contract |
| profili e opzioni generative | 2 | product config, contratti provider e adapter |
| caricamento file e completion diretta | 3 | application service dedicato e CLI |
| sicurezza, terminali e regressione | 4 | test deterministici e harness anti-leak |
| candidato modello sul ThinkPad corrente | 5 | protocollo live e candidate record |
| verdetto e input per nuovo hardware | 6 | report finale e handoff Milestone 15 |

---

# Fase 1 — ADR interaction modes e CLI

Stato: Completata.

## Obiettivo

Congelare il confine di prodotto e l'interfaccia pubblica delle due modalità
prima di estendere configurazione o provider.

## Attività

- inventariare la semantica corrente di `maestro run`, il composition root,
  gli exit code e la configurazione v1;
- definire `maestro chat` come completion diretta e `maestro agent` come nome
  della modalità agentica verificata;
- decidere esplicitamente se `maestro run` resta alias temporaneo, viene
  deprecato o resta invariato, inclusi messaggio e finestra di compatibilità;
- congelare sintassi CLI, provenienza della domanda, cardinalità di `--file`,
  comportamento senza file e conflitti tra positional input e stdin;
- definire terminali, reason code, exit code e separazione stdout/stderr per
  success, input invalido, containment, provider failure, timeout e cancel;
- decidere l'evoluzione dello schema di configurazione e la compatibilità dei
  file v1 esistenti senza ampliare implicitamente autorità;
- definire il confine architetturale del servizio chat e le dipendenze che non
  può ricevere;
- aggiornare security model, CLI contract e ADR index con la decisione
  approvata;
- eseguire la baseline repository-wide prima del primo cambiamento al codice.

## Gate di uscita

- ADR approvata senza decisioni CLI o di compatibilità ancora implicite;
- `chat`, `agent` e l'eventuale `run` hanno semantica, autorità e migrazione
  non ambigue;
- errori, terminali e formato dell'output sono definiti prima
  dell'implementazione;
- la baseline è verde e il profilo v0.2.0 read-only non cambia;
- ogni delta di codice ha una fase owner.

## Deliverable

- ADR delle interaction modes;
- aggiornamento del contratto CLI e di sicurezza;
- `docs/reports/milestone-14-phase-1.md`.

Gate: **superato**.

---

# Fase 2 — Profili `num_ctx`/`thinking` osservabili

## Obiettivo

Introdurre profili generativi separati e opzioni provider-neutral che siano
validate, mappate e osservate senza essere ignorate silenziosamente.

## Attività

- aggiungere profili completi e distinti per chat e agent nello schema deciso
  dalla Fase 1;
- modellare `num_ctx` con limiti espliciti e `thinking` almeno come
  `default`, `true` e `false`, evitando boolean zero-value ambigui;
- estendere la request generativa provider-neutral senza collegarla ai
  contratti di sessione agentica;
- mappare esattamente le opzioni nella native chat API Ollama e coprire body
  non-streaming e streaming;
- dichiarare capability e preflight per adapter o modelli che non possono
  onorare il valore richiesto;
- propagare il profilo agent esistente senza modificarne tool, policy o
  choreography;
- rendere osservabili valore richiesto, valore effettivo quando attestabile,
  truncation, usage e motivo di un mismatch;
- verificare strict parsing, default, range, combinazioni non supportate e
  backward compatibility concordata nell'ADR.

## Gate di uscita

- nessuna opzione richiesta viene omessa o trasformata senza evidenza;
- un adapter non compatibile fallisce il preflight con reason code stabile;
- request catturate per Ollama contengono gli esatti valori configurati;
- chat e agent possono usare modelli, context, thinking e timeout distinti;
- log e diagnostica espongono metadati, non prompt o contenuti.

## Deliverable

- schema e configurazione example development-only;
- contratti provider-neutral e mapping Ollama testati;
- documentazione di configurazione e osservabilità;
- `docs/reports/milestone-14-phase-2.md`.

---

# Fase 3 — Implementazione `maestro chat` single-file

## Obiettivo

Consegnare una completion diretta bounded, inizialmente non-streaming, che
riceve soltanto domanda e file esplicitamente selezionato.

## Attività

- implementare un servizio chat dedicato che dipende dalla sola capability di
  completion provider e non dal runtime agentico;
- aggiungere il comando `maestro chat` secondo il contratto della Fase 1;
- risolvere esclusivamente path relativi sotto il workspace configurato e
  applicare containment e symlink policy prima dell'apertura;
- accettare soltanto un file regolare, entro il limite byte, con encoding
  ammesso e identità stabile durante la lettura;
- costruire messaggi system/user distinti con delimitazione non ambigua tra
  istruzione, path logico e contenuto non fidato;
- inviare una singola request senza `Tools`, senza tool choice agentica e senza
  output che possa attivare effetti;
- gestire assenza del file, file vuoto, risposta vuota, timeout,
  cancellazione e hard limit senza fallback;
- emettere risultato e metadati redatti con output bounded;
- lasciare streaming disabilitato finché l'equivalenza non viene coperta nella
  Fase 4.

## Gate di uscita

- una richiesta valida esegue esattamente una completion e zero tool call;
- retrieval, index, state machine, sessione e approver non vengono costruiti o
  invocati;
- ogni path non ammesso fallisce prima della disclosure al provider;
- nessun terminale di errore avvia un secondo percorso di esecuzione;
- workspace e file restano byte-identici.

## Deliverable

- comando e servizio chat development-only;
- loader single-file bounded;
- test unitari del percorso positivo e dei terminali principali;
- `docs/reports/milestone-14-phase-3.md`.

---

# Fase 4 — Matrice deterministica, negativa e anti-leak

## Obiettivo

Dimostrare in modo ripetibile che il nuovo percorso conserva i confini di
autorità, fallisce chiuso e non regredisce il verified agent.

## Attività

- coprire path assoluto, traversal, symlink interno ed evasivo, directory,
  FIFO/device, file assente e workspace root sostituita;
- coprire file troppo grande, encoding invalido, file modificato o sostituito
  durante la lettura e output provider oltre limite;
- catturare ogni request e attestare zero tool, zero retrieval, zero index,
  zero sessione e zero fallback;
- coprire config sconosciuta/duplicata, valori `num_ctx` e `thinking` invalidi
  e capability provider non supportata;
- coprire provider timeout, deadline, cancel, SIGINT/SIGTERM, risposta vuota,
  risposta malformata e shutdown bounded;
- aggiungere streaming soltanto dietro flag esplicito e provare con provider
  controllato equivalenza di contenuto, usage, terminale e reason code;
- scandire stdout, stderr, log ed evidenze contro prompt, response completa,
  contenuto fixture, secret e path fisici;
- rieseguire test completi, race detector, vet e i test CLI del percorso
  agentico esistente.

## Gate di uscita

- matrice positiva e negativa completamente verde;
- fail-closed e assenza di fallback dimostrati per ogni terminale;
- non-streaming e streaming sono deterministicamente equivalenti prima delle
  prove live C2;
- workspace byte-identico in ogni scenario;
- suite, race detector, vet e diff check verdi, scansione anti-leak negativa.

## Deliverable

- matrice deterministica e negativa versionata;
- harness di equivalenza e anti-leak;
- `docs/reports/milestone-14-phase-4.md`.

---

# Fase 5 — Qualificazione live sul computer attuale

## Obiettivo

Stabilire se l'esatto candidato `qwen2.5-coder:7b` e il profilo direct/chat
sono utilizzabili sul computer attuale senza confondere validità della
superficie e qualità del modello.

## Attività

- congelare commit, binario, provider, modello, digest, template, hardware,
  file, domanda, ground truth, limiti, timeout, `num_ctx` e `thinking`;
- eseguire preflight senza avviare provider o acquisire modelli;
- eseguire C0–C4 nell'ordine, con ripetizioni e stop rule già definite;
- verificare separatamente correttezza epistemica senza file e correttezza
  fattuale con file;
- confrontare streaming e non-streaming sull'esatto input dopo il PASS C1;
- registrare latenza, usage, context, thinking, terminale, reason code,
  immutabilità e leak scan senza payload proibiti;
- classificare ogni failure come contratto, modello, provider, ambiente o
  harness soltanto in presenza di evidenza positiva;
- usare Continue, se utile, esclusivamente come confronto qualitativo esterno
  e non come sostituto di un gate Maestro.

## Gate di uscita

- C0–C4 hanno stato esplicito e tutte le run residue dopo uno stop sono
  `not_run`;
- candidate, profilo ed evidenze sono ripetibili e non sono stati ritoccati
  durante la serie;
- workspace invariato e scansione anti-leak negativa;
- l'esito distingue validità del contratto e qualificazione del modello.

## Deliverable

- candidate record redatto;
- report delle prove C0–C4;
- `docs/reports/milestone-14-phase-5.md`.

---

# Fase 6 — Audit e handoff read-only

## Obiettivo

Chiudere la milestone con un verdetto unico, documentazione coerente e input
riproducibili per la qualificazione sul nuovo hardware della Milestone 15.

## Attività

- auditare implementazione, ADR, configurazione, CLI, security model,
  compatibility e known issues contro il confine approvato;
- verificare che `maestro agent` conservi i gate read-only e che nessun claim
  mutativo o di release sia stato introdotto;
- consolidare esiti deterministici e live senza trasformare failure o run non
  eseguite;
- assegnare uno degli esiti ammessi e motivare eventuali limitazioni;
- congelare candidate record o rinvio, inclusi input necessari a ripetere le
  prove sul nuovo hardware;
- aggiornare roadmap e contesto di progetto;
- eseguire suite finale, race detector, vet, diff check e scansione anti-leak.

## Gate di uscita

- documentazione e comportamento implementato descrivono lo stesso contratto;
- zero blocker non classificati e zero claim oltre l'evidenza raccolta;
- handoff Milestone 15 completo e ripetibile;
- la milestone non ha prodotto release né modificato il support claim v0.2.0.

## Deliverable

- `docs/reports/milestone-14-phase-6.md`;
- `docs/reports/milestone-14-final.md`;
- candidate record o rinvio motivato per la Milestone 15.

---

# Esiti ammessi

| Esito | Significato |
|---|---|
| `direct_chat_candidate` | modalità e profilo superano C0–C4 e possono essere ripetuti sulla nuova piattaforma |
| `direct_chat_model_rejected` | superficie valida, modello/profilo non supera i gate live |
| `direct_chat_contract_rejected` | il contratto non conserva confini, osservabilità o comportamento epistemico |
| `direct_chat_deferred` | nessun candidato utilizzabile entro shortlist e hardware correnti |

Nessun esito produce una release o riapre la matrice della Milestone 13.

---

# Deliverable

- ADR delle modalità `chat` e `agent`;
- comando e configurazione candidate development-only;
- matrice deterministica e negativa;
- report C0–C4 redatto;
- `docs/reports/milestone-14-final.md`;
- candidate record o rinvio motivato per la Milestone 15.
