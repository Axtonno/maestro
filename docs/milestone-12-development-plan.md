# Milestone 12 — Productization v0.2.0 Development Plan

Versione: 0.1.0

Stato: Avviata — Fasi 1–5 completate, Fase 6 da avviare

Data: 2026-08-21

Documenti di riferimento:

- `roadmap.md`;
- `v0.2.0-development-plan.md`;
- `reports/milestone-11-final.md`;
- `adr/ADR-0032.md`;
- `packaging-candidate.md`;
- `release-readiness-audit.md`;
- `compatibility.md`;
- `security-model.md`.

---

# Obiettivo operativo

Consegnare v0.2.0 come release Linux `amd64` installabile, riproducibile e
qualificata esclusivamente sul percorso Ollama/Laravel read-only già
supportato. La release conserva list, read e search come unico tool set
pubblico e `workspace_mutate: deny` come confine di autorità.

Il GO della Milestone 11 non autorizza la productization della Controlled
Mutation. `workspace.write`, `workspace.patch`, approval mutativa,
`ibm/granite4.1:8b` e il profilo mutante restano sperimentali e fuori da
configurazione inclusa, quick start, compatibility promise e artifact
pubblicato. La loro presenza nel repository non costituisce evidenza di
supporto.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratto di release e baseline | Completata | Milestone 11 |
| 2 | Superficie di prodotto read-only | Completata | Fase 1 |
| 3 | Packaging candidate e installazione pulita | Completata | Fase 2 |
| 4 | Gate operativi, sicurezza e anti-leak | Completata | Fase 3 |
| 5 | Qualificazione live e release candidate | Completata | Fasi 3–4 |
| 6 | Documentazione, release finale e tag | Pianificata | Fase 5 |

Le fasi sono sequenziali rispetto al gate. Ogni fase produce un report sotto
`docs/reports/`; la Fase 6 produce anche
`docs/reports/milestone-12-final.md`. Un failure blocca la promozione ma non
viene corretto modificando retroattivamente task, modello, timeout o criteri
del candidate già identificato.

---

# Regole trasversali

- il support claim resta Linux `amd64`, Ollama, `llama3.1:8b`,
  `embeddinggemma:latest` e reference agent Laravel read-only;
- la configurazione pubblicata registra soltanto `workspace.list`,
  `workspace.read` e `workspace.search`, con `workspace_mutate: deny`;
- il file `configs/maestro.mutating.example.yaml` resta materiale sperimentale
  del repository e non viene incluso nell'archive o presentato come esempio
  supportato;
- nessuna prova deterministica sostituisce i quick start con provider e modello
  reali; un prerequisito assente produce `skipped`, mai PASS;
- packaging candidate, release candidate e release sono artifact distinti;
  non vengono rinominati, sovrascritti o promossi per sola equivalenza nominale;
- ogni artifact registra versione, commit, piattaforma, stato e checksum; il
  binario deve esporre la stessa identità tramite `maestro version`;
- prompt, response, contenuti del workspace, arguments tool, secret e path
  fisici non entrano nei report o nei log pubblicati;
- i test live non avviano Ollama, non scaricano modelli e non modificano la
  configurazione del provider;
- shell agentica, Git agentico, sandbox, recovery, multi-agent, tool esterni e
  packaging di plugin terzi restano fuori scope;
- una fase è completata soltanto con codice o documentazione necessari, test e
  verifica esplicita del gate di uscita.

---

# Fase 1 — Contratto di release e baseline

Stato: Completata.

## Obiettivo

Tradurre il verdetto `mutation_deferred` della Milestone 11 in un contratto di
productization non ambiguo e verificare che la baseline da cui parte la release
sia verde e ancora impacchettabile.

## Attività

- congelare support claim, esclusioni, piattaforma e provider della release;
- censire README, configurazioni, documentazione pubblica e contenuto archive;
- identificare i delta v0.2.0 senza ampliare l'autorità del profilo incluso;
- verificare suite completa, race detector e `go vet`;
- ripetere il gate di packaging riproducibile della release v0.1.1;
- definire dipendenze, stop rule, evidenze e deliverable delle fasi successive.

## Gate di uscita

- ADR-0032 è riflessa nel piano senza claim mutativi residui;
- la baseline deterministica e il packaging esistente sono verdi;
- i delta di productization hanno owner e fase assegnati;
- nessun artifact v0.2.0 viene prodotto prima dell'aggiornamento della
  superficie pubblica.

