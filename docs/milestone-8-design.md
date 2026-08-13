# Milestone 8 — v0.1.0 Productization Design

Versione: 0.1.0

Stato: Approvato — ADR-0026 Accepted

Data: 2026-08-13

Documenti di ingresso: `release-readiness-audit.md` e `adr/ADR-0026.md`.

---

# Obiettivo

Trasformare la baseline ingegneristica completata con la Milestone 7 in un
prodotto locale installabile e utilizzabile:

> Maestro può essere installato, configurato e usato da uno sviluppatore per
> eseguire un agente locale controllato su un progetto reale.

La milestone non costruisce un ecosistema generalizzato. Consegna il percorso
ufficiale minimo verso v0.1.0 e rende verificabili le scelte che il Runtime già
richiede esplicitamente.

---

# Scope

## Incluso in Milestone 8 / v0.1.0

- CLI minima: `doctor`, `models`, `agents`, `run`, `version`;
- file di configurazione YAML versionato e strict;
- selezione esplicita di provider, modello, workspace, agente e policy;
- composizione ufficiale di Ollama e plugin Laravel, più il percorso candidato
  llama.cpp soggetto al gate live;
- policy di prodotto e approval terminale;
- esecuzione del reference agent con hard limits;
- packaging del binario, checksum e istruzioni di installazione;
- quick start Laravel riproducibile;
- security model, limitazioni, licenza e compatibility statement;
- matrice live llama.cpp e chiusura del debito della Milestone 3;
- almeno uno scenario live end-to-end e una prova da ambiente pulito.

## Esplicitamente escluso

| Ambito | Destinazione |
|---|---|
| SDK stabile con promessa 1.x | Dopo v0.1.0 |
| Packaging di plugin e tool di terze parti | Dopo v0.1.0 |
| Sandbox e isolamento di processo | Milestone successiva |
| Memoria persistente e recovery | Milestone successiva |
| Multi-agent e delega | Milestone successiva |
| Shell, Git write e Docker | Milestone successiva |
| Selezione automatica di provider o modello | Evoluzione successiva |
| Remote execution e durable runs | Evoluzione successiva |
| Secret manager | Evoluzione successiva |

I comandi `maestro bench` già esistenti restano supportati, ma non sono la UX
principale della release.

---

# Matrice di supporto iniziale

| Area | v0.1.0 | Condizione |
|---|---|---|
| Piattaforma | Linux `amd64` | Artifact e installazione pulita obbligatori |
| Ollama | Supportato | Endpoint locale esplicito e matrice live verde |
| Chat/tool model | `llama3.1:8b` | Fixture positiva canonica Ollama |
| Embedding model | `embeddinggemma:latest` | Fixture positiva canonica Ollama |
| `qwen2.5-coder:7b` | Caso negativo | Non supportato per il reference agent con tool calling |
| llama.cpp | Candidato | Supportato solo dopo acquisizione o riesecuzione del report live |
| Laravel | Percorso reference | Progetto/fixture locale, plugin built-in |
| Altre piattaforme/provider/modelli | Sperimentali | Nessuna promessa finché non validati |

La matrice pubblicata con la release deve riportare anche versioni effettive di
server, modello e artifact osservate nei gate. La presenza di un adapter o di
una capability dichiarata non equivale da sola a supporto operativo.

---

# Principi di prodotto

## Esplicito prima che automatico

Provider, modello chat, workspace, agente, policy e limiti devono provenire da
configurazione o flag visibili. L'assenza di un valore richiesto è un errore;
non abilita discovery o fallback impliciti.

## Diagnosi senza effetti

Help, `version`, `agents` e la parte locale di `doctor` non invocano modelli né
modificano workspace o cataloghi remoti. I probe di rete sono dichiarati e
read-only. `models` può interrogare soltanto il provider selezionato.

## Default-deny conservato

La CLI non aggira Tool Runtime. Ogni effetto continua ad attraversare Prepare,
policy, eventuale Approver e permit interno. L'assenza di policy o approver
nega l'azione quando richiesto.

## Un percorso ufficiale piccolo

La v0.1.0 ottimizza un solo percorso completo: provider locale esplicito,
reference agent, workspace locale, plugin Laravel opzionale ma ufficiale e
workspace tool built-in. Le estensioni generiche restano contratti Go
sperimentali, non un marketplace.

## Offline salvo target configurati

Maestro non scarica modelli, plugin o aggiornamenti durante `run` o `doctor`.
Le uniche connessioni del percorso principale sono verso gli endpoint provider
esplicitamente configurati.

---

# Architettura applicativa

La CLI aggiunge un layer applicativo sopra i servizi esistenti:

```text
arguments + environment + config.yaml
                  |
                  v
       strict product configuration
                  |
                  v
 provider + plugin + policy composition
                  |
                  v
 Runtime / Gestor / Context / Tool / Agent
                  |
                  v
       terminal output and exit status
```

