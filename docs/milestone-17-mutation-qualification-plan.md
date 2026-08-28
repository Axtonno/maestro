# Milestone 17 — Controlled Mutation Qualification Plan

Versione: 0.2.0

Stato: Rinviata — non apribile senza `readonly_baseline_qualified` dalla
Milestone 15, v0.3.0 verde e handoff di protocollo dalla Milestone 16

Data: 2026-08-27; aggiornamento 2026-08-28

Documenti di riferimento:

- `roadmap.md`;
- `milestone-15-reference-hardware-readonly-baseline-plan.md`;
- `milestone-16-controlled-mutation-recovery-plan.md`;
- `mutation-qualification.md`;
- `mutation-qualification-profile.yaml`;
- `adr/ADR-0031.md`;
- `adr/ADR-0032.md`;
- `security-model.md`.

---

# Condizione di apertura

La milestone resta chiusa finché non sono presenti contemporaneamente:

- baseline `direct/chat` qualificato sulla nuova piattaforma;
- verified agent sintetico interamente verde;
- B01 read-only 2/2 `correct` senza falsità materiali;
- artifact v0.3.0 read-only installabile e immutabile;
- protocollo model-facing e compilatore deterministico consegnati dalla
  Milestone 16 oppure verdetto `protocol_unchanged` motivato;
- piattaforma, hardware, provider, modello, `num_ctx`, `thinking`, filesystem e
  limiti congelati.

Stop rule:

> Un failure del baseline read-only multi-file chiude la qualificazione prima
> di Gate A. Controlled Mutation non viene usata per diagnosticare o compensare
> un modello che non comprende stabilmente il codice in read-only.

---

# Obiettivo operativo

Qualificare una sola combinazione piattaforma–hardware–provider–modello per il
vertical slice Controlled Mutation già limitato a una patch su un file PHP
esistente sotto `app/`.

La milestone non sviluppa interaction modes, non cambia il protocollo dopo il
freeze e non produce una release. Un PASS autorizza soltanto la Milestone 18.

## Preparazione mentre le milestone precedenti sono in esecuzione

Finché M15 e M16 non hanno prodotto gli handoff richiesti, sono ammessi
soltanto revisione del piano e inventario read-only delle fixture e dei gate
esistenti. Non si costruisce un candidate mutativo, non si apre una serie live,
non si seleziona un modello e non si modifica il profilo qualificato da M15.

La Fase 1 è il solo punto che può aprire l'esecuzione: importa gli handoff,
verifica che siano completi e li congela in un nuovo candidate record. Nessun
PASS parziale di M15 o M16 può essere usato al loro posto.

---

# Candidate record

Ogni serie congela:

- artifact/candidate build e commit;
- profilo Windows/WSL2/Ubuntu/filesystem Linux;
- RAM, GPU, driver, VRAM e offload;
- Ollama, modello, digest, template e quantizzazione;
- `num_ctx`, `thinking`, temperatura, timeout e limiti effettivi;
- protocollo model-facing, schema, compiler e prompt;
- fixture, modifica attesa, digest e fingerprint;
- stop rule e criteri A/B/C.

Qualsiasi variazione crea un nuovo record e azzera tutti i PASS. Non vengono
provati modelli in serie oltre una shortlist e un ordine congelati prima del
primo Gate A.

---

# Gate A — Proposta mutativa senza effetti

Tre conversazioni nuove e indipendenti devono:

- leggere il file autorevole;
- consumare il risultato della read;
- produrre una edit proposal valida tramite il canale congelato;
- compilare deterministicamente la sola patch attesa;
- lasciare la fixture byte-identica;
- completare 3/3 consecutivo e fail-fast.

Tool Runtime mutativo, preview e approval non vengono eseguiti. Il primo
failure respinge il candidate record e impedisce Gate B/C.

Deliverable: `docs/reports/milestone-17-gate-a.md`.

---

# Gate B — Conservazione read-only

Due run consecutive devono dimostrare che lo stesso candidate build e modello
conservano il verified agent read-only qualificato:

- almeno una read reale per run;
- completion corretta e terminale `completed`;
- zero pseudo-call o tool non dichiarati;
- workspace byte-identico;
- 2/2 consecutivo e fail-fast.

Gate B non sostituisce B01 della Milestone 15; verifica che l'introduzione del
percorso mutativo non abbia degradato il baseline.

Deliverable: `docs/reports/milestone-17-gate-b.md`.

---

# Gate C — Controlled Mutation

Tre fixture nuove e indipendenti devono completare:

```text
read autorevole
    -> edit proposal
    -> compilazione deterministica
    -> preview concreta
    -> approval allow-once su TTY reale
    -> apply atomico
    -> reindex
    -> contesto fresh
    -> risposta finale
```

Criteri:

- 3/3 consecutivo e fail-fast;
- una sola patch sul solo file ammesso;
- approval exact-fingerprint nuova per ogni run;
- digest finale byte-identico all'atteso;
- nessun file estraneo modificato;
- commit atomico, cleanup, reindex e freshness conformi;
- nessun retry implicito, fallback di authority o leak.

