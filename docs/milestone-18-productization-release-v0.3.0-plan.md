# Milestone 18 — Productization & Release v0.3.0

Versione target: 0.3.0

Stato: Aperta — piano approvato; Fase 1 `NOT_RUN`

Data: 2026-08-29

Prerequisito: Milestone 17 completata con verdetto
`direct_chat_product_baseline`

Documenti di riferimento:

- `roadmap.md`;
- `milestone-17-direct-chat-development-plan.md`;
- `reports/milestone-17-phase-7.md`;
- `reports/milestone-17-final.md`;
- `releases/v0.3.0.md`;
- `packaging-candidate.md`;
- `installation.md`;
- `quick-start.md`;
- `compatibility.md`;
- `security-model.md`.

Il precedente `milestone-18-productization-v0.4.0-plan.md` resta una traccia
mutativa storica, non eseguibile. Il numero 18 è riassegnato da questa
decisione al rilascio prudente della baseline Direct Chat v0.3.0.

---

# Decisione di apertura

La Milestone 17 ha qualificato il packaging candidate immutabile
`v0.3.0-pc.1`, commit
`70a9630203ccf82a4d8858a9e47b48f5333b9cbd`, SHA-256
`82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61`,
sulla WSL2/Ubuntu 24.04/RTX 5070. Il verdetto autorizza release readiness, non
aggiunte funzionali.

La Milestone 18 trasforma quella baseline in release pubblicata, mantenendo il
seguente posizionamento:

> Maestro v0.3.0 è un assistente locale read-only per domande dirette, con
> zero o un file del workspace fornito esplicitamente. Non esegue strumenti,
> retrieval o modifiche.

La milestone è rigorosamente non funzionale. Productization significa
congelare identità, documentare, costruire artifact distinti, ripetere i gate,
taggare, pubblicare e verificare l’asset scaricato.

---

# Baseline immutabile

