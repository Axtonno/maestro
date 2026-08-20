# Milestone 10 — Controlled Mutation Development Plan

Versione: 0.1.0

Stato: In corso — Fasi 1–2 completate, Fase 3 pronta

Data: 2026-08-20

Documenti di riferimento:

- `roadmap.md`;
- `v0.2.0-development-plan.md`;
- `reports/milestone-9-final.md`;
- `adr/ADR-0025.md`;
- `adr/ADR-0028.md`;
- `adr/ADR-0029.md`;
- `adr/ADR-0030.md`;
- `adr/ADR-0031.md`.

---

# Obiettivo operativo

Consegnare una sola vertical slice mutativa controllata per il reference agent
Laravel:

```text
read -> prepare patch -> preview -> approval -> apply -> reindex -> final
```

Il percorso candidato modifica al massimo un file PHP esistente sotto `app/`
tramite `workspace.patch`. La proposta mostrata all'utente, il fingerprint
autorizzato e l'effetto applicato devono derivare dallo stesso oggetto
immutabile. Il profilo v0.1.x read-only resta invariato.

La milestone implementa e verifica il contratto in modo deterministico. La
qualificazione live di provider, modello e hardware appartiene alla Milestone
11 e la productization della v0.2.0 alla Milestone 12.

---

# Sequenza delle fasi

| Fase | Titolo | Stato corrente | Dipende da |
|---|---|---|---|
| 1 | Contratto di release e threat boundary | Completata | Milestone 9 |
| 2 | Proposta patch autorevole e preview sicura | Completata | Fase 1 |
| 3 | Approval exact-fingerprint e opt-in di prodotto | Pronta | Fase 2 |
| 4 | Commit atomico e fault injection | Pianificata | Fase 3 |
| 5 | Freshness, reindex e terminali applicativi | Pianificata | Fase 4 |
| 6 | Integrazione, audit e chiusura | Pianificata | Fasi 1–5 |

Le fasi sono sequenziali rispetto ai gate. Ogni fase produce un report sotto
`docs/reports/`; una fase successiva non può trasformare un'invariante non
dimostrata in un'assunzione. Le prove live non fanno parte del gate di questa
milestone.

---

# Regole trasversali

- `workspace.patch` è la sola operazione mutativa candidata al supporto.
- Ogni run può tentare una sola mutazione; un nuovo tentativo richiede una
  nuova run, una nuova read e una nuova approval.
- Nessuna policy mutativa `allow`, modalità non interattiva o flag equivalente
  a `--yes` appartiene al percorso supportato.
- Il fingerprint lega almeno policy, tool ID e versione, run, call, arguments
  normalizzati e action.
- L'inizio di `Execute` rende il contesto stale anche in caso di failure.
- Il successo finale richiede commit riuscito, nuova generazione indicizzata e
  bundle fresh.
- Test, eventi e report non conservano root fisiche, secret, prompt, contenuti
  completi o payload del modello non necessari.
- `workspace.write`, shell, Git, processi, recovery, sandbox, multi-agent e
  modifiche multi-file restano fuori scope.

---

# Fase 1 — Contratto di release e threat boundary

## Obiettivo

Congelare autorità, limiti, stati e superfici candidate prima di modificare il
percorso mutativo.

## Attività

- definire l'unità atomica supportata e le precondizioni fisiche;
- fissare path, tipo e dimensione dei file candidati;
- definire la semantica di preview, allow once, deny e input non interattivo;
- distinguere failure prima e dopo il commit;
- fissare la relazione tra apply, context stale, reindex e risposta finale;
- identificare il profilo provider/modello/hardware candidato senza
  dichiararlo supportato;
- classificare stabilità di CLI, configurazione e API;
- registrare esclusioni e alternative in un ADR.

## Gate di uscita

- ADR accettato e indicizzato;
- matrice degli stati prima/dopo commit non ambigua;
- profilo read-only confermato invariato;
- gap della baseline associati alle fasi che li chiudono;
- suite repository-wide verde prima dell'implementazione.

## Deliverable

- `docs/adr/ADR-0031.md`;
- `docs/reports/milestone-10-phase-1.md`;
- questo piano operativo.

---

# Fase 2 — Proposta patch autorevole e preview sicura

## Obiettivo

Introdurre un oggetto immutabile che sia l'unica origine della preview,
dell'approval e dell'esecuzione.

## Attività

- rappresentare path logico, digest atteso, sostituzione esatta, risultato
  proposto, tool/version/run/call, action e fingerprint;
- costruire la proposta soltanto dopo una read autorevole, non troncata e
  coerente;
- generare una diff deterministica e bounded dalla proposta;
- mostrare tool, intenzione sintetica, path, digest, diff e precondizioni;
- impedire root fisiche, contenuti estranei, prompt e secret nella preview;
- rifiutare file fuori da `app/**/*.php`, symlink, file non regolari, UTF-8
  invalido, NUL, digest invalido, occorrenze zero/multiple e dimensioni oltre
  2 MiB;
- coprire immutabilità, defensive copy, determinismo e redazione.

## Gate di uscita

- preview e Execute consumano la stessa proposta validata;
- qualsiasi variazione della proposta cambia o invalida il fingerprint;
- snapshot test della diff e test anti-leak verdi;
- nessuna approval o scrittura avviene durante la preparazione.

## Deliverable

- contratti e implementazione della proposta patch;
- renderer di preview bounded;
- `docs/reports/milestone-10-phase-2.md`.

---

