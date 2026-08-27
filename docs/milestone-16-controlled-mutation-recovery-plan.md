# Milestone 16 — Controlled Mutation Recovery Plan

Versione: 0.2.0

Stato: Rinviata — successiva alla baseline read-only v0.3.0 della Milestone 15;
solo analisi, progettazione e prototipi development-only

Data: 2026-08-21; aggiornamento 2026-08-27

Documenti di riferimento:

- `roadmap.md`;
- `milestone-13-field-validation-plan.md`;
- `reports/milestone-13-field-validation.md`;
- `milestone-15-reference-hardware-readonly-baseline-plan.md`;
- `milestone-17-mutation-qualification-plan.md`;
- `milestone-18-productization-v0.4.0-plan.md`;
- `milestone-11-development-plan.md`;
- `mutation-qualification.md`;
- `mutation-qualification-profile.yaml`;
- `reports/milestone-11-final.md`;
- `adr/ADR-0031.md`;
- `adr/ADR-0032.md`;
- `security-model.md`.

---

# Obiettivo operativo

Analizzare e progettare un eventuale recovery di Controlled Mutation tramite
un protocollo di proposta verificabile, senza tentare di compensare i limiti
read-only osservati nella Milestone 13.

La milestone separa diagnosi e progettazione del protocollo dalla
qualificazione ufficiale. Può produrre un protocollo model-facing, fixture e
matrici deterministiche e un compilatore candidato interno, quindi un handoff
alla Milestone 17, ma non esegue gate live, non qualifica una combinazione
hardware–provider–modello e non produce archive, release candidate, release o
support claim.

ADR-0032 resta autorevole fino al verdetto finale: `workspace.patch`, il
reference agent mutante, il modello Granite e la configurazione mutativa sono
sperimentali e non supportati. L'artifact storico v0.2.0 e il baseline v0.3.0
restano read-only.

---

# Ipotesi da verificare

Il terminale `patch_tool_call_invalid` osservato nella Milestone 11 non
identifica da solo il limite causale. Può rappresentare almeno:

- JSON malformato o troncato;
- nome del tool errato;
- campo obbligatorio mancante o sconosciuto;
- path non conforme;
- digest errato;
- formato o semantica della patch invalidi;
- tool call resa come testo invece che sul canale nativo;
- sostituzione incompatibile o ambigua rispetto al file letto.

La prima ipotesi è che il modello comprenda la modifica ma non riesca a
serializzare affidabilmente l'intero contratto `workspace.patch`. La seconda è
che il limite appartenga al canale di tool calling o alla capacità generale
del modello. La milestone deve distinguere queste ipotesi con evidenze, senza
attribuire il failure per esclusione.

---

# Contratto candidato di recovery

Il percorso attuale chiede al modello di produrre direttamente gli arguments
completi di `workspace.patch`. Il candidato da valutare separa invece
l'intenzione non fidata dalla patch autorevole:

```text
read verificata
    -> edit proposal non fidata
    -> compilazione deterministica
    -> PreparedInvocation immutabile
    -> preview concreta
    -> approval exact-fingerprint
    -> apply atomico
    -> reindex
    -> final fresh
```

La forma minima da sperimentare è equivalente a:

```json
{
  "path": "app/Http/Controllers/OrderController.php",
  "operation": "replace",
  "old_text": "...",
  "new_text": "..."
}
```

La sintassi definitiva e il canale — tool call nativa oppure structured output
— non sono decisi da questo piano. Vengono scelti dopo la matrice diagnostica e
congelati in un ADR prima dell'implementazione candidata.

Il compilatore deterministico:

- accetta soltanto una proposta strict, bounded e priva di campi sconosciuti;
- limita l'operazione iniziale a `replace` su un singolo file PHP esistente
  sotto `app/`;
- richiede che il path coincida con una read autorevole, riuscita e non
  troncata della stessa run;
- usa il digest della read autorevole, non un digest inventato o corretto dal
  modello;
- richiede che `old_text` compaia una sola volta nel contenuto letto;
- rifiuta no-op, ambiguità, overflow, encoding invalido, path non ammessi e
  stato stale;
- genera contenuto risultante, arguments `workspace.patch`, action, diff e
  fingerprint esclusivamente da dati validati;
- consegna al Tool Runtime lo stesso oggetto immutabile usato da preview,
  approval ed Execute.

Non sono ammessi repair euristici, completamento di campi mancanti, fuzzy
matching, scelta automatica fra occorrenze, inferenza di path o riscrittura
“probabile” del JSON. Una normalizzazione è ammessa soltanto se specificata,
deterministica, lossless e coperta da test.

---

# Confine della milestone

## Incluso

- analisi forense riproducibile di `patch_tool_call_invalid`;
- diagnostica development-only su fixture versionate non sensibili;
- confronto fra patch completa, edit proposal, structured output e tool schema
  minimale;