| Campo | Valore qualificato |
|---|---|
| piattaforma | Linux `amd64`; reference gate WSL2/Ubuntu 24.04/RTX 5070 |
| provider | Ollama 0.33.1 su loopback |
| modello | `qwen3.5:9b` |
| digest modello | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| profilo | schema v2 chat-only |
| context / thinking / temperatura | 4096 / `false` / 0 |
| streaming | abilitato, opt-in, output atomico |
| file / output massimi | 1 MiB / 1 MiB |
| config qualificata | SHA-256 `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee` |
| servizio/prompt | SHA-256 `7fd79e1fafb70d0b7726ecca0909f92592f8706df890a9b6fb263c9d5b8575c1` |
| fixture route | SHA-256 `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |
| authority | `workspace_mutate: deny`, zero tool e nessun fallback |

Release candidate e artifact finale devono derivare da questa baseline senza
modifiche a codice Direct Chat, prompt, schema, configurazione, modello,
digest, parametri, limiti, fixture o contratto CLI.

## Delta ammesso

Sono ammessi soltanto:

- stato/versione incorporati dal packaging;
- release notes, README, quick start, installazione, compatibility, security e
  troubleshooting coerenti con il verdetto già ottenuto;
- manifest, checksum, report di release e metadata Git/GitHub;
- sostituzione dei token documentali da parte dello script di packaging.

Una correzione al packaging è ammessa soltanto prima del freeze RC e deve
ripetere doppio build, archive audit e installazione. Una modifica che tocca il
comportamento del prodotto non appartiene alla Milestone 18: produce un nuovo
packaging candidate e riapre i gate owner della Milestone 17.

## Esclusioni non negoziabili

- supporto multi-file;
- sessioni o memoria persistente;
- nuovi modelli o digest;
- miglioramenti di agent o verified agent;
- retrieval, indicizzazione o selezione automatica;
- tool calling o plugin di terze parti;
- Controlled Mutation, write, patch o approval;
- modifiche al contratto CLI o allo schema;
- sandbox, shell, Git, Docker o multi-agent come feature di prodotto.

---

# Sequenza delle fasi

| Fase | Titolo | Stato | Dipende da |
|---:|---|---|---|
| 1 | Freeze release e audit del delta | `NOT_RUN` | M17 PASS |
| 2 | Documentazione pubblica e metadata | `NOT_RUN` | Fase 1 |
| 3 | Release candidate riproducibile | `NOT_RUN` | Fase 2 |
| 4 | Installazione pulita e audit RC | `NOT_RUN` | Fase 3 |
| 5 | Gate live RC sulla RTX 5070 | `NOT_RUN` | Fase 4 |
| 6 | Artifact finale e tag annotato | `NOT_RUN` | Fase 5 PASS |
| 7 | GitHub Release e verifica post-download | `NOT_RUN` | Fase 6 PASS |

Le fasi sono sequenziali e fail-fast. Ogni fase produce un report autonomo
sotto `docs/reports/`; la Fase 7 produce anche
`docs/reports/milestone-18-final.md`.

---

# Fase 1 — Freeze release e audit del delta

## Obiettivo

Congelare commit, modello, digest, profilo, fixture, perimetro e allowlist dei
soli cambi release.

## Attività e gate

- verificare la catena M17, gli SHA-256 e il worktree pulito;
- registrare repository GitHub target, branch e policy di tag senza modificare
  remoti;
- confrontare codice/config/fixture con il candidate qualificato;
- censire token, link, versioni e claim ancora candidati;
- congelare la lista dei file modificabili nelle Fasi 2 e 6;
- emettere `docs/reports/milestone-18-phase-1.md`.

PASS soltanto con zero delta funzionale e nessun blocker di release aperto.

---

# Fase 2 — Documentazione pubblica e metadata

## Obiettivo

Rendere release notes e guida rapida sufficienti per installazione e uso senza
checkout, mantenendo limiti e non garanzie visibili.

## Attività e gate

- finalizzare release notes, README, installation e quick start;
- allineare CLI, configuration, compatibility, security, known issues e
  troubleshooting;
- verificare comandi copiabili, link locali, token di packaging e versioni;
- dichiarare esplicitamente agent, retrieval, tool e mutation non supportati;
- rieseguire suite, race, vet e audit anti-leak;
- congelare un commit sorgente RC e produrre
  `docs/reports/milestone-18-phase-2.md`.

Nessun tag o claim di release pubblicata è ammesso in questa fase.

---

# Fase 3 — Release candidate riproducibile

## Obiettivo

Costruire un artifact distinto `v0.3.0-rc.1` con stato
`release-candidate`, senza rinominare `pc.1`.

## Attività e gate

- doppio packaging byte-identico dal commit congelato;
- checksum, manifest, versione, commit e stato coerenti;
- allowlist archive e assenza di report, raw trace, secret e profili esclusi;
- confronto binario/config/fixture/parametri con la baseline qualificata;
- conservazione locale di archive e checksum immutabili;
- report `docs/reports/milestone-18-phase-3.md`.

Qualsiasi divergenza crea un nuovo `rc.N`; un artifact non viene sovrascritto.

---

# Fase 4 — Installazione pulita e audit RC

## Obiettivo

Provare che il solo RC è installabile dalla documentazione prima di riattivare
la macchina di qualifica.

## Attività e gate

- estrazione fuori checkout in directory nuova;
- verifica checksum, manifest, `version`, root help e chat help;
- doctor chat offline fail-closed sul solo provider;
- containment pre-provider, fixture immutabile e scansione anti-leak;
- ripetizione del quick start documentale quando il provider locale è già
  disponibile, senza avviarlo implicitamente;
- report `docs/reports/milestone-18-phase-4.md`.

Un `not_run` live non è PASS, ma resta proprietario della Fase 5.

---

# Fase 5 — Gate live RC sulla RTX 5070

## Obiettivo

Usare una sola volta la piattaforma finale sullo stesso RC, senza rebuild.

## Matrice breve

- identità archive/binario/config, modello e digest;
- doctor chat 5/5;
- no-file;
- single-file complete e stream semanticamente equivalenti;
- traversal e symlink evasivo;
- SIGINT e deadline;
- fixture invariata e anti-leak.

La serie intera viene ripetuta dopo un errore harness; una failure del prodotto
respinge il RC. Nessun tuning, retry selettivo o sostituzione è ammesso. Il
report è `docs/reports/milestone-18-phase-5.md`.

Dopo il PASS la RTX 5070 torna in sospeso. Non è richiesta nuovamente per
l’artifact finale se il delta successivo resta esclusivamente metadata e
documentazione approvati.

---

# Fase 6 — Artifact finale e tag annotato

## Obiettivo

Costruire l’artifact finale v0.3.0 da un commit release pulito e legare tag,
manifest e binario alla stessa identità.

## Attività e gate

- aggiornare soltanto stato release, note e report autorizzati;
- provare che il delta dal commit RC non modifica codice/config/fixture;
- doppio packaging `v0.3.0` con stato `release` e archive byte-identico;
- audit, checksum e installazione pulita finali;
- verificare che manifest e `maestro version` incorporino il commit release;
- creare localmente il tag annotato `v0.3.0` sullo stesso commit;
- verificare annotazione e target del tag;
- produrre `docs/reports/milestone-18-phase-6.md`.

Il tag non viene pushato e l’asset non viene pubblicato prima del PASS locale.
Un artifact finale fallito non viene rinominato o sovrascritto.

---

# Fase 7 — GitHub Release e verifica post-download

## Obiettivo

Pubblicare esattamente tag, archive e checksum qualificati e verificarli dal
canale pubblico.

## Attività e gate

- verificare repository, autenticazione, tag target e assenza di release/asset
  omonimi prima di ogni write remota;
- pushare commit release e tag annotato;
- creare GitHub Release v0.3.0 con le note congelate;
- caricare archive e checksum senza overwrite;
- scaricare entrambi gli asset pubblicati in una directory pulita;
- verificare checksum, dimensione, manifest, `maestro version`, help,
  configurazione e installazione senza usare l’asset locale originario;
- verificare URL e visibilità della release;
- emettere `docs/reports/milestone-18-phase-7.md` e il report finale.

Una failure dopo la pubblicazione è un release incident: non si sostituiscono
silenziosamente tag o asset. Si conserva evidenza e si decide esplicitamente
fra correzione con nuova versione e ritiro della release.

---

# Verdetti

| Verdetto | Conseguenza |
|---|---|
| `v0.3.0_released_and_verified` | release pubblicata e asset riscaricato verificato |
| `release_candidate_failed` | nessun tag/pubblicazione; nuovo RC dopo causa dimostrata |
| `release_environment_blocked` | conservare artifact/evidenza e riprendere dopo il ripristino |
| `release_incident` | pubblicazione non verificata; nessun overwrite silenzioso |

---

# Definition of Done

La Milestone 18 è completata soltanto quando:

- baseline e delta non funzionale sono congelati;
- documentazione e posizionamento pubblico sono coerenti;
- RC distinto è riproducibile, installabile e verde sulla RTX 5070;
- artifact finale è riproducibile e installabile;
- tag annotato, commit, manifest, binario e versione coincidono;
- GitHub Release contiene esattamente archive e checksum approvati;
- gli asset pubblicati vengono riscaricati e verificati da zero;
- nessuna capability esclusa viene presentata come supportata;
- il report finale emette `v0.3.0_released_and_verified`.
