# Milestone 11 — Mutation Qualification Development Plan

Versione: 0.1.0

Stato: In corso — Fasi 1–7 pianificate

Data: 2026-08-21

Documenti di riferimento:

- `roadmap.md`;
- `v0.2.0-development-plan.md`;
- `reports/milestone-10-final.md`;
- `mutation-qualification-profile.yaml`;
- `adr/ADR-0031.md`;
- `benchmark-runtime.md`;
- `developer-benchmark.md`;
- `benchmark-reporting.md`.

---

# Obiettivo operativo

Qualificare dal vivo il vertical slice Controlled Mutation consegnato dalla
Milestone 10:

```text
read -> prepare patch -> preview -> approval -> apply -> reindex -> final
```

La milestone deve stabilire se Linux `amd64`, Ollama,
`ibm/granite4.1:8b` e il lower bound hardware dichiarato possono sostenere il
percorso mutativo completo entro i limiti congelati. La prova deve attraversare
il prodotto reale, inclusi preview TTY, `allow once`, commit atomico, reindex e
terminale finale; la sola copertura deterministica non costituisce
qualificazione live.

La milestone non amplia il contratto della Milestone 10 e non effettua
productization. Configurazioni finali, packaging, installazione pulita e
release candidate appartengono alla Milestone 12.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratto di qualificazione e baseline | Pianificata | Milestone 10 |
| 2 | Developer Benchmark mutativo | Pianificata | Fase 1 |
| 3 | Matrice deterministica e candidato congelato | Pianificata | Fase 2 |
| 4 | Preflight live e Gate A | Pianificata | Fase 3 |
| 5 | Gate B read-only | Pianificata | Fase 4 |
| 6 | Gate C Controlled Mutation | Pianificata | Fase 5 |
| 7 | Audit, verdetto e handoff | Pianificata | Fasi 1–6 |

Le fasi sono sequenziali rispetto ai gate. Un failure in Gate A impedisce Gate
B; un failure in Gate B impedisce Gate C; il primo failure di una serie arresta
i tentativi successivi della stessa serie. Una fase live non può correggere
prompt, timeout, budget, fixture o criteri: qualsiasi modifica invalida il
candidato e richiede un nuovo congelamento dalla Fase 3.

Ogni fase produce un report sotto `docs/reports/`. I report live conservano
soltanto evidenze redatte e riferimenti riproducibili; prompt, risposte complete,
arguments, risultati tool, credenziali e path fisici restano esclusi.

---

# Regole trasversali

- il profilo autorevole resta `mutation-qualification-profile.yaml`;
- il candidato iniziale è Linux `amd64`, Ollama,
  `ibm/granite4.1:8b`, Intel Core i5-8365U, 8 CPU logiche, 15 GiB RAM e 4 GiB
  swap;
- Gate A richiede `3/3`, Gate B `2/2` e Gate C `3/3`, tutti consecutivi e
  fail-fast;
- temperatura, prompt, tool schema, fixture, deadline, hard limit e cleanup
  vengono congelati prima della prima run ufficiale;
- ogni run usa una fixture privata appena materializzata e ne registra digest
  iniziale, digest atteso e digest osservato finale;
- una run di failure deve lasciare i file candidati byte-identici, salvo gli
  scenari esplicitamente post-commit, che devono registrare patch applicata e
  contesto stale senza dichiarare successo;
- nessuna approval viene fabbricata: il Gate C attraversa una TTY reale e un
  operatore approva una sola volta la preview concreta;
- fake, seam e approver controllati sono ammessi nei test deterministici, ma
  non contano come PASS live;
- le prove non avviano Ollama, non scaricano modelli, non modificano la
  configurazione del provider e non promuovono un prerequisito assente a PASS;
- ogni tentativo registra identità del commit e del binario, provider e modello,
  hardware, durata, contatori, terminale, lifecycle redatto e stato fisico;
- `workspace.write`, create/delete/rename, multi-file, shell, Git, processi,
  sandbox, recovery, plugin/tool terzi e multi-agent restano fuori scope.