## Deliverable

- questo piano di sviluppo;
- `docs/reports/milestone-12-phase-1.md`.

Gate: **superato**.

---

# Fase 2 — Superficie di prodotto read-only

Stato: Completata.

## Obiettivo

Rendere configurazione, documentazione inclusa e contratti di compatibilità
coerenti con v0.2.0, mantenendo impossibile l'attivazione accidentale di tool
mutanti dal percorso pubblicato.

## Attività

- aggiornare README, changelog, release notes e riferimenti di versione;
- aggiornare compatibility matrix, security model, known issues e contratto
  API sperimentale per v0.2.0;
- verificare che quick start, configurazione e CLI descrivano soltanto il
  percorso read-only qualificabile;
- separare in modo esplicito la documentazione sperimentale Controlled Mutation
  dalla superficie distribuita;
- aggiornare la allowlist dei documenti dell'archive e i relativi test;
- aggiungere controlli che rifiutino `workspace.write`, `workspace.patch` o
  `workspace_mutate` diverso da `deny` nella configurazione inclusa;
- eseguire suite completa, race detector, vet e scansione dei claim pubblici.

## Gate di uscita

- ogni superficie destinata all'archive dichiara v0.2.0 e lo stesso support
  claim read-only;
- nessun esempio o guida pubblicata suggerisce supporto mutativo o Granite;
- il profilo incluso espone esattamente list/read/search e deny mutativo;
- gli eventuali breaking change 0.x sono dichiarati, altrimenti è registrata
  esplicitamente l'assenza di breaking change pubblici.

## Deliverable

- superficie documentale e configurazione v0.2.0;
- test del confine read-only pubblicato;
- `docs/reports/milestone-12-phase-2.md`.

Gate: **superato**.

---

# Fase 3 — Packaging candidate e installazione pulita

Stato: Completata — `v0.2.0-pc.1`.

## Obiettivo

Produrre un `v0.2.0-pc.N` identificabile e riproducibile, installarlo fuori dal
checkout e dimostrare che contiene soltanto la superficie supportata.

## Attività

- adeguare script e manifest al contenuto pubblico v0.2.0;
- costruire due volte il candidate da worktree pulito con input normalizzati e
  confrontare archive e checksum byte per byte;
- verificare path archive, permessi, licenze, attribution, fixture e assenza di
  file o token proibiti;
- installare il binario in una directory pulita priva del checkout;
- eseguire `version`, help, `doctor`, `models` e `agents` dall'installazione;
- confrontare versione e commit tra nome archive, manifest e binario;
- provare che configurazione mutante e documentazione interna non siano state
  incluse.

## Gate di uscita

- doppio build byte-identico e checksum valido;
- installazione e CLI di preflight funzionanti fuori dal checkout;
- archive completo, senza path di build, credenziali o capability mutative
  pubblicate;
- il candidate resta `packaging-candidate` e non viene presentato come RC.

## Deliverable

- `v0.2.0-pc.N` e checksum conservati fuori dal repository;
- evidenza di installazione pulita;
- `docs/reports/milestone-12-phase-3.md`.

Gate: **superato**.

---

# Fase 4 — Gate operativi, sicurezza e anti-leak

Stato: Completata — `v0.2.0-pc.1`.

## Obiettivo

Verificare sul packaging candidate i terminali e gli invarianti di sicurezza
che non richiedono un esito generativo positivo.

## Attività

- verificare deny mutativo e workspace byte-identico;
- verificare EOF, stdin invalido e assenza di TTY senza loop o approval
  implicite;
- verificare SIGINT/SIGTERM, deadline, hard limit e shutdown bounded;
- verificare codici di uscita, stdout/stderr e terminal reason stabili;
- scandire archive, output ed evidenze contro secret, prompt, response,
  contenuti fixture, arguments tool e path fisici;
- rieseguire i controlli di containment, symlink e configurazione strict
  rilevanti al profilo incluso;
- registrare ogni caso come passed, failed, skipped o unsupported senza
  trasformare un caso non eseguito in PASS.

## Gate di uscita

- deny, EOF, no-TTY, SIGINT, deadline e hard limit hanno tutti evidenza;
- ogni failure pre-effetto lascia il workspace byte-identico;
- nessun leak o fallback di autorità viene osservato;
- suite, race detector, vet e `git diff --check` sono verdi.