# Fase 3 — Approval exact-fingerprint e opt-in di prodotto

## Obiettivo

Fare in modo che l'utente possa autorizzare una sola esecuzione della proposta
concreta vista a terminale.

## Attività

- integrare la preview nel prompt di approval;
- limitare le mutazioni supportate ad `allow once`;
- rendere terminali e fail-closed deny, EOF, no-TTY e input invalido;
- verificare replay, cambio di run/call/tool/version/path/digest/diff/action e
  sostituzione della proposta;
- impedire che grant read-only o run-scoped autorizzino una patch differente;
- introdurre una configurazione mutativa separata e opt-in con
  `workspace.patch` e `workspace_mutate: prompt`;
- mantenere byte-invariata la configurazione read-only distribuita.

## Gate di uscita

- nessun percorso mutativo non interattivo;
- una approval autorizza al massimo un tentativo della proposta mostrata;
- replay e proposta modificata sono rifiutati;
- profili read-only e mutativo sono separati e diagnosticabili.

## Deliverable

- approval terminale exact-proposal;
- profilo mutativo candidato separato;
- `docs/reports/milestone-10-phase-3.md`.

---

# Fase 4 — Commit atomico e fault injection

## Obiettivo

Sostituire la riscrittura in-place con un commit filesystem atomico e
classificare con precisione ogni failure.

## Attività

- rivalidare path, tipo, symlink, digest, occorrenza e dimensione subito prima
  del commit;
- preparare un file temporaneo non seguibile nella stessa directory;
- preservare i permessi supportati, sincronizzare contenuto, rinominare
  atomicamente e sincronizzare la directory;
- pulire il temporaneo su deny, cancellazione e failure pre-commit;
- distinguere failure pre-commit da commit riuscito con durability o cleanup
  successivi incerti;
- introdurre seam di fault injection per open, read, write, sync, rename,
  directory sync e cleanup;
- dimostrare assenza di file parziali e retry impliciti.

## Gate di uscita

- i lettori osservano il vecchio o il nuovo file, mai contenuto parziale;
- stale digest, traversal, symlink e sostituzioni concorrenti sono rifiutati;
- ogni fault pre-commit lascia il target byte-identico;
- ogni temporaneo viene rimosso quando la rimozione è ancora possibile;
- il punto di commit è unico e testato.

## Deliverable

- applicatore atomico di patch;
- suite di fault injection;
- `docs/reports/milestone-10-phase-4.md`.

---

# Fase 5 — Freshness, reindex e terminali applicativi

## Obiettivo

Rendere osservabile l'intera sequenza e impedire un successo finale su contesto
stale o stato fisico ambiguo.

## Attività

- esporre in ordine proposta, decisione, apply e reindex con payload redatti;
- marcare stale all'inizio dell'effetto e conservare lo stato dopo ogni esito;
- richiedere generazione indicizzata nuova e bundle corrispondente prima del
  ritorno al modello;
- distinguere apply fallito, apply riuscito con refresh fallito e completato;
- rappresentare SIGINT/deadline prima e dopo il punto di commit;
- impedire testo finale del modello quando apply o refresh non sono riusciti;
- verificare che nessun terminale descriva il workspace come invariato dopo un
  commit riuscito.

## Gate di uscita

- successo finale possibile soltanto su bundle fresh;
- failure di refresh conserva `ContextStale=true` e comunica apply riuscito;
- cancellazione/deadline riflettono il lato corretto del punto di commit;
- eventi e CLI superano test di ordine e anti-leak.

## Deliverable

- lifecycle mutativo completo e osservabile;
- terminali e rendering applicativo aggiornati;
- `docs/reports/milestone-10-phase-5.md`.

---

# Fase 6 — Integrazione, audit e chiusura

## Obiettivo

Dimostrare deterministicamente il vertical slice completo e consegnare un
contratto qualificabile alla Milestone 11.

## Attività

- eseguire la matrice deterministica prepare/preview/approval/apply/reindex;
- coprire positive path, deny, EOF, no-TTY, input invalido, stale digest,
  traversal, symlink, replay, proposta cambiata, cancellazione e fault;
- verificare il profilo read-only e la configurazione v0.1.x senza regressioni;
- aggiornare sicurezza, configurazione, CLI, workspace e known issues;
- eseguire suite, race detector, vet e diff check;
- produrre l'audit finale e il verdetto GO/NO-GO verso la Milestone 11.

## Gate di uscita

- tutti i gate della milestone verdi;
- nessun percorso mutativo non interattivo o non approvato;
- nessun contenuto parziale o successo su contesto stale;
- documentazione coerente con codice e test;
- profilo benchmark mutativo consegnato senza anticipare la qualificazione
  live.

## Deliverable

- `docs/reports/milestone-10-phase-6.md`;
- `docs/reports/milestone-10-final.md`;
- documentazione pubblica allineata.

---

# Gate repository-wide

Le fasi che modificano codice eseguono almeno:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

Le fasi documentali eseguono almeno `go test ./...` e `git diff --check`. Le
prove live della Milestone 11 sono aggiuntive e non sostituibili dalla suite.

---

# Definizione di completamento

La Milestone 10 è completata soltanto quando la stessa proposta concreta
produce preview, approval ed effetto, il commit è atomico, il contesto viene
ricostruito prima del successo finale e tutti gli esiti prima/dopo commit sono
osservabili senza leakage.

La sola presenza di `workspace.patch`, di una diff a terminale o di test del
Tool Runtime non costituisce completamento né supporto mutativo.
