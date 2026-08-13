# v0.1.0 Release Readiness Audit

Versione: 0.1.0

Stato: GO alla Milestone 8 — release non ancora pronta

Data: 2026-08-13

---

# Scopo

Verificare il prodotto risultante dalla Milestone 7 senza riaprire le decisioni
architetturali già validate e definire il divario concreto tra la baseline
ingegneristica corrente e la prima release pubblica di Maestro.

La definizione di prodotto proposta per la release è:

> Maestro può essere installato, configurato e usato da uno sviluppatore per
> eseguire un agente locale controllato su un progetto reale.

Questo audit usa come evidenza primaria:

- `reports/milestone-7-final.md` e i sette report di fase;
- `agent-system-api-compatibility-audit.md`;
- `reports/milestone-3-live-ollama-validation.md`;
- CLI, composition root, package pubblici e test presenti nel repository;
- una nuova esecuzione di `go test ./...` il 2026-08-13.

---

# Verdetto

**GO alla Milestone 8**, con un perimetro ridefinito intorno alla v0.1.0.

La Milestone 7 supera il gate post-milestone come baseline ingegneristica. Tool
System e Agent System sono composti, bounded, default-deny e verificati nello
scenario deterministico `read -> patch -> reindex -> final`.

Il GO autorizza il percorso di productization, non la pubblicazione. La v0.1.0
non è ancora pronta. Il runtime è consumabile da codice Go e la CLI
espone i benchmark, ma manca ancora il percorso di prodotto che trasformi i
contratti esistenti in un'esperienza installabile, configurabile e riproducibile
da un utilizzatore esterno.

| Area | Stato | Evidenza o divario |
|---|---|---|
| Gate deterministico Milestone 7 | Superato | Report finale, suite completa verde e scenario autonomo senza rete |
| Scenario Laravel riproducibile | Parziale | Plugin e benchmark Laravel sono reali; lo scenario agente mutante usa una fixture generica e non è esposto dalla CLI |
| Configurazione di prodotto | Mancante | Esiste `runtime.Config` e configurazione benchmark via ambiente, non un file utente versionato e strict |
| Esperienza CLI esterna | Mancante | Sono disponibili soltanto `bench`; `--help`, `doctor`, `models`, `agents`, `run` e `version` non costituiscono ancora una superficie coerente |
| Approval concreto | Parziale | Il contratto `Approver` e i gate deterministici esistono; manca un approver terminale e il comportamento non interattivo |
| Contratti pubblici | Parziale | Gli audit di compatibilità esistono; manca una dichiarazione di stabilità e supporto per v0.1.0 |
| Sicurezza pubblicamente dichiarata | Parziale | Gli invarianti sono documentati in più documenti; manca un security model unico rivolto all'utente |
| Provider live | Parziale | Ollama è validato; la matrice llama.cpp è ancora pendente |
| Packaging e installazione | Mancante | Nessun artifact di release, procedura di installazione pulita o metadato di versione CLI |
| Licenza e documenti pubblici | Mancante | Nel repository non è presente un file di licenza né una guida di sicurezza per la release |

Il gate non richiede una modifica dell'architettura di Milestone 7. Richiede una
composition applicativa, contratti di configurazione e CLI, documentazione e
validazione live.

---

# Audit del prodotto corrente

## Scenario Laravel

Sono già presenti tre livelli utili ma distinti:

1. `pkg/plugin/laravel` carica e avvia il plugin attraverso Maestro, rileva
   `artisan` e `composer.json`, espone un `WorkspaceProvider` e indicizza il
   workspace;
2. `maestro bench laravel` materializza il dataset embedded
   `maestro-laravel-mini@1.0.0`, carica realmente il plugin e misura cinque task
   generativi e un retrieval embedding;
3. il test `TestReferenceAgentAutonomousWorkspaceScenarioAndRedactedEvents`
   esegue realmente read, patch, refresh e risposta finale, ma usa un provider
   scripted e una fixture Go minimale.