Il layer applicativo possiede parsing, validazione cross-field, composizione,
prompt di approval e rendering. Non sposta nel Runtime Core responsabilità di
CLI e non esporta implementazioni sotto `internal/`.

Per mantenere testabili i comandi, parsing e I/O terminale restano separati
dalla costruzione del runtime. Ogni comando riceve stream e dipendenze
iniettabili; i test non dipendono dal terminale o da processi provider reali.

---

# Contratto di configurazione

## Formato e risoluzione

Il formato iniziale è YAML con campo obbligatorio `version: 1`. Il decoder
rifiuta campi sconosciuti, documenti multipli, alias/cicli, valori duplicati e
trailing data. Le durate usano la sintassi Go documentata; path e URL vengono
validati prima della composizione.

Risoluzione del file, in ordine:

1. `--config <path>`;
2. `MAESTRO_CONFIG`;
3. `$XDG_CONFIG_HOME/maestro/config.yaml`;
4. `$HOME/.config/maestro/config.yaml` come fallback XDG.

Non vengono uniti più file. Un path esplicito mancante è un errore. Help e
`version` non richiedono il file.

Nella Fase 2 `--config` seleziona il documento ma i flag non sovrascrivono
singoli target: provider, modelli, workspace, agent e policy provengono da un
solo snapshot validato. L'ambiente fornisce esclusivamente path del documento e
secret referenziati, non valori impliciti dei target.

## Schema iniziale

Esempio illustrativo del contratto previsto:

```yaml
version: 1

provider:
  id: ollama
  base_url: http://127.0.0.1:11434
  timeout: 2m
  api_key_env: ""

models:
  chat: llama3.1:8b
  embedding: embeddinggemma:latest

workspace:
  id: laravel
  root: /absolute/path/to/project
  framework: laravel

agent:
  id: agent.reference
  streaming: true
  tools:
    - workspace.list
    - workspace.read
    - workspace.search
    - workspace.patch

policy:
  id: policy.local-review
  model: allow
  workspace_inspect: allow
  workspace_mutate: prompt

limits:
  duration: 5m
  model_turns: 12
  tool_calls: 24
  tool_calls_per_turn: 4
  plan_steps: 8
  plan_revisions: 2
  tool_result_bytes: 65536
  session_bytes: 1048576
  input_tokens: 32768
  output_tokens: 8192

context:
  retrieval: lexical
  top_k: 12
  max_tokens: 8192
  reserved_tokens: 1024
  safety_tokens: 512
```

I valori mostrati sono un esempio e non diventano default impliciti. Il file
deve nominare tutti i target necessari a `run`.

`api_key_env` contiene il nome di una variabile d'ambiente, non il secret. La
v0.1.0 non esegue interpolazione arbitraria `${...}` e non accetta credenziali
da flag, per ridurre esposizione nella shell history e nei process listing.

## Validazione

La validazione avviene su tre livelli:

- sintassi e schema strict;
- coerenza locale di ID, URL, path, durata, cardinalità e budget;
- preflight operativo di provider, modello, workspace, plugin, agent, tool e
  policy.

Un errore identifica il campo e usa un reason code stabile quando il comando
produce output machine-readable. Nessun errore include API key, prompt o
contenuto del workspace.

---

# Policy di prodotto e approval

La configurazione di prodotto esprime intenti bounded (`allow`, `deny` o
`prompt`) per quattro classi iniziali:

- invocazione del modello selezionato;
- disclosure del bundle del workspace selezionato;
- ispezione tramite workspace tool ufficiali;
- mutazione tramite workspace tool ufficiali.

Il compiler della policy valuta sempre la `PermissionRequest` concreta e
verifica provider, modello, workspace, tool set ed effect. La sintassi compatta
non introduce wildcard nei contratti core e non concede action di process o
network access ai tool.

In modalità interattiva, una decisione `prompt` mostra subject, effect,
workspace, tool e resource necessari alla decisione locale. Le scelte iniziali
sono deny, allow one-shot e allow per il run quando il contratto lo consente.
EOF, input invalido o terminale non interattivo equivalgono a deny. Timeout e
cancellazione non concedono authority e preservano il relativo terminale del
run.

Non esiste un flag globale `--yes` nella baseline. L'automazione non interattiva
deve usare una policy esplicitamente `allow` per le sole classi necessarie.

---

# Contratto CLI

## `maestro doctor`

Carica e valida la configurazione, controlla versione schema, workspace,
provider raggiungibile, capability richieste, modello chat, plugin Laravel,
agent, tool e policy. Non invoca il modello e non muta cataloghi o workspace.

Il risultato distingue `pass`, `warn`, `fail` e `skip`. I check locali possono
essere eseguiti anche se il provider non è raggiungibile, così una failure non
nasconde le successive. Secret, contenuti e path non necessari sono redatti.