Deliverable: `docs/reports/milestone-17-gate-c.md`.

---

# Matrice negativa e di sicurezza

La matrice conserva almeno:

- deny, EOF e no-TTY;
- JSON/schema invalido e campi sconosciuti;
- path assoluto, traversal, symlink, file fuori `app/` e file non PHP;
- digest stale e modifica dopo read o preview;
- `old_text` assente, ambiguo o no-op;
- cancellazione e timeout prima/dopo commit;
- fault filesystem, sync, rename e cleanup;
- refresh/reindex fallito dopo commit;
- replay approval e secondo tentativo mutativo;
- proposta multi-file o operazione non supportata;
- anti-leak di report, log e artifact.

Ogni failure pre-commit lascia la fixture byte-identica. Ogni failure
post-commit registra stato applicato, durability e contesto stale senza
dichiarare rollback inesistente.

Deliverable: `docs/reports/milestone-17-security.md`.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Entry gate e candidate record | Bloccata | M15 + M16 |
| 2 | Gate A | Non avviata | Fase 1 |
| 3 | Gate B | Non avviata | Gate A 3/3 |
| 4 | Gate C | Non avviata | Gate B 2/2 |
| 5 | Matrice negativa e sicurezza | Non avviata | Gate C 3/3 |
| 6 | Decisione hardware–provider–modello | Non avviata | Fasi 1–5 |

---

# Fase 1 — Entry gate e candidate record

## Obiettivo

Provare che gli handoff M15 e M16 autorizzano davvero la qualificazione e
congelare una sola serie riproducibile prima di qualsiasi run mutativa.

## Attività

- verificare verdetto, report, artifact, tag e checksum v0.3.0 della M15;
- verificare B01 `2/2`, direct/chat, verified agent e sicurezza senza failure
  aperti o claim oltre la matrice eseguita;
- importare da M16 protocollo, schema, compiler, prompt, fixture, matrice
  negativa e verdetto conclusivo;
- riconciliare commit del compiler e lineage del candidate con la baseline
  read-only, senza modificare l'artifact storico v0.3.0;
- congelare piattaforma, filesystem, hardware, driver, provider, modello,
  digest, quantizzazione, context, thinking, timeout e limiti;
- congelare shortlist e ordine dei candidati prima della prima run;
- assegnare candidate ID, directory di evidenza redatta e stop rule;
- eseguire preflight senza avviare provider, scaricare modelli o mutare
  workspace e configurazioni.

## Gate di uscita

- tutti i prerequisiti sono presenti, coerenti e verificabili;
- candidate record completo, immutabile e legato a commit e digest esatti;
- fixture nuove, non sensibili e byte-identiche al baseline congelato;
- provider, GPU, offload e risorse osservati entro il profilo qualificato;
- nessuna run A/B/C avviata prima del freeze.

## Deliverable

- candidate record iniziale;
- checklist di apertura firmata dal report;
- `docs/reports/milestone-17-phase-1.md`.

---

# Fase 2 — Gate A: proposta mutativa senza effetti

## Obiettivo

Dimostrare che il candidato produce in modo ripetibile la sola proposta
ammessa e che il compiler genera la patch attesa senza concedere authority.

## Attività

- creare tre conversazioni e tre fixture indipendenti dal medesimo baseline;
- verificare read autorevole, consumo dell'evidenza e binding a run/call;
- acquisire soltanto metriche e categorie redatte previste dal candidate
  record;
- validare proposta, schema, path, occorrenza e compilazione byte-esatta;
- verificare che Tool Runtime mutativo, preview e approval non siano invocati;
- applicare fail-fast al primo risultato diverso dal contratto congelato;
- verificare digest pre/post e assenza di file aggiunti o modificati.

## Gate di uscita

- Gate A `3/3` consecutivo sullo stesso candidate record;
- patch compilata, action, diff e fingerprint identici all'oracolo;
- workspace byte-identico e zero authority mutativa esercitata;
- zero pseudo-call, repair, retry, truncation o leak.

## Deliverable

- evidenza redatta delle tre run;
- `docs/reports/milestone-17-gate-a.md`;
- `docs/reports/milestone-17-phase-2.md`.

---

# Fase 3 — Gate B: conservazione read-only

## Obiettivo

Provare che l'introduzione del percorso candidato non degrada il comportamento
verified-agent già qualificato dalla Milestone 15.

## Attività

- eseguire due conversazioni nuove con tool esclusivamente read-only;
- richiedere almeno una read reale e una risposta legata all'evidenza;
- confrontare terminali, tool lifecycle, latenza e token con i limiti
  congelati;
- verificare assenza di proposta mutativa, preview, approval e fallback;
- confrontare inventario e digest del workspace prima e dopo ogni run;
- arrestare la serie al primo failure e invalidare i PASS parziali.

## Gate di uscita

- Gate B `2/2` consecutivo, `completed` e corretto;
- nessun tool o canale mutativo osservato;
- workspace byte-identico e metriche entro il profilo;
- nessuna regressione rispetto agli invarianti read-only M15.

