# Milestone 8 — v0.1.0 Productization Development Plan

Versione: 0.4.0

Stato: In esecuzione — Fasi 1–4 completate, Fasi 5–6 da avviare

Data: 2026-08-13

Documenti di riferimento:

- `release-readiness-audit.md`;
- `milestone-8-design.md`;
- `adr/ADR-0026.md`;
- `reports/milestone-7-final.md`.

---

# Obiettivo operativo

Consegnare v0.1.0 come primo percorso installabile e controllato per eseguire
il reference agent di Maestro su un progetto locale reale.

Il gate ha dato GO alla Milestone 8, non alla release. La milestone termina
soltanto quando un nuovo utilizzatore può installare un artifact, configurarlo,
diagnosticarlo ed eseguire il percorso Laravel live senza usare il checkout di
sviluppo o conoscere le API Go interne.

---

# Sequenza

| Fase | Titolo | Stato | Dipende da |
|---|---|---|---|
| 1 | Release contract e audit | Completata | Milestone 7 |
| 2 | Configurazione e CLI minima | Completata | Fase 1 |
| 3 | Esperienza operativa | Completata | Fase 2 |
| 4 | Packaging e installazione | Completata | Fasi 2–3 |
| 5 | Validazione live e release candidate | Da avviare | Fasi 2–4 |
| 6 | Documentazione pubblica e v0.1.0 | Da avviare | Fasi 1–5 |

Le fasi sono sequenziali rispetto al gate, ma documentazione e test vengono
aggiornati in ogni incremento. Ogni fase produce un report sotto
`docs/reports/`; la Fase 6 produce anche `reports/milestone-8-final.md`.

---

# Regole di esecuzione

- Nessun comando può aggirare Provider, Context, Tool o Agent Runtime.
- Provider, modello, workspace, agente, policy, tool set e limiti restano
  espliciti.
- L'assenza di configurazione o capability richiesta è un errore osservabile,
  non un fallback implicito.
- I test live sono separati dai gate deterministici e richiedono configurazione
  esplicita; la loro assenza non produce PASS.
- Nessuna fase anticipa sandbox, multi-agent, recovery, shell o packaging di
  estensioni terze.
- Una fase non è completata dalla sola presenza del codice: servono test,
  documentazione e verifica dei criteri di uscita.

---

# Fase 1 — Release contract e audit

Stato: Completata.

## Obiettivo

Definire formalmente ciò che v0.1.0 promette e ciò che esclude, trasformando il
risultato della Milestone 7 in criteri di prodotto verificabili.

## Decisioni registrate

- piattaforma ufficiale iniziale: Linux `amd64`;
- provider ufficiale già validato: Ollama;
- fixture positiva: `llama3.1:8b` per chat/tool calling e
  `embeddinggemma:latest` per embedding;
- caso negativo canonico: `qwen2.5-coder:7b` per tool calling;
- llama.cpp candidato e condizionato alla presenza della matrice live;
- CLI e configurazione pubbliche ma sperimentali nella serie 0.x;
- runtime, tool, agenti e plugin built-in trusted in-process;
- configuration contract YAML strict `version: 1`;
- esclusioni e criteri di accettazione della release.

## Evidenza llama.cpp

Il repository e la cronologia Git disponibili non contengono il report
`reports/milestone-3-live-llamacpp-validation.md`. Le indicazioni esterne di
completamento non possono essere certificate da questa baseline. La Fase 5
deve recuperare e verificare il report oppure rieseguire la matrice; in assenza
di entrambe, una decisione esplicita deve classificare llama.cpp sperimentale.

## Deliverable

- `docs/release-readiness-audit.md`;
- `docs/milestone-8-design.md`;
- `docs/adr/ADR-0026.md`;
- riallineamento iniziale di roadmap, README e `MAESTRO_CONTEXT.md`.

## Gate di uscita

- GO alla Milestone 8 esplicito;
- support matrix iniziale e limiti di sicurezza definiti;
- contratti sperimentali e ambiti esclusi dichiarati;
- release gate e prova pulita descritti;
- nessun requisito di sandbox, ecosistema o SDK stabile nascosto nel piano.

