# Milestone 18 — Productization v0.4.0 Plan

Versione: 0.2.0

Stato: Non aperta — piano mutativo storico; la dipendenza
`mutation_qualified` non è stata soddisfatta e la numerazione futura deve
essere ridefinita prima di qualsiasi riapertura

Data: 2026-08-21; rinumerazione 2026-08-27; aggiornamento 2026-08-28

Nota storica: i riferimenti alla Milestone 17 contenuti nel piano seguente
indicano il precedente piano di Controlled Mutation Qualification, mai aperto
e ora superato. Non indicano la Milestone 17 Direct/Chat Product Baseline.

Documenti di riferimento:

- `roadmap.md`;
- `milestone-16-controlled-mutation-recovery-plan.md`;
- `milestone-17-mutation-qualification-plan.md`;
- `mutation-qualification.md`;
- `adr/ADR-0031.md`;
- `adr/ADR-0032.md`;
- `compatibility.md`;
- `security-model.md`;
- `packaging-candidate.md`;
- `release-readiness-audit.md`;
- `milestone-12-development-plan.md`.

---

# Condizione di apertura

La milestone può iniziare soltanto se il report finale della Milestone 17:

- contiene il verdetto esatto `mutation_qualified`;
- identifica senza ambiguità piattaforma, hardware, filesystem, provider,
  modello, digest, quantizzazione, context, protocollo e limiti qualificati;
- registra Gate A `3/3`, Gate B `2/2` e Gate C `3/3` formali e fail-fast;
- registra matrice negativa e di sicurezza interamente verde;
- non contiene incidenti di containment, authority, immutabilità, durability,
  freshness o privacy aperti;
- consegna un ADR hardware–provider–modello e un profilo candidato
  riproducibile.

Qualsiasi altro esito — `model_rejected`, `platform_rejected`,
`hardware_insufficient` o `mutation_deferred` — mantiene questa milestone
chiusa. L'esistenza del piano non costituisce un impegno a rilasciare v0.4.0.

## Preparazione mentre le milestone precedenti sono in esecuzione

Prima di `mutation_qualified` sono ammessi soltanto revisione documentale,
inventario read-only delle superfici di release e preparazione di checklist.
Non si cambia versione del prodotto, non si crea un packaging candidate, non
si modifica la configurazione pubblica e non si anticipano claim di supporto.

La pianificazione svolta durante M15/M16/M17 resta quindi priva di effetti sul
prodotto. La Fase 1 riapre e riconcilia tutte le assunzioni con il candidate
record effettivamente qualificato; qualsiasi differenza viene risolta tornando
alla Milestone 17, non adattando implicitamente la release.

---

# Obiettivo operativo

Trasformare l'esatta combinazione Controlled Mutation qualificata dalla
Milestone 17 in una release v0.4.0 installabile, documentata e pubblicata,
senza ampliare il contratto tecnico già provato.

La release conserva il profilo read-only come default e aggiunge un profilo
mutativo separato ed esplicitamente opt-in. Productization significa rendere
visibili configurazione, compatibility promise, requisito hardware, diagnosi,
security model, packaging e percorso di installazione; non significa
aggiungere nuove operazioni mutative durante i gate di release.

---

# Contratto candidato v0.4.0

## Profilo read-only

Il profilo read-only continua a esporre:

- `workspace.list`;
- `workspace.read`;
- `workspace.search`;
- `workspace_mutate: deny`.

Resta il default nella configurazione, nel quick start e nell'esperienza per
hardware che non soddisfa il profilo mutativo. Un upgrade non abilita
automaticamente Controlled Mutation.

## Profilo Controlled Mutation

Il profilo mutativo candidato espone esclusivamente il vertical slice
qualificato:

```text
read verificata
    -> edit proposal/contratto congelato
    -> patch concreta preparata
    -> preview
    -> TTY + allow once
    -> apply atomico
    -> reindex
    -> final su generazione fresh
```

Il confine resta:

- un solo file PHP esistente sotto `app/`;
- una sola sostituzione testuale esatta e non ambigua;
- una sola mutazione tentata per run;
- read autorevole e digest precondition;
- diff concreta e fingerprint exact-proposal;
- policy `workspace_mutate: prompt`;
- TTY reale e `allow once`;
- commit tramite temporaneo nella stessa directory, sync e rename atomico;
- reindex riuscito e contesto fresh prima del testo finale;
- nessun retry implicito o repair euristico.

