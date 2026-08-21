# Milestone 13 — Field Validation & Adoption Plan

Versione: 0.1.0

Stato: Pianificata — avvio subordinato alla pubblicazione remota di v0.2.0

Data: 2026-08-21

Documenti di riferimento:

- `roadmap.md`;
- `reports/milestone-12-final.md`;
- `releases/v0.2.0.md`;
- `compatibility.md`;
- `security-model.md`;
- `operational-experience.md`;
- `field-validation-task-matrix.md`;
- `adr/ADR-0032.md`.

---

# Obiettivo operativo

Verificare Maestro v0.2.0 su almeno due progetti Laravel reali, esterni alle
fixture e al repository di sviluppo, raccogliendo evidenze sufficienti per
decidere il contratto di prodotto della v0.3.0.

La milestone misura installazione, utilità read-only, affidabilità,
prestazioni, comportamento del tool calling, compatibilità e invarianti di
sicurezza. Non introduce feature, non amplia la compatibility promise e non
presuppone una nuova release. Il suo output conclusivo è un report di prodotto.

La pubblicazione remota di v0.2.0 è una precondizione operativa, non una nuova
milestone tecnica. Il tag annotato e l'artifact finale sono già qualificati
localmente dalla Milestone 12; devono essere distribuiti senza ricostruire,
rinominare o spostare il tag.

---

# Confine del ciclo

## Incluso

- pubblicazione del commit finale, del tag annotato `v0.2.0`, della GitHub
  Release, delle release notes, dell'archive e del checksum già qualificati;
- installazione dell'archive riscaricato dalla GitHub Release pubblica;
- due repository Laravel reali autorizzati, uno piccolo e uno medio quando la
  coorte disponibile lo consente;
- task ripetibili su struttura del progetto, controller, service, dipendenze e
  flussi applicativi multi-file;
- raccolta di terminale, reason code, durata, turni, token e tool call;
- valutazione manuale `correct`, `partial`, `incorrect` o `unevaluable`;
- prove di containment, immutabilità, redazione e assenza di accesso a Docker;
- triage di bug, limiti del modello, ambiente, UX e richieste evolutive;
- feedback esterno redatto quando repository e canali di feedback lo
  consentono.

## Escluso

- nuove capability o modifiche del core eseguite come parte della raccolta;
- `workspace.write`, `workspace.patch` e approval mutativa nel profilo usato;
- shell, Git agentico, Docker API/socket, Composer, Artisan, PHPUnit, sandbox,
  recovery, remote execution, multi-agent e tool di terze parti;
- qualificazione di nuovi modelli o hardware mutativi;
- ampliamenti silenziosi della serie `v0.2.x`;
- definizione anticipata delle feature v0.3.0.

Il modello supportato resta `llama3.1:8b` tramite Ollama sul profilo Linux
`amd64`. Profili diversi possono essere annotati come esplorativi, ma non
entrano nei risultati primari e non modificano il support claim.

---

# Regole trasversali

- tutte le prove di prodotto usano l'archive scaricato dalla GitHub Release,
  mai un binario costruito dal checkout;
- versione, commit incorporato, URL della release e SHA-256 vengono congelati
  prima della prima run;
- configurazione, task, target logici, limiti, timeout, modello e criteri di
  valutazione vengono congelati prima della campagna e non cambiano dopo un
  risultato negativo;
- il profilo espone esattamente `workspace.list`, `workspace.read` e
  `workspace.search`, con `workspace_mutate: deny`;
- Ollama può essere già disponibile, ma Maestro non lo avvia, non scarica
  modelli e non modifica il suo catalogo;
- i repository devono essere usabili con autorizzazione esplicita e privi di
  secret o dati personali necessari alla valutazione; `vendor/`, asset
  generati e storage runtime restano fuori dallo snapshot del plugin;
- ogni repository riceve un alias, per esempio `project-a`; root fisiche,
  remote, nomi cliente e contenuti proprietari non entrano nei report;
- prompt istanziati, risposte complete e note di ground truth restano evidenza
  locale con permessi `0600` e non vengono committati;
- ogni run confronta lo stato fisico pre/post del workspace. Una differenza
  inattesa arresta immediatamente la campagna e apre un incidente di sicurezza;
- `passed`, `failed`, `skipped` e `not_run` restano distinti. Un prerequisito
  assente o un caso non eseguito non diventa PASS;
- nessun failure viene attribuito al modello, al prodotto o all'ambiente per
  esclusione: serve un'evidenza positiva o la classe resta `unresolved`;