Queste evidenze provano separatamente integrazione Laravel, comportamento del
reference agent e pipeline di benchmark. Non provano ancora che un utente possa
installare il binario, puntarlo a un progetto Laravel reale, approvare una
mutazione e ottenere un run completo contro un provider locale.

Gate richiesto per la release: un'unica procedura versionata deve unire plugin
Laravel, Context Engine, reference agent, provider live, policy esplicita,
approval e verifica della modifica entro un workspace temporaneo o sacrificabile.

## Configurazione

`pkg/runtime.Config` è uno snapshot key/value destinato ai componenti. Gli
adapter Ollama e llama.cpp hanno configurazioni Go tipizzate; i benchmark
compongono il runtime da variabili d'ambiente. Nessuno di questi elementi è un
contratto di configurazione del prodotto.

Per la v0.1.0 mancano:

- formato e versione del file;
- parsing strict e diagnostiche per campo;
- precedenza tra flag, file e ambiente;
- target espliciti per provider, modello, workspace, agente e policy;
- limiti agentici configurabili e bounded;
- riferimento sicuro ai secret senza serializzarli nei report;
- comando che validi la configurazione senza eseguire effetti.

Il nuovo contratto non deve riusare direttamente `runtime.Config` come schema
pubblico: quella interfaccia resta il meccanismo generico del Runtime Core.

## CLI ed esperienza esterna

Il binario corrente supporta `maestro bench` e, senza argomenti, stampa alcune
informazioni di sistema. `maestro --help` viene classificato come comando
sconosciuto. Non esistono oggi i comandi di prodotto proposti per v0.1.0.

La superficie minima necessaria è:

```text
maestro doctor
maestro models
maestro agents
maestro run
maestro version
```

I comandi benchmark restano disponibili ma non definiscono il quick start. Il
percorso principale deve avere help coerente, exit code documentati, output
umano su stdout, diagnostiche su stderr e nessun effetto implicito durante
discovery o diagnosi.

## Approval

Milestone 7 consegna il confine corretto: action concrete, policy default-deny,
decisione `prompt`, `Approver`, grant one-shot o run-scoped e permit interno
consumabile. Il composition root non registra policy permissive.

Per l'uso esterno manca ancora:

- una policy costruita dalla configurazione di prodotto;
- una rappresentazione locale comprensibile delle action da approvare;
- un approver su terminale con scelta esplicita e cancellabile;
- deny sicuro quando stdin non è un TTY o l'approver non è disponibile;
- test che dimostrino assenza di auto-approval e di leakage di secret/contenuti.

L'approvazione CLI è un'interfaccia sul permission model trusted in-process, non
una sandbox e non una barriera contro codice host o plugin malevolo.

## Contratti pubblici candidati per v0.1.0

La release deve dichiarare sperimentali, ma intenzionali e versionati:

- CLI, nomi dei comandi, exit code e forma base dell'output;
- schema del file di configurazione `version: 1`;
- composition root `maestro.New` e relativi accessor;
- contratti sotto `pkg/provider`, `pkg/contextengine`, `pkg/tool`, `pkg/agent`,
  `pkg/plugin` e `pkg/gestor`;
- facade ufficiali Ollama, llama.cpp e Laravel;
- formati manifest/report già versionati del Benchmark Layer.

`internal/` non è API. Il formato degli eventi resta pubblico soltanto per i
topic e payload allowlist documentati. Nessuna promessa di compatibilità 1.x è
implicita: fino a v1, cambi breaking sono ammessi solo se documentati in note di
release e accompagnati da migrazione.

Prima della release serve un audit unico che registri versione corrente,
ownership, compatibility promise e deprecation policy. Gli audit di milestone
rimangono le evidenze tecniche sottostanti.

## Limiti di sicurezza da dichiarare

La documentazione pubblica della v0.1.0 deve affermare almeno che:

- runtime, agenti, tool e plugin Go sono trusted e in-process;
- non esistono sandbox, isolamento di processo o contenimento dei privilegi OS;
- la policy controlla il percorso orchestrato, non codice ostile nello stesso
  processo;