- implementazione interna dell'eventuale compilatore deterministico;
- matrice positiva e negativa deterministica senza provider o effetti;
- definizione dei gate sintetici da eseguire sulla nuova piattaforma;
- definizione dei requisiti di handoff per hardware e modelli superiori;
- verdetto sul protocollo e handoff alla Milestone 17.

## Escluso

- modifiche del profilo ufficiale read-only v0.2.x;
- pubblicazione del diagnostic mode o dei raw trace negli artifact;
- conservazione di dati diagnostici da repository reali;
- correzione euristica dell'output del modello;
- approval sintetica o non interattiva nel Gate C;
- più file, create/delete/rename, `workspace.write`, shell, Git, processi,
  Docker, Composer, Artisan, PHPUnit, sandbox, recovery o multi-agent;
- acquisto di hardware prima dell'evidenza diagnostica;
- nuove run live, selezione di modelli o tuning sul ThinkPad corrente;
- qualificazione ufficiale di WSL2, GPU, provider o modello;
- Gate A/B/C diagnostici o formali, che appartengono alla Milestone 17;
- release v0.4.0 o ampliamento della compatibility matrix.

---

# Regole trasversali

- la diagnostica usa soltanto fixture sintetiche/versionate prive di secret e
  non viene eseguita su repository reali o proprietari;
- raw response, arguments, schema e contenuti necessari alla diagnosi restano
  locali, cifrati o protetti secondo l'ambiente, con directory `0700` e file
  `0600`; non entrano in Git, eventi normali, log pubblici o report redatti;
- il diagnostic harness è interno e development-only e non viene incluso
  nell'archive di prodotto;
- ogni cattura applica limiti di byte, retention esplicita e cancellazione
  dopo la derivazione dei soli codici diagnostici pubblicabili;
- il fail-fast della qualificazione ufficiale resta invariato. Il lavoro
  forense è separato, senza effetti e senza approval;
- schema, fixture, payload sintetici e criteri restano congelati all'interno
  di ogni variante; modello, provider e hardware non sono variabili di questa
  milestone;
- nessuna prova avvia provider, scarica modelli o modifica cataloghi;
- i report pubblici conservano soltanto identità di variante, contatori,
  codici, classificazione, digest e risultato; mai payload grezzi;
- ogni failure prima del commit lascia la fixture byte-identica; gli scenari
  deterministici post-commit registrano esplicitamente stato applicato e contesto
  stale, senza rollback implicito;
- ogni modifica a compiler, prompt o schema invalida il protocollo candidato e
  richiede nuovo freeze prima dell'handoff alla Milestone 17.

---

# Contratto dell'evidenza diagnostica development-only

La Milestone 16 definisce i campi che un futuro harness live dovrà acquisire.
Non avvia nuove catture provider; può classificare raw locali preesistenti
soltanto se conformi ai vincoli di retention e permesso. Il contratto comprende:

| Campo | Scopo |
|---|---|
| raw provider response bounded | distinguere canale nativo, testo e truncation |
| tool name e arguments grezzi | localizzare parsing e schema failure |
| validation stage e error code preciso | separare sintassi, schema, path, digest e semantica |
| finish reason e usage token | rilevare limite o troncamento |
| schema esatto mostrato al modello | rendere il campione riproducibile |
| modello, digest, parametri e durata | legare il risultato al profilo |
| digest fixture pre/post | provare assenza di effetti |

Questa evidenza userà un sink distinto dalle allowlist operative. La futura
attivazione deve essere esplicita, rifiutare workspace non allowlisted e
fallire chiusa se permessi, destinazione o limiti di cattura non sono conformi.
I report committabili contengono soltanto categorie come `invalid_json`,
`wrong_tool_name`, `missing_field`, `invalid_path`, `digest_mismatch`,
`ambiguous_replace`, `textual_tool_call`, `truncated` o `unresolved`.

---

# Matrice di progettazione

Ogni variante usa la stessa read sintetica, la stessa modifica attesa e
payload positivi e negativi versionati. Non vengono aperte conversazioni con
un provider e non vengono eseguiti Tool Runtime, preview, approval o apply.

| ID | Variante | Scopo | Evidenza positiva |
|---|---|---|---|
| `MR-P0` | Contratto patch attuale | rappresentare la baseline Milestone 11 | read valida e classificazione esatta della patch completa |
| `MR-P1` | Edit proposal search/replace | ridurre la complessità della rappresentazione | path, operation, old/new esatti e compilabili |
| `MR-P2` | Structured output senza tool call | isolare il canale tool calling | oggetto strict valido e semanticamente esatto |
| `MR-P3` | Tool call con schema minimale | verificare il canale nativo con schema ridotto | singola call nativa valida e semanticamente esatta |
| `MR-P4` | Patch completa compilata | verificare l'output deterministico risultante | arguments, diff e digest identici all'atteso |

