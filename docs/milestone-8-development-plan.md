# Milestone 8 — v0.1.0 Productization Development Plan

Versione: 0.1.0

Stato: Proposto — implementazione non avviata

Data: 2026-08-13

Documenti di riferimento:

- `release-readiness-audit.md`;
- `milestone-8-design.md`;
- `reports/milestone-7-final.md`.

---

# Obiettivo operativo

Consegnare v0.1.0 come primo percorso installabile e controllato per eseguire
il reference agent di Maestro su un progetto locale reale, chiudendo prima
della release il debito live della Milestone 3 oppure delimitandolo tramite una
decisione di scope esplicita.

---

# Regole di esecuzione

- Le fasi 1–5 sono incrementi implementativi; la Fase 6 è il gate di release.
- Ogni fase termina con test, documentazione e un report sotto `docs/reports/`.
- Nessuna fase può introdurre default impliciti per provider, modello,
  workspace, agente o policy.
- Nessun comando può aggirare Provider, Context, Tool o Agent Runtime.
- I test live sono separati dai gate deterministici e richiedono configurazione
  esplicita; la loro assenza non deve produrre un falso PASS.
- La matrice llama.cpp può procedere mentre vengono implementate le prime fasi,
  ma è un prerequisito della release candidate salvo decisione di scope.

---

# Sequenza

| Fase | Titolo | Stato | Dipende da |
|---|---|---|---|
| 0 | Gate post-Milestone 7 | Completata | Milestone 7 |
| 1 | Contratti di prodotto e configurazione | Da avviare | Fase 0 |
| 2 | Composition applicativa e diagnostica | Da avviare | Fase 1 |
| 3 | CLI di discovery e metadata | Da avviare | Fase 2 |
| 4 | Run, policy, approval e quick start Laravel | Da avviare | Fasi 2–3 |
| 5 | Packaging, documentazione pubblica e compatibility audit | Da avviare | Fasi 1–4 |
| 6 | Validazione live, release candidate e installazione pulita | Da avviare | Fasi 1–5 e matrice llama.cpp |

La Fase 0 è documentata in `release-readiness-audit.md` e non richiede un
ulteriore report.

---

# Track live parallela — chiusura Milestone 3

## Obiettivo

Eseguire su llama.cpp la matrice già completata per Ollama, idealmente con lo
stesso modello base Llama 3.1, per distinguere comportamento del modello da
differenze del runtime provider.

## Attività

- registrare versione di llama.cpp, modalità server/router, hardware e modelli;
- eseguire integration suite con listing, completion, streaming,
  cancellazione, embedding e, quando disponibile, discovery/lifecycle;
- eseguire `maestro bench smoke --provider llama.cpp` con mutation guard
  disabilitata salvo ambiente sacrificabile;
- verificare structured output, tool calling non-stream e streaming;
- classificare ogni skip come previsto, capability unavailable o gap di
  fixture;
- conservare report JSON canonico e Markdown derivato con permessi sicuri;
- rieseguire suite deterministica, race detector e vet dopo eventuali fix;
- produrre `reports/milestone-3-live-llamacpp-validation.md` e report finale
  della Milestone 3.

## Gate

- zero failure negli scenari obbligatori per le capability dichiarate e usate
  dal reference agent;
- cleanup e assenza di mutazioni non autorizzate;
- differenze version-sensitive documentate;
- nessuna regressione su Ollama;
- Milestone 3 formalmente chiusa.

Se l'ambiente live non consente di completare il gate, prima della RC deve
essere approvata una decisione che classifichi llama.cpp sperimentale nella
v0.1.0 e delimiti comandi, documentazione e supporto. Il semplice rinvio non è
un esito valido.

---

# Fase 1 — Contratti di prodotto e configurazione

## Obiettivo

Definire e implementare il confine strict tra input utente e composition root,
senza ancora eseguire un agente dalla CLI.

## Sviluppo

- introdurre tipi di configurazione prodotto separati da `runtime.Config`;
- implementare schema YAML `version: 1` e decoder strict;
- definire risoluzione `--config`, `MAESTRO_CONFIG`, XDG senza merge implicito;
- modellare provider Ollama/llama.cpp, modelli, workspace, agent, tool, policy,
  limiti e context budget;
- supportare secret soltanto tramite nome di variabile d'ambiente;
- implementare validazione field-level e cross-field;
- definire errori tipizzati/reason code e redazione;
- aggiungere una configurazione di esempio usata dai test e dal quick start;
- fissare grammatica CLI, help ed exit code in test golden o strutturali.

## Invarianti

- campi sconosciuti, duplicati e documenti multipli falliscono;
- un file esplicitamente richiesto e assente fallisce;
- nessun secret è memorizzato nel valore serializzabile o stampato;
- root, URL e target finali sono normalizzati una sola volta e poi immutabili;
- ogni limite viene validato contro i bound pubblici di Agent/Tool/Context;
- la config non abilita shell, process, network tool o plugin terzi.

## Test