- i report pubblici non includono prompt, response, contenuti del workspace,
  arguments tool, secret, path fisici o output grezzi;
- un'eventuale correzione esce dalla campagna, entra nel ramo di manutenzione
  `v0.2.x` e ripete i gate pertinenti prima di una possibile v0.2.1.

---

# Coorte e unità di misura

La coorte minima contiene due applicazioni Laravel non derivate dalle fixture
Maestro. La dimensione viene misurata prima della prova sul solo sorgente
versionato e project-owned, escludendo dipendenze e output generati:

| Strato target | File PHP project-owned | Requisito |
|---|---:|---|
| Piccolo | fino a 250 | almeno un controller, un service e un flusso verificabile |
| Medio | 251–2.000 | dipendenze e flusso multi-file non banali |

Le soglie servono a descrivere la coorte, non costituiscono una nuova promessa
di compatibilità. Se i repository autorizzati non coprono entrambi gli strati,
la deviazione viene dichiarata e i progetti non vengono rietichettati.

Per ogni progetto si registrano in forma redatta: versione PHP e Laravel,
conteggio dei file project-owned, byte indicizzati, struttura applicativa,
presenza di Dockerfile o file Compose, stato iniziale pulito/sporco, sistema
operativo, CPU, RAM e backend Ollama. Nessun identificatore hardware è
presentato come requisito minimo.

L'unità primaria è una `run` di un task congelato su un artifact, un progetto
e un profilo esatti. I quattro task core della matrice vengono eseguiti due
volte per progetto senza cambiare istruzione o configurazione. I task di
verifica e Docker seguono la frequenza indicata nella matrice.

---

# Metriche e criteri di valutazione

| Area | Misura primaria | Regola |
|---|---|---|
| Installazione | successo da GitHub Release | download pubblico, checksum, estrazione sicura, manifest e `maestro version` coerenti |
| Preflight | risultato di `doctor` | conteggio `pass`/`warn`/`fail`/`skip` per progetto, senza collassare i warning |
| Affidabilità | completion rate | run `completed` / run eseguite; `skipped` e `not_run` sono riportate fuori dal denominatore |
| Qualità | esito manuale | distribuzione `correct`/`partial`/`incorrect`/`unevaluable`, per task e progetto |
| Prestazioni | costo della run | durata, turni modello, input/output token e massimo; mediana e massimo per gruppo |
| Tool calling | call ed errori | tool call totali, run con call, `tool_failure`, call invalide confermate e casi `unresolved` |
| Sicurezza | invarianti | stato e digest pre/post identici, zero leak nei report, zero authority mutativa o Docker |
| Compatibilità | profilo osservato | Laravel/PHP, dimensione, struttura, OS/hardware e risultato senza generalizzazioni |

La qualità viene assegnata contro una scheda di evidenza preparata prima delle
run:

- `correct`: soddisfa tutti i punti obbligatori, è coerente con il sorgente e
  non contiene affermazioni materiali false;
- `partial`: contiene evidenza utile e corretta ma omette almeno un punto
  obbligatorio, senza falsità materiale;
- `incorrect`: contraddice il sorgente, inventa simboli o flussi, oppure non
  risponde al task;
- `unevaluable`: l'evidenza autorizzata non consente una decisione. La run
  resta nel conteggio di affidabilità e separata dalla percentuale di qualità.

Una seconda revisione è obbligatoria per gli esiti `partial`, `incorrect` e
per ogni disaccordo segnalato da un utilizzatore esterno. Non viene calcolato
un punteggio globale: distribuzioni e failure conservano progetto, task e
profilo come dimensioni separate.

Il contatore CLI è la fonte per le tool call totali. Una call è dichiarata
invalida soltanto quando l'evidenza redatta del runtime lo dimostra; un generico
`tool_failure` non viene reinterpretato e resta `unresolved` finché non è
riprodotto o diagnosticato.

---

# Record minimo per run

Ogni riga della raccolta locale contiene almeno:

| Campo | Contenuto ammesso nel report pubblico |
|---|---|
| `run_id` | alias non reversibile |
| `release`, `commit`, `sha256` | identità pubblica dell'artifact |
| `project_alias`, `task_id`, `repetition` | identificatori redatti |
| `profile_id` | OS/arch/provider/model/limiti congelati |
| `started_at`, `duration_ms`, `exit_code` | metadati operativi |
| `terminal`, `reason_code` | stato sintetico |
| `model_turns`, `tool_calls` | contatori CLI |
| `input_tokens`, `output_tokens` | usage disponibile; assente resta assente |
| `quality`, `rubric_code` | giudizio e codice, non risposta completa |
| `workspace_unchanged` | risultato del confronto pre/post |
| `leak_scan` | `passed`, `failed` o `not_run` |
| `classification` | product, model, environment, UX, evolution, security o unresolved |
| `note_code` | identificatore breve; nessun testo proprietario |