`MR-P4` non chiede al modello una seconda rappresentazione completa: verifica
che il compilatore trasformi l'output valido della variante scelta nella
stessa patch autorevole richiesta dal contratto di sicurezza.

Un protocollo supera la matrice di progettazione soltanto quando tutti i
payload positivi producono la compilazione esatta, tutti i payload negativi
sono rifiutati, non esiste repair euristico e la fixture resta invariata. Il
risultato qualifica esclusivamente il contratto deterministico: non misura un
modello, non anticipa un gate live e non è un PASS della qualificazione
ufficiale.

---

# Matrice negativa minima

Il candidato deterministico deve coprire almeno:

- deny;
- EOF e no-TTY;
- JSON/schema invalido e campi sconosciuti;
- path assoluto, traversal, file fuori `app/`, file non PHP e symlink;
- digest obsoleto;
- file modificato dopo read e dopo preview;
- `old_text` assente, ambiguo o uguale a `new_text`;
- output troncato;
- cancellazione e timeout prima e dopo commit;
- fault filesystem prima del rename;
- refresh/reindex fallito dopo commit;
- replay dell'approval;
- secondo tentativo mutativo nello stesso run;
- proposta di secondo file o operazione non supportata.

La matrice estende quella della Milestone 11; non ne rimuove casi. Preview,
fingerprint, commit atomico, reindex, freshness e terminali devono continuare a
derivare dalla stessa patch concreta.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratto di recovery e analisi forense | Non avviata | Milestone 13, ADR-0032 |
| 2 | Progettazione del protocollo model-facing | Non avviata | Fase 1 |
| 3 | Compilatore deterministico e hardening | Non avviata | Fase 2 |
| 4 | Audit del recovery e handoff hardware | Non avviata | Fasi 1–3 |

Le quattro fasi sono ricerca e non contano come qualificazione. La Milestone 17
riparte da una nuova baseline candidata e possiede piattaforma, hardware,
modello e gate ufficiali. Ogni fase produce un report sotto `docs/reports/`.

---

# Fase 1 — Contratto di recovery e analisi forense

## Obiettivo

Rendere osservabile il failure originale senza indebolire la redazione del
prodotto o confondere diagnosi e qualificazione.

## Attività

- congelare fixture, schema, tassonomia e retention dell'analisi forense;
- mappare ogni punto di parsing e validazione che oggi converge in
  `patch_tool_call_invalid` a un codice diagnostico preciso;
- progettare il sink development-only, opt-in, bounded e non incluso nel
  packaging, senza attivarlo contro un provider in questa milestone;
- riprodurre parsing e validazione della sequenza `read -> result -> patch`
  con payload sintetici o evidenze locali già disponibili, senza Tool Runtime,
  approval o effetti;
- derivare un report redatto per categoria dalle sole evidenze ammesse e
  applicare la retention congelata agli eventuali raw preesistenti;
- verificare che eventi, log e renderer pubblici restino byte-invariati e non
  espongano i nuovi dati.

## Gate di uscita

- contratto diagnostico limitato al percorso di sviluppo e a fixture
  allowlisted;
- ogni failure già osservabile riceve una categoria precisa oppure
  `unresolved` motivato;
- fixture byte-identica, zero approval e zero effetti in tutti i campioni;
- raw trace assenti da Git, artifact, log ed eventi normali;
- baseline repository-wide e anti-leak verdi.

## Deliverable

- design del diagnostic harness interno;
- tassonomia degli errori;
- `docs/reports/milestone-16-phase-1.md`.

---

# Fase 2 — Progettazione del protocollo model-facing

## Obiettivo

Stabilire quale rappresentazione minima possa essere compilata da Maestro in
una modifica semplice e corretta senza trasferire authority al modello.

## Attività

- eseguire `MR-P0`–`MR-P4` su payload versionati positivi e negativi;
- modellare separatamente tool calling nativo e structured output senza
  invocare un provider;
- verificare validità sintattica, validità schema, esattezza semantica e
  compilabilità;
- confrontare le varianti con criteri congelati;
- rifiutare varianti che richiedono repair euristico o inferenza di dati
  assenti;
- scegliere al massimo uno schema di protocollo deterministico da implementare
  e definire i gate sintetici che un futuro modello dovrà superare;
- congelare la decisione model-facing e il confine del compilatore in un ADR
  prima di modificare il percorso mutativo candidato.

## Gate di uscita

- tutte le varianti hanno fixture positive e negative eseguite o uno stato
  `skipped` motivato;