Gate: **superato**. Report: `reports/milestone-8-phase-1.md`.

---

# Fase 2 — Configurazione e CLI minima

Stato: Completata.

## Obiettivo

Implementare esclusivamente il percorso necessario per configurare,
diagnosticare, ispezionare ed eseguire Maestro dalla CLI.

## Configurazione

- introdurre tipi prodotto separati da `runtime.Config`;
- implementare un singolo documento YAML con campo obbligatorio `version: 1`;
- usare parsing strict: campi sconosciuti o duplicati, documenti multipli,
  alias ciclici e trailing data falliscono;
- risolvere il file tramite `--config`, `MAESTRO_CONFIG`, poi percorso XDG,
  senza merge implicito;
- modellare provider Ollama/llama.cpp, modelli, workspace, agent, tool, policy,
  limiti e context budget;
- referenziare secret soltanto tramite nome di variabile d'ambiente;
- applicare validazione field-level e cross-field;
- produrre errori stabili, redatti e associati al campo.

## CLI minima

- `maestro doctor`;
- `maestro models`;
- `maestro agents`;
- `maestro run`;
- `maestro version`;
- root help, `help`, `--help` e help di ogni comando coerenti;
- stdout, stderr e codici di uscita separati e documentati;
- namespace `bench` esistente conservato.

## Composition applicativa

- costruire composition root, provider, plugin Laravel e policy dalla config;
- registrare soltanto il provider selezionato;
- ottenere il Workspace dal plugin Laravel quando richiesto;
- costruire query, budget e `RunRequest` con target immutabili;
- applicare timeout e shutdown bounded anche su failure parziale;
- mantenere parsing/rendering separati dai servizi per consentire test senza
  terminale o provider reale.

## Semantica dei comandi

### `doctor`

Valida schema, workspace, provider, modelli, capability, plugin, agent, tool,
policy e limiti. I probe provider sono read-only; non invoca il modello, non
carica/rimuove modelli e non modifica il workspace. Ogni check produce
`pass`, `warn`, `fail` o `skip` senza nascondere i check indipendenti.

### `models`

Interroga soltanto il provider esplicito, elenca modelli e capability
osservabili e non sceglie, carica, scarica o rimuove modelli.

### `agents`

Elenca descriptor e capability tramite Agent Runtime/Gestor. Deve includere
`agent.reference` e non eseguire I/O provider.

### `run`

Riceve l'istruzione come argomento o stdin, carica il workspace, indicizza il
contesto, registra la policy e invoca `Agent Runtime.Run`. La Fase 2 consegna il
percorso funzionale; rendering interattivo e UX completa vengono consolidati
nella Fase 3.

### `version`

Stampa versione semantica e commit. I release build devono restituire
esattamente la versione del tag e non un valore vuoto o `(devel)`.

## Exit code iniziali

| Codice | Significato |
|---:|---|
| 0 | Operazione completata |
| 1 | Failure operativa o run non completed |
| 2 | Uso CLI o configurazione non valida |
| 3 | Permission negata o approval non disponibile |
| 4 | Provider, modello o capability non disponibile |
| 130 | Cancellazione tramite interrupt |

## Invarianti

- help e version non richiedono configurazione;
- doctor non produce model call o mutation;
- models e agents non selezionano target impliciti;
- un file esplicitamente richiesto e assente fallisce;
- nessun secret appare nel valore serializzabile, negli errori o nell'output;
- la CLI usa `Agent Runtime.Run` e non invoca direttamente istanze tool;
- nessuna policy permissiva è registrata per default.

## Test

- configurazioni minime Ollama e llama.cpp;
- versione config assente/non supportata, typo, duplicati e trailing document;
- path, URL, ID, durate, cardinalità e budget ai limiti;
- env secret assente/presente senza leakage;
- precedenza flag/ambiente/file;
- help, comando sconosciuto e flag invalido;
- doctor con failure multiple e assenza di effetti;
- models ordinato, endpoint failure e capability unavailable;
- agents senza I/O e con reference agent;
- run deterministico read-only e mutante tramite provider scripted;
- version development/release ed exit code 0/1/2/3/4/130.