---

# Modello delle evidenze

Ogni scenario deve produrre uno dei seguenti stati:

| Stato | Significato |
|---|---|
| `passed` | Tutti i criteri congelati sono soddisfatti |
| `failed` | Il prodotto o il modello viola un criterio del candidato |
| `skipped` | La run non è partita per un prerequisito dichiarato assente |
| `unsupported` | Il profilo non espone una capability richiesta |

`skipped` e `unsupported` non soddisfano un gate. Ogni failure viene inoltre
classificato come prodotto, modello, ambiente, operatore o harness. Una causa
non classificata impedisce il verdetto finale.

Per ogni campione vengono conservati soltanto:

- ID di profilo, gate, scenario e tentativo;
- commit, digest del binario e versioni dichiarate;
- profilo hardware e endpoint redatto;
- stato, reason code, exit code e durata;
- numero di turni, tool call e tentativi mutativi;
- ordine redatto di proposal, approval, apply, reindex e terminale;
- digest iniziale, atteso e finale della fixture;
- freshness del contesto e stato del cleanup;
- classificazione dell'eventuale failure.

---

# Fase 1 — Contratto di qualificazione e baseline

## Obiettivo

Trasformare il profilo consegnato dalla Milestone 10 in un protocollo eseguibile
e non ambiguo, prima di aggiungere codice benchmark o avviare prove live.

## Attività

- verificare coerenza tra ADR-0031, profilo YAML, configurazione mutativa e
  implementazione corrente;
- congelare task, file candidato, sostituzione esatta, digest, output atteso e
  criteri semantici dei Gate A/B/C;
- fissare timeout provider, deadline di run, hard limit, temperatura, streaming,
  turni massimi e tool call massime;
- definire prerequisiti, preflight e criteri di stop senza fallback automatici;
- definire chi approva, come viene attestata la TTY e quali dati dell'approval
  possono essere conservati;
- mappare tutti gli scenari di `mutation_matrix` a un livello di prova
  deterministico, live o entrambi;
- definire schema e retention delle evidenze, redazione e scansione anti-leak;
- eseguire la baseline repository-wide senza cambiare il support claim.

## Gate di uscita

- nessun criterio di accettazione resta implicito;
- comandi, ordine, stop rule e cleanup sono riproducibili;
- ogni scenario del profilo ha un owner e un livello di prova;
- la baseline deterministica è verde;
- Controlled Mutation resta candidato non supportato.

## Deliverable

- specifica operativa del benchmark mutativo;
- eventuale aggiornamento backward-compatible del profilo;
- `docs/reports/milestone-11-phase-1.md`.

---

# Fase 2 — Developer Benchmark mutativo

## Obiettivo

Implementare una suite versionata che eserciti il percorso mutativo Laravel e
produca evidenze machine-readable senza incorporare contenuti sensibili.

## Attività

- aggiungere manifest, loader e validazione del profilo mutativo;
- introdurre scenari distinti per protocollo diretto, reference agent
  read-only e Controlled Mutation;
- materializzare una fixture privata nuova per ogni campione e rimuoverla con
  cleanup bounded;
- calcolare prima della run i digest dell'intero workspace rilevante e il digest
  finale ammesso per il solo scenario positivo;
- osservare lifecycle mutativo, terminale, context freshness, commit e cleanup
  senza registrare payload proibiti;
- rappresentare Gate A/B/C, serie consecutive, fail-fast e tentativi non
  eseguiti senza usare uno score aggregato come sostituto del gate;
- estendere JSON canonico e Markdown deterministico soltanto se il contratto
  esistente non può esprimere le nuove evidenze;
- aggiungere test di validazione, report round-trip, permessi `0600`, atomicità
  di pubblicazione e anti-leak.

## Gate di uscita

- il benchmark rifiuta profili incompleti o incoerenti;
- fixture e report non condividono stato fra campioni;
- il fail-fast è verificato deterministicamente;
- JSON e Markdown descrivono lo stesso esito senza prompt, response o path
  fisici;