## Deliverable

- `docs/reports/milestone-17-gate-b.md`;
- `docs/reports/milestone-17-phase-3.md`.

---

# Fase 4 — Gate C: Controlled Mutation

## Obiettivo

Eseguire il vertical slice completo con approval umana exact-proposal e
verificare stato fisico, durability, reindex e freshness dopo il commit.

## Attività

- creare tre fixture indipendenti e ripristinate da un baseline verificato;
- eseguire la sequenza completa congelata senza scorciatoie o retry;
- usare un TTY reale e un nuovo permit `allow once` per ogni run;
- verificare che preview, fingerprint ed Execute consumino la stessa patch;
- verificare file target, digest, inventory, temporanei, sync e rename;
- verificare reindex, generazione fresh e risposta finale post-commit;
- registrare separatamente stato logico e fisico, senza rollback implicito;
- applicare fail-fast e non riusare una fixture dopo un failure.

## Gate di uscita

- Gate C `3/3` consecutivo sul candidate record invariato;
- digest finale byte-identico all'oracolo in tutte le run;
- una sola modifica autorizzata e nessun file estraneo coinvolto;
- approval one-shot non riutilizzabile, cleanup e freshness conformi;
- zero leak, repair euristico, retry o fallback di authority.

## Deliverable

- `docs/reports/milestone-17-gate-c.md`;
- `docs/reports/milestone-17-phase-4.md`.

---

# Fase 5 — Matrice negativa e sicurezza

## Obiettivo

Dimostrare che il candidate fallisce chiuso in ogni scenario obbligatorio e
descrive correttamente i failure che avvengono dopo il commit.

## Attività

- eseguire tutti i casi della matrice negativa congelata da M16;
- verificare deny, EOF, no-TTY, replay e secondo tentativo;
- verificare containment, traversal, symlink, scope PHP e ambiguità;
- iniettare stale state e fault pre-commit/post-commit previsti dal profilo;
- verificare cancellazione, timeout, sync, rename, cleanup e reindex failure;
- scandire report, stdout, stderr, log e artifact per dati non ammessi;
- rieseguire suite completa, race detector, vet e audit API.

## Gate di uscita

- tutti i casi obbligatori hanno terminale e stato fisico attesi;
- ogni failure pre-commit lascia la fixture byte-identica;
- ogni failure post-commit dichiara modifica applicata e contesto stale senza
  presentare rollback o completion;
- zero bypass di policy, containment, fingerprint o permit;
- suite repository-wide e anti-leak interamente verdi.

## Deliverable

- `docs/reports/milestone-17-security.md`;
- `docs/reports/milestone-17-phase-5.md`.

---

# Fase 6 — Decisione hardware–provider–modello

## Obiettivo

Riconciliare tutte le evidenze, emettere un solo verdetto e consegnare a M18
l'esatta combinazione qualificata oppure un rinvio non ambiguo.

## Attività

- verificare consecutività, candidate ID e immutabilità dei PASS A/B/C;
- riconciliare report Markdown/JSON, metriche, fixture digest e artifact;
- verificare che nessuna variazione di modello, protocollo, prompt, compiler,
  hardware o limiti sia avvenuta durante la serie;
- confrontare risorse osservate e limiti dichiarabili senza inferire un lower
  bound non provato;
- aggiornare o creare l'ADR hardware–provider–modello;
- eseguire audit finale di privacy, support claim e completezza;
- emettere esattamente uno degli esiti ammessi e preparare l'handoff M18 solo
  per `mutation_qualified`.

## Gate di uscita

- verdetto derivato meccanicamente dai gate e coerente con lo stato fisico;
- reference profile completo e riproducibile oppure causa di rinvio precisa;
- nessun PASS parziale reinterpretato come qualificazione;
- M18 resta chiusa per ogni esito diverso da `mutation_qualified`.

## Deliverable

- candidate record finale e ADR hardware–provider–modello;
- `docs/reports/milestone-17-phase-6.md`;
- `docs/reports/milestone-17-final.md`;
- handoff esatto alla Milestone 18 oppure rinvio motivato.

---

# Esiti ammessi

| Esito | Conseguenza |
|---|---|
| `mutation_qualified` | GO alla Milestone 18 sul candidate record esatto |
| `model_rejected` | nuovo candidato possibile soltanto entro shortlist; Gate A da zero |
| `platform_rejected` | la topologia osservata non entra nel support mutativo |
| `hardware_insufficient` | risorse impediscono i gate entro limiti congelati |
| `mutation_deferred` | protocollo o causa non qualificabili; mutazione resta non supportata |

`mutation_qualified` richiede Gate A 3/3, B 2/2, C 3/3 e matrice negativa
interamente verdi. Nessun risultato parziale autorizza una release mutativa.

---

# Deliverable

- report A/B/C e sicurezza;
- candidate record finale;
- ADR hardware–provider–modello;
- `docs/reports/milestone-17-final.md`;
- handoff esatto alla Milestone 18 oppure rinvio motivato.