`workspace.write`, create/delete/rename, multi-file, shell, Git, processi,
Docker, Composer, Artisan, PHPUnit, sandbox, recovery, rollback generale,
approval automatica, remote execution e multi-agent restano non supportati.

---

# Requisito hardware e piattaforma

La compatibility matrix pubblica riporta l'esatta combinazione qualificata,
non una categoria vaga come “GPU consigliata”. Almeno i seguenti campi sono
normativi:

- piattaforma e topologia, inclusa l'eventuale combinazione Windows/WSL2;
- distribuzione, kernel e architettura;
- tipo e posizione del filesystem workspace;
- RAM effettiva minima dichiarata dal contratto approvato;
- GPU, VRAM, backend e vincoli di offload;
- provider/versione/topologia;
- modello, digest o versione qualificata, quantizzazione e context;
- limiti e timeout del profilo.

Una singola macchina qualificata dimostra un reference profile, non permette
di inferire automaticamente il minimo teorico. Se la Milestone 17 non ha
qualificato un lower bound, v0.4.0 pubblica il reference hardware esatto come
requisito conservativo. Un requisito più debole richiede una nuova prova di
confine prima del freeze della compatibility matrix.

Se la Milestone 17 qualifica WSL2, la Fase 1 deve decidere tramite ADR se v0.4.0
supporta quel profilo direttamente oppure se la release richiede prima la
replica su Linux nativo con hardware equivalente. Windows nativo e workspace
sotto `/mnt/*` non vengono dedotti da un PASS WSL2.

---

# Regole trasversali

- modello, quantizzazione, protocollo, prompt, schema, hardware, topologia,
  filesystem, limiti e criteri sono quelli qualificati dalla Milestone 17;
- qualsiasi modifica a questi elementi invalida il candidate record e richiede
  una nuova qualificazione, non un hardening opportunistico nella milestone;
- il profilo mutativo è separato, opt-in e mai selezionato automaticamente;
- config preesistenti e upgrade mantengono `workspace_mutate: deny` finché
  l'utente non installa e seleziona esplicitamente il profilo mutativo;
- non esistono `--yes`, auto-approval, policy mutativa `allow`, grant
  run-scoped o fallback non interattivi;
- il diagnostic harness e i raw trace della Milestone 16 non entrano nel
  binario, nell'archive, nei log o nei documenti pubblici;
- ogni packaging candidate, release candidate e release finale è un artifact
  distinto, immutabile e identificato da versione, commit, stato e SHA-256;
- un artifact fallito non viene sovrascritto, rinominato o promosso;
- i test live non avviano provider, non scaricano modelli e non modificano
  driver, cataloghi o configurazione hardware;
- ogni Gate A/B/C usa fixture nuova e profilo esatto; fail-fast resta
  obbligatorio;
- ogni failure pre-commit lascia la fixture byte-identica; i terminali
  post-commit registrano stato applicato/stale senza rollback implicito;
- report, log e archive non includono prompt, response, arguments, diff,
  contenuti fixture, secret, root fisiche o raw telemetry identificante;
- la serie storica v0.2.x e la serie read-only v0.3.x non ricevono la nuova
  authority;
- una fase è completata soltanto con deliverable, test e gate espliciti.

---

# Superficie pubblica richiesta

La release deve pubblicare e mantenere coerenti almeno:

- configurazione read-only di default;
- configurazione Controlled Mutation opt-in;
- quick start read-only e quick start mutativo distinti;
- compatibility matrix a due profili;
- requisito/reference hardware mutativo;
- preflight e troubleshooting specifici per provider, modello, GPU e
  filesystem;
- security model aggiornato con authority, preview, approval e failure
  post-commit;
- known issues e non-garanzie;
- contratto CLI/configuration/API 0.x e note di upgrade;
- release notes v0.4.0;
- licenza, attribution, manifest, checksum e istruzioni di installazione.