- un approver di test non può essere scambiato per evidenza Gate C live.

## Deliverable

- suite Developer Benchmark mutativa e relativo manifest;
- documentazione di esecuzione e reporting;
- `docs/reports/milestone-11-phase-2.md`.

---

# Fase 3 — Matrice deterministica e candidato congelato

## Obiettivo

Dimostrare che harness e prodotto classificano correttamente ogni esito fisico
prima di usare provider e modello live, quindi congelare il candidato ufficiale.

## Attività

- coprire `positive_exact_patch`, stale digest, traversal, symlink, deny, EOF,
  no-TTY, input invalido, cancellazione prima e dopo commit, fault filesystem,
  refresh failure, tool non dichiarato, replay e secondo tentativo mutativo;
- verificare proposal, approval, apply, reindex e terminale nell'ordine atteso;
- verificare byte-identità delle failure pre-commit e digest esatto del caso
  positivo;
- verificare stato applicato/stale dei failure post-commit e assenza di testo
  finale;
- verificare cleanup dei temporanei e assenza di retry impliciti;
- eseguire suite completa, race detector, vet, validazione manifest, round-trip
  dei report e diff check;
- congelare commit, digest del binario di qualificazione, profilo, fixture,
  prompt, limiti e comandi ufficiali.

## Gate di uscita

- l'intera `mutation_matrix` è verde deterministicamente;
- nessun contenuto parziale, authority riutilizzata o successo su contesto
  stale;
- il profilo read-only resta invariato;
- il checkout del candidato è pulito e l'identità del binario è registrata;
- ogni modifica successiva al candidato obbliga a ripetere questa fase.

## Deliverable

- report deterministico JSON e Markdown;
- record immutabile del candidato di qualificazione;
- `docs/reports/milestone-11-phase-3.md`.

---

# Fase 4 — Preflight live e Gate A

## Obiettivo

Verificare ambiente, provider e modello e dimostrare il protocollo diretto
`read -> result -> patch` senza eseguire effetti.

## Attività

- raccogliere hardware, sistema operativo, versione Ollama, model ID e digest
  del candidato senza avviare o installare servizi;
- verificare risorse, endpoint, modello, capability tool calling, configurazione
  e spazio temporaneo;
- eseguire tre conversazioni indipendenti a temperatura congelata;
- richiedere una sola tool call nativa di read, validarne nome, schema e path,
  quindi fornire dall'harness contenuto e digest autorevoli;
- richiedere una sola proposta patch e validarne path, digest, occorrenza e
  sostituzione;
- non invocare il Tool Runtime e verificare la fixture byte-identica;
- arrestare la serie al primo failure e classificare l'esito.

## Gate di uscita

- preflight interamente superato;
- Gate A superato `3/3` con sequenze consecutive valide;
- zero approval e zero effetti;
- fixture invariata e report redatto;
- in caso di failure, Gate B e Gate C non vengono avviati.

## Deliverable

- evidenza live Gate A JSON e Markdown;
- `docs/reports/milestone-11-phase-4.md`.

---

# Fase 5 — Gate B read-only

## Obiettivo

Confermare che il candidato conserva il percorso reference agent read-only
prima di concedere autorità mutativa.

## Attività

- materializzare una fixture nuova per ciascun tentativo;
- eseguire due run consecutive con reference agent, streaming, contesto e
  limiti congelati;
- richiedere una read reale e una risposta finale semanticamente verificabile;
- rifiutare tool non dichiarati, pseudo-call e qualsiasi richiesta mutativa;
- verificare terminale `completed`, exit code, contatori, durata e cleanup;
- confrontare l'intero workspace prima e dopo ogni run;
- arrestare il gate al primo failure e non avviare Gate C.

## Gate di uscita

- Gate B superato `2/2` con run consecutive valide;
- ogni run usa almeno una read reale e nessun effetto;
- workspace byte-identico e contesto coerente;
- nessun contenuto sensibile nei report.