- file minimo valido per Ollama e llama.cpp;
- versione assente/non supportata, typo, duplicati e trailing document;
- path relativi/assoluti, symlink e workspace inesistente;
- durate, budget e cardinalità ai limiti;
- env secret assente/presente senza leakage;
- precedenza flag/ambiente/file e nessun merge inatteso;
- error output stabile e redatto.

## Deliverable

- loader e tipi di configurazione;
- fixture `configs/maestro.example.yaml`;
- test del contratto;
- report `reports/milestone-8-phase-1.md`.

## Gate di uscita

Una configurazione valida descrive senza ambiguità l'intero input di un run;
una configurazione invalida non costruisce né avvia il runtime.

---

# Fase 2 — Composition applicativa e diagnostica

## Obiettivo

Costruire da configurazione i servizi ufficiali e fornire un preflight
read-only capace di spiegare ogni blocker operativo.

## Sviluppo

- creare un application builder per composition root, provider e plugin;
- registrare esclusivamente l'adapter provider selezionato;
- caricare il plugin Laravel quando richiesto e ottenere il Workspace pubblico;
- compilare la policy di prodotto sulle PermissionRequest concrete;
- costruire query, budget, estimator e `RunRequest` senza eseguirla;
- implementare check diagnostici indipendenti e ordinati;
- distinguere check locali da probe provider read-only;
- applicare timeout e shutdown bounded anche su failure parziale;
- rendere gli output diagnostici redatti e testabili.

## Invarianti

- build e doctor non invocano completion/embedding e non mutano cataloghi;
- discovery non seleziona automaticamente provider o modello;
- una failure non nasconde check indipendenti successivi;
- plugin e runtime vengono fermati se erano stati avviati;
- la root non entra in eventi o richieste modello;
- nessuna policy permissiva viene registrata per default.

## Test

- composition Ollama e llama.cpp con trasporti in-memory;
- provider/model mismatch e capability mancante;
- Laravel valido, non rilevato e non leggibile;
- agent/tool/policy assenti;
- timeout, cancellazione e cleanup dopo failure;
- prova negativa di assenza di model call e mutation;
- output senza API key, prompt, contenuto o root non necessaria.

## Deliverable

- application builder;
- servizio doctor e check catalog;
- documentazione diagnostica;
- report `reports/milestone-8-phase-2.md`.

## Gate di uscita

Lo stesso input che verrà usato da `run` può essere validato end-to-end senza
effetti e con una diagnosi actionable.

---

# Fase 3 — CLI di discovery e metadata

## Obiettivo

Consegnare una shell non interattiva coerente per help, diagnosi, cataloghi e
identità dell'artifact.

## Sviluppo

- sostituire il comportamento root implicito con help contrattualizzato;
- implementare `doctor`, `models`, `agents` e `version`;
- mantenere il namespace `bench` esistente e compatibile;
- introdurre metadata build per versione e commit;
- separare stdout, stderr e exit status;
- gestire segnali e timeout per i probe;
- documentare output umano ed eventuali reason code machine-readable;
- aggiungere test da binario oltre ai test della funzione `run`.

## Invarianti

- help e version non richiedono configurazione;
- agents non effettua I/O provider;
- models usa un solo provider esplicito e non muta il catalogo;
- nessun comando diagnostico chiede approval;
- gli artifact di release non riportano `(devel)` o versione vuota.

## Test

- root help, alias help e help per ogni comando;
- comando sconosciuto e flag invalido;
- config assente per comandi che la richiedono;
- models ordinato, endpoint failure e capability unavailable;
- agents contiene `agent.reference` e non esegue run;
- version development/release con commit;
- exit code 0/1/2/4 e interrupt.

## Deliverable

- quattro comandi di prodotto;
- metadata build;
- reference CLI aggiornata;
- report `reports/milestone-8-phase-3.md`.

## Gate di uscita

Un utilizzatore può installare un build di sviluppo, ottenere help, validare la
configurazione e ispezionare provider e agent senza leggere il codice.

---

# Fase 4 — Run, policy, approval e quick start Laravel

## Obiettivo

Esporre il reference agent tramite il percorso CLI ufficiale e dimostrare un
run controllato su Laravel.

## Sviluppo

- implementare `maestro run` con istruzione da argomento o stdin;
- indicizzare il workspace e costruire context/query entro budget;
- registrare policy compilata dalla configurazione;
- implementare Approver terminale cancellabile;
- mostrare action locali sufficienti alla decisione senza loggarle negli eventi;
- negare prompt in assenza di TTY/approver;
- renderizzare progresso bounded, risultato e terminale;
- gestire SIGINT, shutdown e stale refresh;
- creare quick start versionato sulla fixture Laravel;
- aggiungere scenari deterministici CLI read-only, allow, prompt/deny e patch.

## Invarianti

- la CLI chiama soltanto `Agent Runtime.Run` e non istanze tool;
- provider, modello, workspace, agente, policy e tool set sono visibili e
  immutabili per il run;