---

# Precondizione e sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| Gate 0 | Pubblicazione pubblica v0.2.0 | Non avviata | Milestone 12 |
| 1 | Protocollo, coorte e baseline | Non avviata | Fase 0 |
| 2 | Installazione e preflight sul campo | Non avviata | Fase 1 |
| 3 | Campagna di task read-only | Non avviata | Fase 2 |
| 4 | Sicurezza, operatività e Docker denial | Non avviata | Fase 3 |
| 5 | Triage, manutenzione e feedback | Non avviata | Fasi 3–4 |
| 6 | Report di prodotto e decisione v0.3.0 | Non avviata | Fase 5 |

Il Gate 0 precede formalmente la milestone; le Fasi 1–6 producono un report
sotto `docs/reports/`. L'esecuzione può essere interrotta per un incidente di
sicurezza, ma le evidenze già raccolte restano classificate e non vengono
presentate come una campagna completa.

---

# Gate 0 — Pubblicazione pubblica v0.2.0

## Obiettivo

Completare la distribuzione dell'esatto artifact qualificato dalla Milestone
12 e provarne l'installazione dal canale pubblico.

## Attività

- verificare che il worktree sia pulito, che i commit finali siano sul branch
  remoto atteso e che `v0.2.0^{commit}` risolva al commit incorporato
  `5b05237362370fa79f133e159105a6a99050e81a`;
- eseguire il push degli eventuali commit finali non ancora pubblicati;
- eseguire il push di `refs/tags/v0.2.0` senza ricreare o spostare il tag;
- creare la GitHub Release sul tag annotato usando
  `docs/releases/v0.2.0.md` come release notes;
- caricare `maestro-v0.2.0-linux-amd64.tar.gz` e il relativo `.sha256` già
  qualificati, senza ricostruirli;
- verificare dalla pagina pubblica tag, note e presenza dei due asset;
- in una nuova directory temporanea, scaricare gli asset dalla release
  pubblica, verificare SHA-256 ed estrazione, installare il binario ed eseguire
  `version`, help e `doctor` con il profilo supportato;
- registrare URL pubblico, timestamp, digest scaricato e risultato del
  preflight senza pubblicare endpoint o path locali.

## Gate di uscita

- branch finale e tag sono raggiungibili dal remote;
- la release non è draft o prerelease e punta al tag annotato corretto;
- il checksum pubblico verifica l'archive con SHA-256
  `c2d2a6f35178e91ad0c62d3c27f4ff2c33eedb46fd5fb327535890638e963758`;
- installazione e `maestro version` funzionano dall'asset riscaricato;
- `doctor` non dipende dal checkout e non espone dati sensibili.

## Deliverable

- GitHub Release pubblica v0.2.0;
- `docs/reports/milestone-13-release-publication.md`.

---

# Fase 1 — Protocollo, coorte e baseline

## Obiettivo

Congelare campione, task, rubriche e modalità di raccolta prima di osservare il
comportamento del prodotto.

## Attività

- selezionare almeno due repository autorizzati e assegnare alias redatti;
- inventariare framework, dimensione, struttura, sensibilità e presenza di
  configurazione Docker senza pubblicare identità o contenuti;
- scegliere controller, service, dipendenze e flussi verificabili usati per
  istanziare `field-validation-task-matrix.md`;
- preparare per ogni task una scheda di ground truth con evidenze obbligatorie
  e riferimenti locali;
- congelare artifact, configurazione read-only, hardware, provider, modello,
  limiti, task, ripetizioni, stop rule e criteri di classificazione;
- definire snapshot pre/post, conservazione locale, redazione e anti-leak;
- svolgere una dry run soltanto sulla fixture pubblica per verificare il
  raccoglitore, senza contarla nei risultati field.

## Gate di uscita

- coorte minima autorizzata e descritta in forma redatta;
- task e rubriche congelati prima della prima run reale;
- record minimo, denominatori e trattamento dei dati mancanti verificati;
- nessuna evidenza proprietaria destinata al repository.

## Deliverable

- matrice istanziata conservata localmente;
- `docs/reports/milestone-13-phase-1.md`.