## `maestro models`

Elenca i modelli osservabili del provider esplicitamente selezionato e le
capability disponibili. Non sceglie un modello e non esegue pull, load o
remove. Un endpoint non configurato è un errore, non uno skip silenzioso.

## `maestro agents`

Elenca gli agenti registrati e le capability da Gestor/Agent Runtime. La
v0.1.0 deve includere `agent.reference`. Il comando non esegue agenti o
provider call.

## `maestro run`

Riceve l'istruzione come argomento o da stdin, compone il runtime, carica il
plugin richiesto, rileva e indicizza il workspace, registra la policy, costruisce
una `RunRequest` con target e limiti espliciti ed esegue il reference agent.

La root assoluta non viene inserita nel prompt. Output finale e stato terminale
sono distinti dalle diagnostiche. SIGINT cancella il run e attiva lo shutdown
bounded; non implica rollback di effetti già completati.

## `maestro version`

Stampa versione semantica, commit e stato dirty quando disponibili. Gli
artifact di release devono restituire esattamente `v0.1.0` e il commit usato per
la build. Il formato testuale è stabile per la serie 0.1; un'opzione JSON può
essere aggiunta solo se coperta da schema e test.

## Help ed exit code

`maestro`, `maestro help`, `maestro --help` e `maestro <command> --help` devono
essere coerenti e privi di side effect.

| Codice | Significato |
|---:|---|
| 0 | Operazione completata |
| 1 | Failure operativa o run non completed |
| 2 | Uso CLI o configurazione non valida |
| 3 | Permission negata o approval non disponibile |
| 4 | Provider/modello/capability non disponibile |
| 130 | Cancellazione tramite interrupt |

La tassonomia interna più ricca resta disponibile nei messaggi e nei reason
code; la CLI non espone direttamente tutti i sentinel Go.

---

# Percorso ufficiale Laravel

Il quick start usa un progetto Laravel già esistente o la fixture versionata
inclusa nel repository di release. Il percorso è:

1. installare il binario e verificare `maestro version`;
2. avviare e configurare Ollama o llama.cpp con un modello validato;
3. creare il file `config.yaml` con target e policy espliciti;
4. eseguire `maestro doctor`;
5. controllare `maestro models` e `maestro agents`;
6. eseguire prima un task read-only;
7. eseguire un task di patch con approval one-shot;
8. verificare file modificato, reindex e terminale completed.

La prova di release deve usare una copia temporanea del progetto. Maestro non
promette rollback generale e il quick start deve suggerire un workspace sotto
version control senza eseguire comandi Git.

---

# Packaging e supporto

Il release artifact obbligatorio è un singolo binario Linux `amd64` con versione
incorporata, accompagnato da checksum SHA-256, licenza, README/quick start,
security model, compatibility statement e note di release. Ulteriori target
sono dichiarati sperimentali finché non vengono costruiti e sottoposti almeno a
smoke di avvio, help, version e doctor locale; diventano supportati soltanto
dopo una prova documentata sulla piattaforma.

`go install ...@v0.1.0` può essere un percorso aggiuntivo, non sostituisce la
verifica dell'artifact pubblicato. La pipeline di release deve essere
riproducibile e non dipendere da file non tracciati.

---

# Compatibilità

La v0.1.0 dichiara sperimentali CLI, schema config e package Go pubblici. I
contratti sono intenzionali, versionati dove serializzati e soggetti a audit;
non sono ancora stabili come una API 1.x.

Durante la serie 0.x:

- cambi breaking richiedono nota di release e istruzione di migrazione;
- campi serializzati non cambiano significato senza una nuova versione schema;
- campi sconosciuti continuano a fallire per evitare configurazioni ignorate;
- `internal/` e dettagli di rendering umano non costituiscono API;
- un SDK stabile resta fuori scope.

---

# Sicurezza

Il security model della release deve consolidare i limiti elencati nell'audit.
In particolare, Maestro v0.1.0 è un runtime trusted in-process eseguito con i
privilegi dell'utente locale. Permission e approval governano il percorso
orchestrato ma non equivalgono a sandbox. La release non deve descrivere plugin
o tool terzi come sicuri né suggerire l'esecuzione su workspace non fidati senza
una valutazione dell'utente.

---

# Criteri di successo

Il design è soddisfatto quando un nuovo utilizzatore, partendo dall'artifact e
dalla documentazione pubblicata, può:

- identificare versione e piattaforma supportata;
- produrre una configurazione valida senza leggere il codice Go;
- diagnosticare provider, modello, workspace e agent;
- comprendere e controllare ogni mutazione proposta;
- completare lo scenario Laravel live entro i limiti dichiarati;
- comprendere cosa Maestro protegge e cosa non protegge;
- ripetere la procedura da un ambiente pulito.
