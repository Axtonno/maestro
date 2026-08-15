# Milestone 8 — Fase 5 Validazione live e release candidate

Data: 2026-08-15

Stato: Completata

Verdetto: **GO alla Fase 6; `v0.1.0-rc.2` è la release candidate validata, non
ancora la release finale.**

---

# Contratto validato

ADR-0029 restringe la v0.1.0 al reference agent read-only su Linux `amd64`,
Ollama e `llama3.1:8b`, con `embeddinggemma:latest` come fixture embedding. Il
profilo ufficiale contiene soltanto `workspace.list`, `workspace.read` e
`workspace.search` e imposta `workspace_mutate: deny`. llama.cpp, tool mutanti,
approval mutativa e reference agent mutante sono sperimentali/non supportati e
non bloccano questo percorso.

# Evidenze conservate

- Smoke matrix Ollama provider-level: 13 passed, 1 skipped, 0 failed;
- `llama3.1:8b`: provider/tool calling e reference agent read-only validati;
- matrice mutativa 8B conclusa senza vincitori e senza modifiche ai criteri;
- due tentativi llama.cpp router mode invalidati dagli OOM e non usati come
  support claim;
- nessuna mutazione non autorizzata osservata.

Il dettaglio storico è in `milestone-8-phase-5-interim.md` e
`milestone-8-model-selection.md`.

# Evoluzione dei candidate

`pc.4` ha superato il packaging ma fallito il primo quick start con
`tool_failure`; è storico e non promuovibile. Il prompt capability-aware
successivo è incorporato in `pc.5`, che ha superato installazione pulita e due
quick start consecutivi.

Il distinto `v0.1.0-rc.1`, commit
`05dcd46eb39296ded0572de013af6c421a45b5b8`, è stato prodotto in modo
riproducibile con SHA-256
`77a6f6236a8de02a5f54d75e58ec142a57da6cc0e953f3831401d07a7376f515`
e dimensione 3598575 byte. Checksum, manifest `release-candidate`, version,
help, configurazione, fixture, doctor 9/9, models e agents erano verdi da una
directory contenente inizialmente soltanto archive e checksum. Il run di
conferma ha però terminato `tool_failure` dopo 164093 ms, un turno e una tool
call, con 1340/37 token e exit code 1. Il controller è rimasto invariato.
`rc.1` è quindi immutabile ma non promuovibile.

Senza cambiare modello, temperatura, timeout, task o criteri, il protocollo
read-only è stato irrobustito affinché il modello usi un nome funzione esatto,
soltanto campi dello schema e path logici relativi senza root fisica, slash
iniziale, URI o parent traversal. Il comportamento è coperto da test.

`v0.1.0-pc.6`, commit `ab109a5f878b8e1f10d69327736f014ad916a970`,
ha superato doppio packaging byte-identico con SHA-256
`a148df8ff46d412ba85a39429f02048911d0793d3494db031a79cfa8ea76207b`
e dimensione 3598595 byte, installazione pulita, CLI completa e due quick
start consecutivi:

| Run | Terminale | Turni | Tool call | Token in/out | Durata | Risposta |
|---:|---|---:|---:|---:|---:|---|
| 1 | `completed` | 2 | 1 | 2886 / 87 | 320075 ms | `OrderService::create` |
| 2 | `completed` | 2 | 1 | 2887 / 93 | 66128 ms | `OrderService::create` |

Entrambe le call sono letture reali, gli exit code sono 0 e il digest del
controller resta
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`.

# Release candidate validata

Il nuovo artifact non rinomina alcun candidate precedente:

| Campo | Valore |
|---|---|
| Artifact | `maestro-v0.1.0-rc.2-linux-amd64.tar.gz` |
| Versione | `v0.1.0-rc.2` |
| Commit | `ab109a5f878b8e1f10d69327736f014ad916a970` |
| SHA-256 | `442090c6e2dac6095aa4532d658def42cd39e04a34baff401b3a92aec1fd9105` |
| Dimensione | 3598576 byte |
| Manifest | `release-candidate` |

Due build indipendenti dello stesso commit sono byte-identiche. Da una nuova
directory pulita, usando soltanto archive e checksum, sono verdi:

- `sha256sum -c`, estrazione, `maestro version` e root help;
- corrispondenza fra nome, versione, commit e manifest;
- licenze Apache-2.0, documentazione, configurazione e fixture incluse;
- assenza di tool mutanti e `workspace_mutate: deny` nel profilo ufficiale;
- `doctor` 9/9, `models` e `agents` contro Ollama;
- run di conferma `completed` con exit code 0, due turni, una read reale,
  2888/94 token e durata 64296 ms;
- risposta corretta `OrderService::create` e digest della fixture invariato.

# Gate deterministici finali

Sul commit dell'RC sono verdi:

- `go test -count=3 ./...`;
- `go test -race ./...`;
- `go vet ./...`;
- syntax check degli script e `git diff --check`.

Il modello è stato arrestato dopo la prova e `ollama ps` non riportava modelli
residenti.

# Sicurezza e supporto

Il gate non trasforma permission in sandbox e non amplia l'autorità del
runtime. La configurazione inclusa non rende disponibili tool mutanti e nega
comunque `workspace_mutate`. Le capacità generiche restano trusted in-process e
sperimentali. llama.cpp resta adapter sperimentale; la Milestone 3 non viene
chiusa retroattivamente.

# Verdetto

**Fase 5 completata. GO alla Fase 6, non ancora alla pubblicazione v0.1.0.**

`v0.1.0-rc.2` è l'unica release candidate promuovibile. La Fase 6 deve
completare documentazione pubblica, compatibility matrix, security model,
changelog e gate finale prima di produrre tag e artifact `v0.1.0`.