## Deliverable

- loader e tipi di configurazione;
- application builder;
- cinque comandi minimi;
- metadata di build;
- test del contratto CLI/config;
- report `reports/milestone-8-phase-2.md`.

## Gate di uscita

Un utilizzatore può descrivere un run senza leggere codice Go, validare la
configurazione, ispezionare modelli/agenti ed eseguire il reference agent dal
binario di sviluppo con target e codici di uscita deterministici.

Gate: **superato**. Report: `reports/milestone-8-phase-2.md`.

---

# Fase 3 — Esperienza operativa

Stato: Completata.

## Obiettivo

Rendere comprensibili e controllabili dal terminale le capacità esistenti,
senza ampliare l'autorità del runtime.

## Sviluppo

- implementare Approver terminale cancellabile;
- mostrare una sintesi delle action concrete prima dell'approvazione;
- offrire deny, allow one-shot e allow per il run quando valido;
- negare su EOF, timeout, input invalido, assenza di TTY o approver;
- visualizzare piano, stato step, contatori e limiti in forma sintetica;
- distinguere progresso, risultato finale, terminale ed errori;
- gestire SIGINT e shutdown bounded;
- redigere output e diagnostiche per evitare esposizione implicita di prompt,
  contenuti workspace, arguments, output tool, API key o root assoluta;
- mantenere un output non interattivo deterministico per automazione;
- evitare un flag globale `--yes` nella baseline.

## Invarianti

- approval non fabbrica permit e non aggira Tool Runtime;
- prompt senza input attendibile equivale a deny;
- allow one-shot non sopravvive alla action;
- allow run non sopravvive al run o al processo;
- provider, modello, workspace, agente, policy e limiti non cambiano durante il
  run;
- il modello non può modificare piano visualizzato senza una revisione valida;
- cancellazione non promette rollback di effetti già iniziati;
- l'output umano può mostrare localmente ciò che serve ad approvare, ma eventi
  e report mantengono le allowlist redatte.

## Test

- approval one-shot e run-scoped;
- deny esplicito, EOF, no TTY, timeout e input invalido;
- action cambiata dopo approval e replay del grant;
- piano con transizioni, terminale completed e terminali di failure;
- SIGINT prima/durante model call e tool call;
- hard limit di durata, turni, tool, token e byte;
- stdout/stderr separati e assenza di dati sensibili;
- observer in errore/panic senza corruzione dello stato.

## Deliverable

- Approver terminale;
- renderer di piano/stato/terminale;
- gestione cancellazione e non-interactive mode;
- documentazione dell'UX operativa;
- report `reports/milestone-8-phase-3.md`.

## Gate di uscita

Un utente comprende cosa Maestro intende fare, può negare o approvare l'effetto,
può cancellare il run e riceve un terminale chiaro senza leakage implicito.

Gate: **superato**. Report: `reports/milestone-8-phase-3.md`.

---

# Fase 4 — Packaging e installazione

Stato: Completata.

## Obiettivo

Produrre un packaging candidate installabile fuori dal repository e ripetibile
dallo stesso commit. La validazione live che lo promuove a release candidate
appartiene alla Fase 5.

## Sviluppo

- definire build riproducibile del binario Linux `amd64`;
- incorporare versione e commit;
- produrre archivio e checksum SHA-256;
- adottare Apache-2.0 secondo la decisione esplicita del maintainer e ADR-0027;
- includere licenza e documenti richiesti nell'archive/source release;
- creare procedura di installazione, verifica, upgrade e uninstall;
- aggiungere `configs/maestro.example.yaml` coerente con lo schema;
- includere o materializzare una fixture Laravel versionata e priva di secret;
- creare uno script/gate che non dipenda da file non tracciati;
- verificare contenuto artifact e assenza di credenziali/path utente;
- classificare sperimentali eventuali artifact non Linux `amd64`.