## Deliverable

- evidenza live Gate B JSON e Markdown;
- `docs/reports/milestone-11-phase-5.md`.

---

# Fase 6 — Gate C Controlled Mutation

## Obiettivo

Dimostrare tre volte consecutive il vertical slice completo sul prodotto reale
con approval umana exact-proposal e stato fisico finale esatto.

## Attività

- materializzare una fixture nuova e verificata per ogni tentativo;
- eseguire il reference agent mutante con la sola `workspace.patch` candidata;
- verificare read autorevole, proposta unica, preview concreta e fingerprint;
- richiedere a un operatore su TTY reale `allow once` soltanto dopo la preview;
- osservare un solo tentativo di Execute, commit atomico, nuova generazione
  indicizzata, bundle fresh e risposta finale;
- verificare digest finale esatto, assenza di altri cambiamenti, lifecycle e
  cleanup;
- ripetere fino a tre PASS consecutivi, fermandosi al primo failure;
- eseguire i terminali live pertinenti della matrice negativa che richiedono
  TTY o provider, senza contarli come sostituti dei tre successi.

## Gate di uscita

- Gate C superato `3/3` con run consecutive valide;
- ciascuna run presenta e consuma una nuova approval exact-fingerprint;
- ogni run applica una sola patch e raggiunge contesto fresh prima del final;
- il workspace converge al solo digest ammesso e non contiene temporanei;
- failure o interruzioni non vengono trasformati in retry o PASS.

## Deliverable

- evidenza live Gate C JSON e Markdown;
- registro redatto delle decisioni interattive;
- `docs/reports/milestone-11-phase-6.md`.

---

# Fase 7 — Audit, verdetto e handoff

## Obiettivo

Riconciliare prove deterministiche e live e scegliere una sola conclusione
supportata dalle evidenze.

## Attività

- verificare integrità, completezza e coerenza di tutti i report;
- classificare ogni failure e confermare l'applicazione delle stop rule;
- ripetere suite, race detector, vet, manifest validation e scansione anti-leak;
- confrontare profilo congelato, binario, hardware e comandi effettivamente
  usati;
- aggiornare roadmap, compatibility, security model, known issues e piano
  v0.2.0 senza anticipare packaging o release;
- registrare una sola decisione finale e le sue conseguenze per la Milestone
  12;
- preservare come non supportate tutte le superfici fuori dal candidato.

## Esiti ammessi

1. Controlled Mutation supportabile sul lower bound dichiarato;
2. Controlled Mutation supportabile soltanto su un requisito hardware superiore
   provato con l'intera matrice ripetuta;
3. Controlled Mutation rinviata, senza indebolire gate, limiti o threat
   boundary.

Uno skip, un risultato parziale o la sola matrice deterministica non sono un
quarto esito.

## Gate di uscita

- tutti i campioni eseguiti e non eseguiti sono contabilizzati;
- nessun failure o differenza fisica resta senza classificazione;
- support claim e requisito hardware coincidono con l'evidenza;
- report finale con verdetto GO/NO-GO verso la Milestone 12;
- nessun artifact di release viene prodotto nella Milestone 11.

## Deliverable

- `docs/reports/milestone-11-final.md`;
- eventuale ADR sul support claim o sul requisito hardware;
- handoff vincolante alla Milestone 12.

---

# Gate finale della milestone

La Milestone 11 è completata soltanto quando:

- il benchmark mutativo è versionato, deterministico e redatto;
- la matrice fisica copre tutti gli scenari congelati;
- Gate A, B e C hanno un esito fail-fast riproducibile;
- il profilo read-only v0.1.x non ha acquisito nuova autorità;
- il verdetto è uno dei tre esiti ammessi;
- la Milestone 12 riceve un support claim preciso oppure un rinvio esplicito.

Fino al superamento del gate finale, la presenza di `workspace.patch`, della
configurazione mutativa o dei report intermedi non rende Controlled Mutation
una capability supportata.