- prompt, output modello e arguments dei tool sono input non fidati;
- i workspace tool ufficiali applicano root containment, rifiuto dei symlink,
  limiti di output e digest precondition, ma il processo conserva i privilegi
  dell'utente che lo avvia;
- le chiamate di rete avvengono verso provider configurati esplicitamente;
- non esiste un secret manager; i secret non devono essere inseriti nel file e
  vanno referenziati tramite variabili d'ambiente;
- eventi e report ufficiali sono redatti, ma stdout, risposte del modello e file
  modificati sono dati visibili all'utente locale;
- non esistono recovery persistente, rollback generale o transazioni dei tool;
- non sono supportati tool shell, Git write, Docker o esecuzione remota nella
  baseline della release.

## Funzionalità non ancora validate live

Restano da validare prima della release candidate:

- matrice integration e Smoke per llama.cpp, inclusi tool calling e streaming;
- reference agent contro un provider reale;
- run agente read-only e mutante su una fixture Laravel reale;
- approval interattivo e deny non interattivo;
- caricamento del file di configurazione da un ambiente utente pulito;
- artifact installato fuori dal checkout del repository;
- comportamento di `doctor`, help, version e codici di uscita sul sistema
  dichiarato supportato.

Nel repository corrente e nella cronologia Git disponibile non è presente un
report `milestone-3-live-llamacpp-validation.md`. Le decisioni successive che
indicano la matrice llama.cpp come completata non sono quindi verificabili dalla
baseline locale: prima del gate release occorre recuperare il report oppure
rieseguire la matrice.

Il gate Ollama già completato resta valido. Il caso negativo
`qwen2.5-coder:7b` e la fixture positiva `llama3.1:8b` devono rimanere
documentati senza trasformare una capability dichiarata in supporto implicito.

---

# Decisioni per il percorso v0.1.0

- La Milestone 7 è accettata e non viene riaperta.
- La prima parte della Milestone 8 è ridefinita come productization per v0.1.0.
- La piattaforma ufficiale iniziale è Linux `amd64`; altri target restano
  sperimentali finché non superano una prova documentata.
- Ollama con `llama3.1:8b` e `embeddinggemma:latest` è il percorso locale già
  validato; `qwen2.5-coder:7b` resta il caso negativo canonico per tool calling.
- llama.cpp è candidato al supporto ufficiale e lo diventa soltanto dopo la
  presenza e la verifica della matrice live verde.
- Se la matrice llama.cpp non può essere completata, la release richiede una
  decisione esplicita di scope e la classificazione dell'adapter come
  sperimentale; non è consentito lasciare lo stato ambiguo.
- SDK stabile, plugin/tool terzi, sandbox, memoria persistente, multi-agent,
  shell/Git/Docker e selezione automatica restano fuori dalla v0.1.0.
- ADR-0026 rende vincolanti il perimetro, il supporto iniziale e i gate.

---

# Gate di ingresso alla release candidate

La release candidate può essere creata soltanto quando:

- i cinque comandi minimi sono implementati e coperti da test;
- il file di configurazione `version: 1` è documentato, parsed strict e validato;
- provider, modello, workspace, agente, policy e limiti sono espliciti;
- reference agent e approval funzionano dalla CLI;
- quick start Laravel e scenario live end-to-end sono riproducibili;
- matrice llama.cpp e debito formale della Milestone 3 sono chiusi, oppure il
  supporto llama.cpp è delimitato da una decisione documentata;
- security model, limitazioni, licenza, installazione e contratti sperimentali
  sono pubblicati;
- artifact con versione e commit è prodotto per la piattaforma supportata;
- suite, race detector, vet, benchmark deterministico e audit API sono verdi.

# Gate di pubblicazione

Dopo la release candidate sono inoltre obbligatori:

- installazione dell'artifact in un ambiente pulito;
- esecuzione del quick start senza usare file non inclusi nella release;
- verifica checksum e `maestro version`;
- nessun blocker noto di severità release e nessuna credenziale negli artifact;
- note di release con capability supportate, limiti e problemi noti.