---

# Fase 2 — Installazione e preflight sul campo

## Obiettivo

Verificare che l'esperienza pubblica di installazione e diagnostica funzioni
per ogni progetto della coorte.

## Attività

- ripetere download e checksum dell'asset pubblico nell'ambiente di prova;
- estrarre e installare senza usare file dal checkout Maestro;
- derivare una configurazione per progetto a partire dall'esempio incluso,
  cambiando soltanto la root autorizzata;
- confermare tool set read-only e `workspace_mutate: deny` prima dell'uso;
- eseguire `version`, help, `doctor`, `models` e `agents`;
- registrare ogni check e classificare warning, failure e prerequisiti assenti;
- verificare che il preflight non modifichi repository, catalogo Ollama o
  processi Docker.

## Gate di uscita

- installazione pubblica riuscita per ciascun ambiente ammesso;
- identità di binario, tag, manifest e checksum coerente;
- `doctor` completato per ogni progetto oppure failure riproducibile e
  classificato;
- workspace e servizi esterni invariati.

## Deliverable

- `docs/reports/milestone-13-phase-2.md`.

---

# Fase 3 — Campagna di task read-only

## Obiettivo

Misurare utilità, ripetibilità e costo operativo del reference agent su
domande reali e multi-file.

## Attività

- eseguire i task core e le ripetizioni definite nella matrice senza cambiare
  prompt, target, configurazione o limiti;
- conservare localmente stdout e stderr separati con permessi `0600` fino alla
  valutazione e poi applicare la retention definita;
- acquisire terminale, reason code, durata, turni, token e tool call;
- confrontare ogni risposta completata con la rubrica congelata;
- sottoporre `partial`, `incorrect` e casi contestati a seconda revisione;
- confrontare ripetizioni dello stesso task e distinguere completion rate da
  qualità semantica;
- arrestare il solo profilo interessato al primo failure di immutabilità o
  leak, conservando lo stato per l'analisi.

## Gate di uscita

- tutti i task core eseguiti due volte su entrambi i progetti; un caso
  `skipped` o `not_run` resta visibile ma non supera il gate;
- zero run prive di terminale o classificazione;
- qualità valutata per ogni completion, salvo `unevaluable` motivati;
- metriche aggregate riconciliabili con i record delle singole run.

## Deliverable

- `docs/reports/milestone-13-phase-3.md`;
- dataset redatto delle run, privo di contenuti dei progetti.

---

# Fase 4 — Sicurezza, operatività e Docker denial

## Obiettivo

Dimostrare che l'uso su repository realistici non amplia l'autorità del
profilo e conserva i terminali operativi della release.

## Attività

- verificare snapshot e stato del workspace prima e dopo ogni run della
  campagna, includendo baseline inizialmente sporche senza ripulirle;
- eseguire almeno una prova SIGINT, una deadline e un hard limit per profilo,
  senza trasformare questi terminali attesi in failure di affidabilità;
- su almeno un progetto che contiene Dockerfile o file Compose, eseguire il
  task Docker della matrice con soli tool read-only;
- non montare, configurare o esporre il Docker socket a Maestro e non
  registrare tool shell/process/container;
- confrontare prima e dopo lo stato dei container tramite osservazione
  dell'operatore, distinta dall'autorità concessa a Maestro;
- verificare che nessun output o report contenga secret, path fisici, remote,
  prompt, risposte complete, arguments o sorgenti;
- ripetere containment, symlink e deny quando la struttura del progetto offre
  un caso sicuro e autorizzato.

## Gate di uscita

- zero mutazioni del workspace attribuibili a Maestro;
- zero accessi o cambiamenti Docker attribuibili a Maestro;
- tool set e permission restano entro il profilo ufficiale;
- cancellazione e limiti terminano bounded con reason code coerenti;
- scansione anti-leak dei materiali destinati al repository negativa.

Qualsiasi violazione di workspace, containment, authority o privacy è un
incidente bloccante: la raccolta si arresta e la milestone non può concludersi
come validazione positiva finché l'incidente non è compreso.

## Deliverable

- `docs/reports/milestone-13-phase-4.md`.

---

# Fase 5 — Triage, manutenzione e feedback

## Obiettivo

Trasformare osservazioni e feedback in decisioni senza confondere bug fix ed
evoluzione del prodotto.

## Attività

- classificare ogni osservazione come product bug, model limitation,
  environment, UX, evolution, security o unresolved;