Il quick start read-only viene presentato per primo. Il quick start mutativo
richiede che l'utente verifichi compatibility, copi esplicitamente il profilo
dedicato, scelga una fixture non sensibile e confermi la preview su TTY.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Entry gate e contratto di release | Condizionata | Milestone 17 `mutation_qualified` |
| 2 | Superficie pubblica e profili supportati | Non avviata | Fase 1 |
| 3 | Packaging candidate e installazione pulita | Non avviata | Fase 2 |
| 4 | Gate deterministici, operativi e di sicurezza | Non avviata | Fase 3 |
| 5 | Gate live e promozione a release candidate | Non avviata | Fase 4 |
| 6 | Qualificazione RC e freeze documentale | Non avviata | Fase 5 |
| 7 | Artifact finale, tag e pubblicazione | Non avviata | Fase 6 |

Le fasi sono sequenziali. Un failure live impedisce la promozione e richiede un
nuovo candidate soltanto quando la correzione non cambia il contratto
qualificato. Se il fix cambia modello, protocollo, schema, authority, hardware,
filesystem, limiti o criterio, il lavoro torna alla Milestone 17.

---

# Fase 1 — Entry gate e contratto di release

## Obiettivo

Provare che la condizione di apertura è soddisfatta e congelare una sola
promessa v0.4.0 prima di modificare configurazione o packaging pubblico.

## Attività

- verificare completezza e anti-leak del report finale della Milestone 17;
- verificare `mutation_qualified` e riconciliare profilo, ADR, report JSON,
  report Markdown, commit e digest binario;
- decidere supporto WSL2 oppure replica obbligatoria su Linux nativo;
- se viene richiesto Linux nativo, sospendere la Milestone 18 e riaprire la
  Milestone 17 con un nuovo candidate record; la productization riprende
  soltanto dopo `mutation_qualified` sul profilo nativo;
- congelare piattaforme, hardware, provider, modello, protocollo, filesystem,
  limiti e non-garanzie della release;
- congelare il doppio profilo read-only/mutativo e la regola di opt-in;
- censire superfici pubbliche, archive allowlist, compatibility, security,
  installation, CLI, config e release notes da aggiornare;
- eseguire baseline repository-wide e packaging v0.3.0 senza cambi per
  dimostrare un punto di partenza verde;
- approvare un ADR di productization v0.4.0.

## Gate di uscita

- condizione `mutation_qualified` verificata e non reinterpretata;
- piattaforma mutativa di release decisa con evidenza sufficiente;
- support claim ed esclusioni completi;
- nessuna capability oltre il vertical slice qualificato;
- baseline read-only e packaging verdi.

## Deliverable

- ADR del contratto v0.4.0;
- matrice di handoff dalla Milestone 17;
- `docs/reports/milestone-18-phase-1.md`.

---

# Fase 2 — Superficie pubblica e profili supportati

## Obiettivo

Rendere configurazione, diagnostica e documentazione coerenti con l'autorità
aggiuntiva senza alterare il default read-only.

## Attività

- mantenere l'esempio read-only come configurazione predefinita;
- promuovere una sola configurazione mutativa separata, strict e opt-in con
  tool/permission/limiti esatti;
- impedire la combinazione accidentale di tool mutativi con profili hardware,
  modelli, filesystem o policy non supportati;
- aggiungere preflight mutativo per identità piattaforma, filesystem, provider,
  modello, quantizzazione/capability e configurazione;
- aggiornare README, quick start, installation, configuration, CLI,
  compatibility, security model, troubleshooting e known issues;
- descrivere chiaramente pre-commit, commit, post-commit/stale e assenza di
  rollback generale;
- aggiornare API compatibility e note di migrazione 0.x;
- aggiungere release notes v0.4.0 inizialmente candidate;
- aggiornare allowlist del packaging per includere soltanto configurazione e
  documenti mutativi approvati;
- aggiungere test che provino default deny, opt-in esplicito e rifiuto dei
  profili non qualificati.

## Gate di uscita

- installazione/upgrade non abilita mutazioni implicitamente;
- profilo read-only invariato e profilo mutativo esatto;
- preflight fallisce chiuso fuori dal reference profile;
- compatibility e requisito hardware coincidono con la Milestone 17;
- nessun diagnostic payload o materiale sperimentale entra nell'archive.

## Deliverable

- configurazioni pubbliche v0.4.0;
- superficie documentale candidate;
- `docs/reports/milestone-18-phase-2.md`.

