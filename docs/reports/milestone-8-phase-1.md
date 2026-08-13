# Milestone 8 — Fase 1 Release Contract and Audit Report

Data: 2026-08-13

Stato: Completata

---

# Esito

Il gate della Fase 1 è superato. La Milestone 7 è accettata come baseline
ingegneristica e la Milestone 8 riceve **GO** con un perimetro ridefinito
intorno alla productization della v0.1.0.

Il verdetto non autorizza ancora la release. La correttezza deterministica del
runtime è provata, ma installazione, configurazione pubblica, esperienza CLI,
approval terminale, artifact e scenario live completo devono essere
consegnati e verificati nelle fasi successive.

# Evidenze esaminate

L'audit ha usato come evidenza primaria:

- `docs/reports/milestone-7-final.md` e i report delle sette fasi;
- `docs/agent-system-api-compatibility-audit.md`;
- `docs/reports/milestone-3-live-ollama-validation.md`;
- composition root, CLI, package pubblici e test presenti nel repository;
- suite `go test ./...` rieseguita durante l'audit del 2026-08-13.

La Milestone 7 certifica lo scenario deterministico
`read -> patch -> reindex -> final`, con Tool e Agent Runtime bounded,
workspace-aware e default-deny. Non certifica ancora il percorso di un nuovo
utilizzatore a partire da un artifact installato.

# Contratto di prodotto

La v0.1.0 è definita dal seguente risultato osservabile:

> Maestro può essere installato, configurato e usato da uno sviluppatore per
> eseguire un agente locale controllato su un progetto reale.

Il percorso minimo approvato è:

```text
install -> configure -> doctor -> inspect -> run -> approve -> verify
```

La release deve quindi includere:

- CLI minima `doctor`, `models`, `agents`, `run` e `version`;
- configurazione YAML strict e versionata;
- target espliciti per provider, modello, workspace, agente, policy, tool e
  limiti;
- reference agent e workspace Laravel;
- approval concreta e cancellabile;
- artifact versionato, checksum e installazione pulita;
- quick start, security model, compatibility matrix e limitazioni;
- almeno uno scenario agentico live end-to-end.

# Decisioni registrate

ADR-0026 rende vincolanti le seguenti decisioni:

- piattaforma ufficiale iniziale: Linux `amd64`;
- percorso provider già validato: Ollama;
- fixture positiva: `llama3.1:8b` per chat/tool calling e
  `embeddinggemma:latest` per embedding;
- caso negativo canonico: `qwen2.5-coder:7b` per tool calling;
- llama.cpp candidato al supporto ufficiale, subordinato a una matrice live
  documentata;
- configurazione pubblica iniziale: singolo documento YAML strict con
  `version: 1`;
- CLI, schema config, composition root e package Go pubblici intenzionali ma
  sperimentali nella serie 0.x;
- runtime, agenti, tool e plugin built-in trusted in-process;
- permission e approval non equivalgono a sandbox o isolamento OS;
- cambi breaking 0.x documentati con note di release e migrazione; formati
  serializzati modificati soltanto tramite versione schema.

# Supporto e sicurezza

Il release contract richiede di dichiarare che:

- Maestro opera con i privilegi dell'utente locale;
- prompt, output modello e arguments dei tool sono input non fidati;
- i workspace tool applicano containment, rifiuto dei symlink, limiti e digest
  precondition, ma non isolano il processo;
- le chiamate di rete sono rivolte soltanto al provider configurato;
- i secret sono referenziati tramite ambiente e non inseriti nel file;
- eventi e report ufficiali sono redatti, mentre risultato finale e file
  modificati sono visibili localmente;
- non esistono rollback generale, recovery persistente o transazioni dei tool.

# Ambiti esclusi dalla v0.1.0

Sono rinviati oltre la release:

- sandbox e isolamento di processo;
- memoria persistente e recovery;
- multi-agent e remote execution;
- shell, Git e Docker completi;
- secret manager;
- packaging generalizzato di agent e tool di terze parti;
- SDK stabile;
- selezione automatica di provider o modello.

Queste esclusioni impediscono che la Milestone 8 torni a essere un contenitore
indefinito denominato “Ecosistema”.

# Evidenza llama.cpp

Nel repository e nella cronologia Git disponibili durante l'audit non è
presente `docs/reports/milestone-3-live-llamacpp-validation.md`. Le indicazioni
esterne di completamento non sono quindi certificabili dalla baseline locale.

Prima del gate della release candidate occorre:

1. recuperare e verificare il report live; oppure
2. rieseguire la matrice llama.cpp; oppure
3. adottare una decisione esplicita che classifichi llama.cpp come
   sperimentale nella v0.1.0.

Uno stato ambiguo non soddisfa il release contract.

# Deliverable

La Fase 1 ha prodotto e approvato:

- `docs/release-readiness-audit.md`;
- `docs/milestone-8-design.md`;
- `docs/milestone-8-development-plan.md`;
- `docs/adr/ADR-0026.md`;
- riallineamento di `README.md`, `docs/roadmap.md` e `MAESTRO_CONTEXT.md`.

# Gate di uscita

| Criterio | Esito |
|---|---|
| GO alla Milestone 8 esplicito | Superato |
| Product promise della v0.1.0 definita | Superato |
| Piattaforma e support matrix iniziali definiti | Superato |
| Contratti pubblici sperimentali dichiarati | Superato |
| Security boundary trusted in-process dichiarato | Superato |
| Funzionalità escluse delimitate | Superato |
| Gate di release candidate e pubblicazione descritti | Superato |
| Debito llama.cpp identificato senza assunzioni | Superato |
| GO immediato alla pubblicazione v0.1.0 | Non concesso |

# Rischi e lavoro residuo

Al termine della Fase 1 restano aperti:

- configurazione e CLI di prodotto;
- esperienza operativa e approval terminale;
- packaging, artifact e installazione pulita;
- scenario live reference agent su Laravel;
- chiusura o delimitazione della matrice llama.cpp;
- documentazione pubblica, licenza, security model e release artifacts.

Questi elementi sono assegnati in modo verificabile alle Fasi 2–6 e non
richiedono una nuova revisione architetturale generale.

# Verdetto

**GO alla Fase 2 — Configurazione e CLI minima.**

La v0.1.0 resta **non pronta** fino al completamento dei gate successivi.