- causa del failure baseline e ruolo del canale tool sono distinti;
- un solo schema model-facing è congelato con il suo esito di progettazione; se
  nessuno schema conserva il confine di sicurezza, la recovery si arresta;
- nessun risultato viene presentato come supporto operativo.

## Deliverable

- report comparativo redatto;
- ADR del protocollo candidato oppure decisione di recovery rinviata da
  consegnare alla Fase 4;
- `docs/reports/milestone-16-phase-2.md`.

---

# Fase 3 — Compilatore deterministico e hardening

## Obiettivo

Implementare la trasformazione proposta → patch senza trasferire al Runtime
intenzione o authority non presenti nell'input validato.

## Attività

- introdurre tipi interni strict e immutabili per la proposta;
- legare proposta, read autorevole, run, call e workspace;
- compilare digest, contenuto risultante, arguments, action e diff;
- riusare preparazione, preview, permission fingerprint, permit one-shot,
  Execute atomico e reindex della Milestone 10;
- impedire retry, fallback alla patch completa e combinazione di proposte;
- coprire la matrice negativa, fault injection e concorrenza;
- provare che profilo, prompt e artifact read-only v0.3.x restino invariati;
- eseguire suite completa, race detector, vet, audit API e anti-leak.

## Gate di uscita

- una sola proposta valida produce una sola patch attesa byte per byte;
- ogni input incompleto, ambiguo, stale o fuori scope fallisce prima
  dell'approval;
- preview, fingerprint ed Execute consumano lo stesso oggetto concreto;
- matrice deterministica interamente verde;
- nessuna nuova capability è esposta dal prodotto read-only.

## Deliverable

- compilatore interno e test;
- aggiornamento della specifica mutativa sperimentale;
- `docs/reports/milestone-16-phase-3.md`.

---

# Fase 4 — Audit del recovery e handoff hardware

## Obiettivo

Riconciliare diagnosi, protocollo e compilatore e consegnare alla Milestone 17
un solo contratto mutativo da qualificare sul reference hardware.

## Attività

- verificare integrità, completezza, redazione e riproducibilità dei report;
- confermare che nessuna evidenza diagnostica grezza sia entrata in Git o
  negli artifact;
- rieseguire suite, race detector, vet, packaging read-only e anti-leak;
- aggiornare ADR-0032 con un nuovo ADR che conservi, sostituisca o rinvii il
  contratto model-facing senza indebolire gli invarianti della Milestone 10;
- dichiarare schema, canale model-facing, prompt, fixture, limiti e matrice
  esatti da trasferire;
- consegnare alla Milestone 17 il protocollo e il compilatore candidati, i
  criteri immutabili di Gate A/B/C mutativi e la matrice negativa, senza
  eseguire qualificazione live in questa milestone. I requisiti read-only di
  `num_ctx`, `thinking` e binding dell'evidenza restano quelli già qualificati
  e congelati dalla Milestone 15.

## Esiti ammessi

| Esito | Significato |
|---|---|
| `protocol_candidate_ready_for_qualification` | protocollo e compilatore deterministici sono pronti per i gate sulla piattaforma della Milestone 17 |
| `protocol_unchanged` | nessun protocollo semplificato è migliore senza indebolire il confine; la Milestone 17 riceve il contratto ADR-0031 |
| `mutation_deferred_protocol` | nessun contratto sicuro e qualificabile può essere congelato |

Uno skip, una singola fixture positiva o il solo compilatore deterministico
non autorizzano un esito `mutation_qualified`.

## Gate di uscita

- tutte le osservazioni hanno classificazione e destinazione;
- verdetto unico coerente con la matrice diagnostica e lo stato fisico delle
  fixture;
- compatibility v0.3.x e profilo read-only invariati;
- nessuna release prodotta o capability presentata come supportata;
- handoff completo alla Milestone 17 oppure rinvio motivato.

## Deliverable

- `docs/reports/milestone-16-final.md`;
- ADR conclusivo;
- profilo di handoff alla Milestone 17 oppure nuovo rinvio motivato.

---

# Condizioni per l'handoff

La Milestone 17 può dare GO alla productization soltanto quando sono presenti
contemporaneamente:

- protocollo model-facing affidabile e non euristico;
- compilatore deterministico e matrice negativa verde;
- protocollo di handoff congelato dalla Milestone 16;
- modello qualificato sul reference hardware;
- profilo piattaforma, hardware e topologia dichiarato;
- Gate A/B/C ufficiali completi nella Milestone 17.

L'ingresso successivo in una release richiede inoltre:

- configurazione mutativa separata ed esplicitamente opt-in;
- artifact distinto che ripete packaging, installazione, live, sicurezza e
  anti-leak.

Questo ampliamento dell'autorità è destinabile alla Milestone 18 per v0.4.0 o
a una release successiva, mai a v0.2.1.