- nessuna scelta del modello aumenta i limiti;
- prompt senza input valido è deny;
- allow one-shot non sopravvive alla action e allow run non sopravvive al run;
- la mutazione usa digest precondition e il run non completa su contesto stale;
- cancellazione non promette rollback di effetti già iniziati.

## Test

- run read-only completed con provider scripted;
- patch approvata one-shot, digest valido, reindex e final;
- deny esplicito, EOF, no TTY, timeout e SIGINT;
- action cambiata dopo approval e grant replay;
- provider/tool failure con exit code corretto;
- redazione di eventi e diagnostiche non verbose;
- Laravel detection, index, read e patch dalla CLI.

## Deliverable

- comando `run`;
- policy compiler e terminal Approver;
- `docs/quick-start.md`;
- report `reports/milestone-8-phase-4.md`.

## Gate di uscita

Il percorso deterministico CLI completa `read -> patch -> reindex -> final` su
Laravel con policy e approval espliciti e senza bypass del runtime.

---

# Fase 5 — Packaging, documentazione pubblica e compatibility audit

## Obiettivo

Produrre un artifact candidato alla release accompagnato da tutti i contratti e
limiti necessari a un utilizzatore esterno.

## Sviluppo

- scegliere e aggiungere la licenza del progetto con decisione del maintainer;
- creare security model e canale di segnalazione vulnerabilità;
- documentare installazione, upgrade, uninstall e checksum;
- aggiungere note di release e changelog iniziale;
- produrre build riproducibili con versione/commit incorporati;
- generare checksum SHA-256 e inventario artifact;
- dichiarare piattaforme testate e supportate;
- consolidare gli audit milestone in un compatibility statement v0.1.0;
- dichiarare sperimentali CLI, config e package pubblici;
- aggiornare README, roadmap e contesto senza promesse fuori scope.

## Invarianti

- nessuna credenziale o path utente entra negli artifact;
- il source archive contiene licenza e documentazione referenziata;
- versione CLI, tag e note di release coincidono;
- i package `internal/` non sono dichiarati API;
- sandbox, plugin terzi e recovery non sono descritti come capability presenti.

## Test e verifiche

- build due volte dallo stesso commit e confronto degli artifact dove
  tecnicamente applicabile;
- verifica versione e commit incorporati;
- checksum e contenuto degli archivi;
- link/documentation audit;
- `go list`, API diff e compatibility audit;
- scansione per secret e path locali negli artifact.

## Deliverable

- pipeline/script di release;
- `LICENSE`, security model, quick start, installazione e release notes;
- `docs/v0.1.0-api-compatibility.md`;
- report `reports/milestone-8-phase-5.md`.

## Gate di uscita

Esiste un artifact RC identificabile e verificabile che può essere consegnato a
un tester senza accesso al checkout di sviluppo.

---

# Fase 6 — Validazione live, release candidate e installazione pulita

## Obiettivo

Dimostrare che l'artifact, non soltanto il repository, soddisfa la definizione
di prodotto della v0.1.0.

## Prerequisiti

- Fasi 1–5 completate;
- matrice llama.cpp verde e Milestone 3 chiusa, oppure decisione esplicita di
  classificazione sperimentale;
- gate Ollama precedente ancora valido o rieseguito se adapter/CLI sono cambiati
  materialmente.

## Esecuzione

- creare tag release candidate e artifact con checksum;
- predisporre ambiente pulito sulla piattaforma supportata;
- installare senza usare il repository di lavoro;
- eseguire help, version, doctor, models e agents;
- eseguire quick start Laravel read-only;
- eseguire patch Laravel con approval one-shot;
- verificare contenuto, digest, reindex e terminale completed;
- ripetere il deny non interattivo;
- raccogliere un report redatto con hardware, provider, modello e versioni;
- eseguire suite, race detector, vet, benchmark deterministico e diff check;
- verificare note di release, problemi noti e artifact finali.

## Gate release candidate

- installazione e tutti i comandi minimi funzionano dall'artifact;
- scenario live agentico completed entro hard limits;
- nessuna mutazione avviene senza decisione allow valida;
- nessun secret, prompt o contenuto workspace compare in eventi/report;
- shutdown e cancellazione sono bounded;
- documentazione riproduce i comandi osservati;
- zero blocker di release aperti.

## Deliverable

- report live end-to-end;
- report di installazione pulita;
- `reports/milestone-8-final.md`;
- artifact finali, checksum e note v0.1.0.

---

# Gate repository-wide

Ogni fase implementativa deve eseguire almeno:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

Le fasi che modificano il loop agentico ripetono inoltre lo scenario autonomo e
il benchmark deterministico. Le fasi che modificano adapter o composition
rieseguono i test di integrazione live pertinenti prima della RC.

# Definizione di completamento

Milestone 8 e v0.1.0 sono completate soltanto quando codice, artifact,
documentazione pubblica, audit, matrice live e prova pulita descrivono lo stesso
prodotto. La sola presenza dei comandi o il solo tag Git non chiudono la
milestone.