## Deliverable

- matrice operativa e di sicurezza del candidate;
- report anti-leak redatto;
- `docs/reports/milestone-12-phase-4.md`.

Gate: **superato**.

---

# Fase 5 — Qualificazione live e release candidate

Stato: Completata — `v0.2.0-rc.1`.

## Obiettivo

Qualificare l'esatto percorso supportato con Ollama e modello reali, quindi
produrre e congelare un release candidate distinto.

## Attività

- congelare candidate, task, fixture, modello, endpoint redatto, timeout e
  hard limit prima della prima run;
- eseguire il preflight live senza avviare provider o scaricare modelli;
- eseguire dall'archive installato almeno due quick start read-only consecutivi
  con una read reale, risposta semanticamente corretta e digest invariato;
- verificare sul percorso live cancellazione, deadline e hard limit che
  richiedono il provider;
- arrestare la serie al primo failure e classificare prodotto, modello,
  ambiente oppure operatore;
- soltanto dopo il PASS del candidate costruire un distinto
  `v0.2.0-rc.N` dallo stesso commit e ripetere installazione, preflight e due
  quick start consecutivi sull'esatto archive RC;
- congelare nome, commit e SHA-256 del release candidate; qualsiasi correzione
  richiede un nuovo `pc.N` e un nuovo `rc.N`.

## Gate di uscita

- l'esatto RC supera due quick start consecutivi e tutti i gate live richiesti;
- tool set osservato, policy e stato fisico restano read-only;
- report e log sono redatti e la scansione anti-leak è negativa;
- il release candidate è immutabile e non viene rinominato come release;
- nessun blocker di release resta aperto.

## Deliverable

- evidenze live del candidate e dell'RC;
- identità immutabile del `v0.2.0-rc.N` qualificato;
- `docs/reports/milestone-12-phase-5.md`.

Gate: **superato**.

---

# Fase 6 — Documentazione, release finale e tag

Stato: In corso — documentazione pubblica congelata al commit
`fac2ae347d9fd6e03e9faef466d11bafa961370c`.

## Obiettivo

Chiudere la documentazione sulla base dell'evidenza RC, produrre un artifact
finale da un commit successivo alla documentazione pubblica e verificarne
l'identità fino al tag.

## Attività

- finalizzare README, changelog, note di release, installazione, quick start,
  compatibility, security model, known issues e troubleshooting;
- registrare in modo esplicito Controlled Mutation e Granite come non
  supportati, senza usare i risultati deterministici come claim live;
- congelare e committare la documentazione pubblica prima della build finale;
- produrre `v0.2.0` con stato `release` da un distinto commit pulito discendente
  dal commit documentale, senza prerelease nel nome;
- ripetere doppio packaging, installazione pulita, CLI, anti-leak e quick start
  di conferma sull'artifact finale;
- verificare che archive, checksum, manifest e `maestro version` concordino;
- creare il tag annotato `v0.2.0` soltanto dopo il PASS e verificare che il tag
  punti al commit incorporato nel binario e dichiarato nel manifest;
- produrre il report finale con verdetto GO/NO-GO e catena delle evidenze.

## Gate di uscita

- documentazione pubblica completa prima del commit di build finale;
- artifact finale distinto dall'RC, riproducibile e verificato fuori dal
  checkout;
- conferma live read-only positiva e workspace invariato;
- tag, commit del binario e manifest coincidono esattamente;
- suite, race detector, vet, anti-leak e audit documentale sono verdi;
- nessuna superficie mutativa entra nel support claim v0.2.0.

## Deliverable

- archive `maestro-v0.2.0-linux-amd64.tar.gz` e checksum;
- tag annotato verificato `v0.2.0`;
- `docs/releases/v0.2.0.md`;
- `docs/reports/milestone-12-phase-6.md`;
- `docs/reports/milestone-12-final.md`.

---

# Gate finale della milestone

Milestone 12 è completata soltanto se l'artifact finale soddisfa tutti i gate
deterministici e live sul confine read-only, è costruito da un commit
successivo alla documentazione pubblica e il tag verificato punta allo stesso
commit incorporato. Qualsiasi claim mutativo, quick start non consecutivo,
artifact rinominato o evidenza mancante produce NO-GO alla release, non una
deroga ai criteri.
