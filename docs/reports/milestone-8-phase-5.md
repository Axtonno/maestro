# Milestone 8 — Fase 5 Validazione live e release candidate

Data: 2026-08-15

Stato: Gate live read-only superato; produzione RC in corso

Verdetto: **GO alla promozione `v0.1.0-rc.1`, non ancora alla release finale**

---

# Contratto validato

ADR-0029 restringe la v0.1.0 al reference agent read-only su Linux `amd64`,
Ollama e `llama3.1:8b`, con `embeddinggemma:latest` come fixture embedding. Il
profilo ufficiale contiene soltanto list/read/search e nega le mutazioni.
llama.cpp, tool mutanti, approval mutativa e reference agent mutante sono
sperimentali/non supportati e non bloccano il percorso v0.1.0.

# Evidenze precedenti conservate

- Smoke matrix Ollama provider-level: 13 passed, 1 skipped, 0 failed;
- `llama3.1:8b`: provider/tool calling e reference agent read-only validati;
- matrice mutativa 8B conclusa senza vincitori;
- due tentativi llama.cpp router mode invalidati dagli OOM e non usati come
  support claim;
- nessuna mutazione non autorizzata osservata.

Il dettaglio storico è in `milestone-8-phase-5-interim.md` e
`milestone-8-model-selection.md`.

# Candidate read-only

`pc.4` ha superato il packaging ma fallito il primo quick start con
`tool_failure`; è storico e non promuovibile. Il prompt capability-aware
successivo è incorporato in `pc.5` e coperto deterministicamente sia per il
ramo read-only sia per il ramo mutativo sperimentale.

`pc.5`, commit `2732f26af4550833ad1b2d9cd4ca1caf5d72cd30`, supera:

- doppio build byte-identico e checksum
  `4eb9abdfab6efbd00dc624b509581ec57666da1c4645d60abadc9316104ffe11`;
- ispezione archive, licenze, secret/path scan e configurazione read-only;
- installazione in directory pulita senza checkout;
- version, help, doctor 9/9, models e agents dall'artifact;
- due quick start read-only consecutivi con exit code 0, una read reale e
  risposta corretta;
- digest della fixture invariato;
- suite ripetuta tre volte, race detector, vet e diff check.

Il report dedicato è `milestone-8-clean-installation.md`.

# Sicurezza e supporto

Il gate non trasforma permission in sandbox e non amplia l'autorità del
runtime. La configurazione inclusa non rende disponibili tool mutanti e nega
comunque `workspace.mutate`. Le capacità generiche restano trusted in-process e
sperimentali. llama.cpp resta adapter sperimentale; la Milestone 3 non viene
chiusa retroattivamente.

# Promozione

`pc.5` non viene rinominato. Un nuovo archive `v0.1.0-rc.1` deve essere prodotto
dallo stesso contenuto validato dopo il presente gate, con versione, commit,
manifest `release-candidate` e checksum propri. Il candidate RC deve ripetere
almeno checksum, version/help, profilo read-only, doctor e un run di conferma
dall'archive estratto.

La Fase 5 sarà formalmente completata quando identità e checksum di `rc.1`
saranno registrati. La Fase 6 può iniziare subito dopo; tag e artifact finali
`v0.1.0` restano fuori da questo verdetto.