## Invarianti

- versione di packaging, `maestro version`, nome artifact e manifest
  coincidono e il commit è esatto;
- il binario non dipende dal checkout o dalla directory corrente;
- la configurazione di esempio non contiene secret o default permissivi;
- l'artifact non scarica modelli/plugin durante installazione;
- la fixture non contiene dipendenze installate o dati utente;
- la build non dichiara supportata una piattaforma non provata.

## Verifiche

- build ripetuta dallo stesso commit e confronto dove applicabile;
- verifica checksum;
- ispezione contenuto archive;
- esecuzione `version`, help e doctor locale dall'artifact estratto;
- scansione per secret e path locali;
- installazione in directory pulita senza repository.

## Deliverable

- pipeline/script di release;
- packaging candidate Linux `amd64` e checksum;
- configurazione di esempio;
- fixture Laravel di release;
- guida d'installazione preliminare;
- report `reports/milestone-8-phase-4.md`.

## Gate di uscita

Esiste un packaging candidate identificabile che un tester può installare e
avviare senza accesso al checkout di sviluppo. Non è ancora presentato come
release candidate o candidato alla pubblicazione.

Gate: **superato**. Report: `reports/milestone-8-phase-4.md`.

---

# Fase 5 — Validazione live e release candidate

Stato: In corso — validazione live sospesa dopo incidente OOM; NO-GO RC.

## Obiettivo

Provare l'intero percorso CLI con modelli reali e verificare la release
candidate come farebbe un nuovo utilizzatore.

## Matrice Ollama

- usare `llama3.1:8b` come fixture positiva chat/tool calling;
- usare `embeddinggemma:latest` per embedding quando richiesto;
- eseguire doctor, models, agents e run dall'artifact;
- completare uno scenario Laravel read-only;
- completare `read -> patch -> reindex -> final` con approval one-shot;
- verificare cancellazione, deny, hard limits e shutdown;
- conservare il caso `qwen2.5-coder:7b` come failure negativa canonica senza
  promuoverlo a fixture del reference agent.

## Matrice llama.cpp

- cercare/acquisire il report live indicato dalle decisioni successive;
- verificarne commit, versioni, configurazione, artifact e risultati;
- se il report non è recuperabile, rieseguire integration e Smoke matrix;
- coprire listing, discovery disponibile, completion, streaming,
  cancellazione, embedding, structured output e tool calling;
- usare idealmente lo stesso modello base Llama 3.1 per isolare il runtime;
- classificare skip e capability version-sensitive;
- produrre `reports/milestone-3-live-llamacpp-validation.md`;
- chiudere formalmente la Milestone 3 oppure adottare prima della RC una
  decisione che renda llama.cpp sperimentale.

## Installazione pulita

- predisporre un ambiente Linux `amd64` privo del checkout;
- installare e verificare checksum/versione;
- creare la configurazione dalla guida pubblica;
- eseguire l'intero quick start Laravel;
- verificare file modificato, digest, nuova generazione e terminale completed;
- ripetere approval deny, no TTY e SIGINT;
- registrare hardware, versioni e risultati in forma redatta.

## Gate release candidate

- zero failure nello scenario Ollama supportato;
- capability necessarie del reference agent validate live;
- llama.cpp allineato con report verificato oppure esplicitamente sperimentale;
- nessuna mutazione senza permission allow valida;
- cancellazione e limiti osservati;
- nessun secret, prompt o contenuto workspace negli eventi/report;
- CLI completa eseguita dall'artifact installato;
- nessun blocker di release aperto.

## Deliverable

- report live Ollama agentico;
- report live llama.cpp verificato o nuovo;
- report d'installazione pulita;
- release candidate e checksum aggiornati;
- report `reports/milestone-8-phase-5.md`.

## Gate di uscita

La release candidate soddisfa la definizione di prodotto su un modello reale e
da un ambiente nuovo; la matrice di supporto è basata su evidenze presenti.