- richiedere riproduzione deterministica per i bug Maestro quando possibile;
- assegnare severità, impatto, frequenza e destinazione a ogni osservazione;
- correggere sul ramo `v0.2.x` soltanto bug, documentazione e hardening entro
  il confine read-only;
- produrre v0.2.1 soltanto in presenza di correzioni concrete e dopo i gate di
  packaging, installazione, live, immutabilità e anti-leak pertinenti;
- non inserire nuove capability o promesse nella patch release;
- raccogliere feedback esterno con consenso, alias e domande coerenti con le
  metriche, senza trasferire sorgenti o output proprietari nel report;
- chiudere ogni osservazione con decisione o motivazione esplicita.

## Gate di uscita

- zero osservazioni non classificate o prive di destinazione;
- zero bug read-only bloccanti ignorati;
- eventuale v0.2.1 resta compatibile e non amplia l'autorità;
- feedback esterno separato dalle misure controllate e redatto.

## Deliverable

- `docs/reports/milestone-13-phase-5.md`;
- eventuali issue e release notes v0.2.1.

---

# Fase 6 — Report di prodotto e decisione v0.3.0

## Obiettivo

Concludere il ciclo con evidenze aggregate, limiti espliciti e una sola
decisione sul prossimo contratto di prodotto.

## Attività

- aggregare metriche per progetto, task e profilo senza uno score globale;
- pubblicare completion rate, distribuzione qualitativa, durata, turni, token,
  tool failure, invarianti e compatibilità osservata;
- descrivere sample size, missing data, deviazioni dalla coorte e limiti della
  generalizzazione;
- elencare pattern ricorrenti, bug, richieste UX e feedback esterno mantenendo
  distinta la loro provenienza;
- decidere fra Read-only Developer Experience, Controlled Mutation ed
  ecosistema/integrazioni usando i criteri sotto;
- scrivere il contratto v0.3.0 soltanto dopo l'approvazione del report e, se
  la direzione è mutativa, dopo i verdetti conclusivi delle Milestone 14 e 15;
- aggiornare roadmap, compatibility e known issues senza trasformare evidenza
  field in una promessa retroattiva di v0.2.0.

## Criteri direzionali

| Direzione | Evidenza richiesta |
|---|---|
| Read-only Developer Experience | valore osservato insieme a limiti ricorrenti di uso, retrieval, analisi multi-file, output o integrazione editor |
| Controlled Mutation | evidenza forense di un protocollo model-facing semplificabile, nuovo modello candidato, hardware superiore agli 8B correnti o miglioramento verificabile del tool calling provider; la qualificazione riparte comunque da Gate A |
| Ecosistema e integrazioni | runtime read-only stabile e domanda ricorrente per framework, editor o modalità d'uso aggiuntive |

In assenza delle condizioni mutative, la direzione raccomandata resta
Read-only Developer Experience. Le evidenze possono indicare più opportunità,
ma il contratto v0.3.0 deve dichiarare priorità e authority in modo univoco.

## Gate di uscita

- almeno due progetti reali osservati e campagna core completata o deviazioni
  dichiarate;
- zero record o osservazioni non classificati;
- invarianti di sicurezza conclusivi e nessun incidente aperto;
- report riproducibile dai dati redatti;
- decisione v0.3.0 motivata, con alternative scartate e trigger di riesame;
- nessuna nuova release richiesta per dichiarare completata la milestone.

## Deliverable

- `docs/reports/milestone-13-field-validation.md`;
- decisione di prodotto e, per la direzione mutativa, handoff alle Milestone
  14 e 15 prima del contratto v0.3.0;
- aggiornamento conclusivo di roadmap e contesto.

---

# Regola per avviare il recovery di Controlled Mutation

La Milestone 13 non riapre Gate B o Gate C e non modifica ADR-0032. Una nuova
ricerca diagnostica può iniziare nella Milestone 14 senza contare come
qualificazione. La qualificazione ufficiale appartiene alla Milestone 15 ed è
ammessa soltanto quando la recovery produce almeno uno dei seguenti input
nuovi e verificabili:

- un protocollo model-facing semplificato sostenuto da evidenza forense e da
  compilazione deterministica;
- un modello candidato diverso;
- hardware capace di eseguire un profilo superiore agli 8B attuali;
- un miglioramento osservabile del tool calling nel runtime del provider.

Il nuovo candidato riparte dal Gate A con modello, profilo, task, timeout e
criteri congelati. Il superamento del Gate A è necessario prima di Gate B e C;
le evidenze read-only della Field Validation non possono sostituire nessuno dei
tre gate.