---

# Fase 3 — Packaging candidate e installazione pulita

## Obiettivo

Produrre un `v0.4.0-pc.N` riproducibile e installarlo fuori dal checkout sulla
piattaforma di release qualificata.

## Attività

- adeguare script, manifest, guide renderizzate e allowlist a v0.4.0;
- costruire due volte da worktree pulito con input normalizzati;
- confrontare archive e checksum byte per byte;
- verificare inventory, permessi, path, link, licenza e attribution;
- scandire l'archive contro source interni, diagnostic harness, raw trace,
  secret, path fisici e configurazioni non supportate;
- estrarre in una directory nuova sul filesystem qualificato;
- installare binario e configurazioni senza dipendere dal checkout;
- verificare `version`, help, `doctor`, `models` e `agents` per entrambi i
  profili;
- verificare che il profilo mutativo richieda selezione esplicita e che quello
  read-only continui a negare ogni mutazione;
- registrare identità e digest del candidate senza presentarlo come RC.

## Gate di uscita

- doppio build byte-identico e checksum valido;
- installazione pulita e manifest coerente;
- preflight read-only e mutativo sul reference profile;
- archive completo, redatto e privo di materiale development-only;
- candidate immutabile con stato `packaging-candidate`.

## Deliverable

- `v0.4.0-pc.N` e checksum fuori dal repository;
- evidenza di installazione pulita;
- `docs/reports/milestone-18-phase-3.md`.

---

# Fase 4 — Gate deterministici, operativi e di sicurezza

## Obiettivo

Verificare sull'archive installato tutti gli invarianti che non richiedono un
PASS generativo completo prima di concedere authority al Gate C live.

## Attività

- rieseguire suite completa, race detector, vet e audit API;
- eseguire matrice deterministica positiva e negativa Controlled Mutation;
- verificare containment, path, symlink, read/digest, ambiguità, preview,
  fingerprint, permit one-shot, replay e secondo tentativo;
- verificare fault pre-rename, sync, rename atomico, cleanup e terminali
  post-commit/reindex;
- verificare deny, EOF, no-TTY, input invalido, SIGINT, deadline e hard limit;
- verificare rifiuto di platform/filesystem/hardware/modello fuori profilo;
- verificare che ogni failure pre-commit lasci la fixture byte-identica;
- verificare che failure post-commit non dichiarino rollback o successo;
- verificare profilo read-only deny e workspace invariato;
- eseguire scansione anti-leak di archive, stdout, stderr e report.

## Gate di uscita

- matrice deterministica interamente verde;
- terminali e stato fisico esatti per ogni scenario;
- zero fallback di authority, retry, auto-approval o repair euristici;
- profilo read-only invariato;
- zero leak e gate repository-wide verdi.

## Deliverable

- matrice operativa e di sicurezza del candidate;
- `docs/reports/milestone-18-phase-4.md`.

---

# Fase 5 — Gate live e promozione a release candidate

## Obiettivo

Ripetere sul packaging candidate installato l'intera qualificazione live e
produrre un release candidate distinto soltanto dopo il PASS.

## Attività

- congelare candidate, provider, modello/digest, hardware, filesystem,
  fixture, prompt, schema, temperatura, context, timeout e limiti;
- eseguire preflight live senza avviare provider o scaricare modelli;
- eseguire Gate A `3/3` formale, consecutivo e fail-fast, senza effetti;
- eseguire Gate B `2/2` read-only, consecutivo e fail-fast;
- eseguire Gate C `3/3`, consecutivo e fail-fast, con TTY e `allow once`;
- eseguire gli scenari negativi live richiesti dal profilo qualificato;
- registrare RAM, VRAM/offload, latenza, turni, token, tool call, lifecycle,
  digest, freshness e cleanup in forma redatta;
- arrestare la serie al primo failure e mantenere immutabile il candidate;
- dopo il PASS completo, costruire un distinto `v0.4.0-rc.N` dallo stesso
  source candidate congelato senza rinominare il packaging candidate.

## Gate di uscita

- Gate A `3/3`, B `2/2` e C `3/3` completi sull'archive installato;
- matrice negativa live verde;
- profilo risorse entro il contratto dichiarato;
- stato fisico, approval e freshness esatti;
- release candidate distinto, riproducibile e immutabile.

