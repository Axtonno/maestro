# Milestone 8 — Phase 2 Report

Stato: Completata

Data: 2026-08-13

---

# Risultato

La Fase 2 consegna il primo confine di prodotto eseguibile di Maestro: una
configurazione YAML strict compone target espliciti e la CLI espone `doctor`,
`models`, `agents`, `run` e `version` conservando i benchmark esistenti.

# Configurazione

- schema obbligatorio `version: 1`;
- loader separato da `runtime.Config`;
- risoluzione flag, `MAESTRO_CONFIG` e XDG senza merge;
- root relativa risolta rispetto al file;
- unknown/duplicate fields, multi-document, anchor/alias e trailing data
  rifiutati;
- provider, modello, workspace, agent, tool, policy e limiti espliciti;
- secret llama.cpp referenziato solo tramite ambiente;
- percorso v0.1 limitato a WorkspaceProvider Laravel autorevole;
- configurazione di esempio in `configs/maestro.example.yaml`.

# Composition e policy

`internal/application` costruisce provider, Runtime, plugin Laravel e policy di
prodotto. La policy valida PermissionRequest concrete contro provider, modello,
workspace e tool set configurati e non abilita process/network effect. Ogni
effect continua ad attraversare Tool Runtime e permit interno.

Il test applicativo esegue una patch reale su una fixture Laravel attraverso
reference agent, provider scripted, workspace tool, digest precondition e
refresh del Context Engine. Il terminale è completed e la generazione passa da
1 a 2 senza contesto stale.

# CLI

- root/help coerenti e senza config;
- `doctor` con check indipendenti e probe read-only;
- `models` con listing ordinato sul provider esplicito;
- `agents` senza I/O provider;
- `run` con istruzione posizionale o stdin bounded;
- `version` con build info e fallback development;
- stdout/stderr separati;
- exit code 0, 1, 2, 3, 4 e 130 mappati.

# Sicurezza e limiti

- doctor non invoca il modello e non muta workspace/cataloghi;
- nessun secret è serializzato o stampato;
- la CLI non chiama direttamente istanze tool;
- nessuna policy permissiva è implicita;
- `prompt` senza Approver termina come permission denied;
- target e hard ceiling vengono ricostruiti dai contratti pubblici;
- la root resta fuori dai messaggi del modello.

L'Approver interattivo, il rendering sintetico di piano/stato e l'UX non-TTY
completa appartengono alla Fase 3. Packaging, artifact e metadata release
appartengono alla Fase 4. Nessun gate live è dichiarato da questa fase.

# Test

La suite copre:

- parsing strict, versione, path, precedenza e secret reference;
- target impliciti/unsupported e bound;
- policy allow/prompt/deny e target mismatch;
- doctor senza model invocation;
- patch agentica con reindex;
- tutti i comandi, help, version e config invalid;
- permission/provider failure e cancellazione con exit code stabile;
- compatibilità dei comandi benchmark esistenti.

# Gate

La configurazione descrive senza ambiguità un run e la CLI di sviluppo può
diagnosticare, ispezionare ed eseguire il reference agent senza bypassare i
runtime esistenti. I criteri di uscita della Fase 2 sono soddisfatti.