## Gate intermedio 2026-08-13

Il preflight del candidate e la Smoke matrix Ollama provider-level sono verdi,
ma lo scenario mutativo agentico non è ancora concluso. Due tentativi della
matrice llama.cpp in router mode hanno causato OOM sull'host da 15 GiB e
destabilizzato VS Code; le relative prove sono invalidate. Nessun RC è stato
prodotto. Evidenze, contenimento e strategia di ripresa sono in
`reports/milestone-8-phase-5-interim.md`.

Gli hardening successivi a `pc.1` sono inclusi nel packaging candidate
`v0.1.0-pc.2`, prodotto e verificato integralmente dal commit
`b9f571ac5914d2565e2a7bd28f4d5d6fc14a2710`. Il quick start read-only esatto
dall'archive è positivo, ma il primo run mutativo ha confermato la necessità di
una coreografia deterministica tra lettura e patch. `pc.2` non è quindi
promuovibile. La ripresa dei gate live richiede un nuovo candidate prodotto
dopo l'hardening descritto in ADR-0028; non è ancora una release candidate.

---

# Fase 6 — Documentazione pubblica e v0.1.0

Stato: Da avviare.

## Obiettivo

Pubblicare un prodotto la cui documentazione, matrice di supporto, artifact e
comportamento osservato coincidono.

## Sviluppo

- riscrivere il README intorno a installazione e primo utilizzo;
- pubblicare quick start riproducibile;
- pubblicare guida completa alla configurazione `version: 1`;
- documentare reference agent e scenario Laravel;
- consolidare security model e canale di segnalazione vulnerabilità;
- pubblicare compatibility matrix per piattaforma/provider/modello;
- aggiungere troubleshooting basato sui failure osservati;
- creare changelog e note v0.1.0;
- verificare e pubblicare definitivamente la licenza scelta nella Fase 4;
- completare compatibility audit dei package pubblici sperimentali;
- dichiarare esplicitamente gli ambiti esclusi e i known issue;
- creare tag e artifact finali v0.1.0 con checksum.

## Documenti minimi

- README orientato all'utente;
- guida installazione;
- quick start;
- reference della configurazione;
- guida reference agent Laravel;
- security model;
- compatibility matrix;
- troubleshooting;
- changelog e release notes;
- licenza;
- `docs/v0.1.0-api-compatibility.md`;
- `reports/milestone-8-final.md`.

## Invarianti

- nessun documento presenta permission come sandbox;
- nessun provider/modello/piattaforma è detto supportato senza gate live;
- `internal/` non è API;
- i cambi futuri della config richiedono versione schema;
- SDK stabile, plugin terzi e funzionalità rinviate non vengono promessi;
- tutti i comandi copiabili corrispondono all'artifact pubblicato.

## Gate di pubblicazione

- Fasi 1–5 completate;
- suite, race detector, vet, benchmark deterministico e diff check verdi;
- installazione pulita e quick start ripetuti sull'artifact finale;
- checksum e `maestro version` verificati;
- licenza e security model presenti;
- compatibility matrix coerente con i report live;
- nessuna credenziale negli artifact;
- zero blocker release;
- tag, artifact, changelog e note indicano v0.1.0.

## Gate di uscita

Un nuovo utilizzatore può installare, configurare, diagnosticare ed eseguire
Maestro seguendo soltanto la documentazione inclusa nella release.

---

# Gate repository-wide

Ogni fase implementativa esegue almeno:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

Le modifiche al loop ripetono scenario autonomo e benchmark deterministico. Le
modifiche ad adapter, composition o CLI eseguono i relativi test live prima del
gate RC. L'assenza di un servizio live produce skip esplicito durante sviluppo,
mai completamento della Fase 5.

# Definizione di completamento

Milestone 8 e v0.1.0 sono completate soltanto quando codice, artifact,
documentazione pubblica, audit, matrici live e prova pulita descrivono lo stesso
prodotto. La sola presenza dei comandi, di un tag o di un report non verificato
non chiude la milestone.