## Deliverable

- `v0.4.0-rc.N` e checksum;
- report live del packaging candidate;
- `docs/reports/milestone-18-phase-5.md`.

---

# Fase 6 — Qualificazione RC e freeze documentale

## Obiettivo

Confermare che il release candidate distribuisce lo stesso prodotto qualificato
e congelare tutta la documentazione pubblica prima dell'artifact finale.

## Attività

- costruire due volte il release candidate e verificarne riproducibilità;
- installarlo da estrazione fresca fuori dal checkout;
- ripetere preflight e Gate A `3/3`, B `2/2`, C `3/3` sul RC esatto;
- ripetere deny, no-TTY, digest stale, modifica dopo preview, cancellazione e
  refresh failure come matrice live critica;
- verificare profilo read-only con due quick start consecutivi e immutabilità;
- riconciliare artifact, manifest, binario, configurazioni, modello, hardware,
  compatibility e security model;
- finalizzare README, quick start, installazione, compatibility, security,
  troubleshooting, known issues, API compatibility e release notes;
- rimuovere ogni linguaggio candidate dalle superfici destinate alla release;
- congelare il commit documentale e vietare ulteriori cambiamenti funzionali.

## Gate di uscita

- RC riproducibile, installabile e interamente qualificato;
- entrambi i profili pubblici superano i rispettivi gate;
- matrice live critica, immutabilità e anti-leak verdi;
- documentazione pubblica coerente e congelata;
- nessun failure, warning ambiguo o support claim non provato.

## Deliverable

- release candidate qualificato;
- commit di freeze documentale;
- `docs/reports/milestone-18-phase-6.md`.

---

# Fase 7 — Artifact finale, tag e pubblicazione

## Obiettivo

Costruire, verificare e distribuire v0.4.0 come artifact finale distinto dal
release candidate.

## Attività

- creare un commit pulito, discendente e successivo al freeze documentale;
- costruire due volte `maestro-v0.4.0-linux-amd64.tar.gz` con stato `release`;
- confrontare archive/checksum byte per byte e verificare inventory/manifest;
- installare l'artifact finale in directory pulita sul filesystem supportato;
- ripetere preflight e l'intera sequenza Gate A `3/3`, B `2/2`, C `3/3` sul
  final artifact;
- ripetere quick start read-only, deny, no-TTY, SIGINT, hard limit,
  immutabilità e anti-leak;
- verificare che `maestro version`, manifest e nome archive concordino;
- creare il tag annotato `v0.4.0` soltanto dopo tutti i PASS e verificare che
  punti al commit incorporato nel binario;
- pubblicare commit finale e tag senza spostarli o ricrearli;
- creare la GitHub Release con note, archive e checksum esatti;
- riscaricare gli asset pubblici in una nuova directory, verificare SHA-256,
  installazione, `version`, preflight e una conferma mutativa controllata;
- pubblicare il report finale senza prompt, response, diff o dati diagnostici.

## Gate di uscita

- artifact finale riproducibile e distinto dal RC;
- Gate A/B/C finali, read-only e sicurezza interamente verdi;
- tag, commit, binario, manifest, archive e checksum coerenti;
- GitHub Release pubblica con asset esatti;
- installazione e conferma dall'asset riscaricato;
- zero leak, incidenti o claim oltre la matrice qualificata.

## Deliverable

- `maestro-v0.4.0-linux-amd64.tar.gz` e checksum;
- tag annotato e GitHub Release `v0.4.0`;
- `docs/releases/v0.4.0.md`;
- `docs/reports/milestone-18-final.md`.

---

# Stop rule e possibili esiti

La milestone può concludere con:

| Esito | Significato |
|---|---|
| `release_published` | v0.4.0 mutativa pubblicata sul profilo esatto |
| `productization_blocked` | contratto qualificato ma gate di prodotto non superato |
| `requalification_required` | una correzione necessaria cambia il candidate record e torna alla Milestone 17 |

Un failure non autorizza una v0.4.0 read-only con lo stesso contratto nominale,
né l'indebolimento di Gate A/B/C o del security model. Una release read-only
alternativa richiederebbe una decisione e un piano distinti.
