# Milestone 7 — Agent System Design

Versione: 0.1.0

Stato: Progettato — milestone aperta

Data: 2026-08-11

---

# Contesto

Maestro possiede le primitive necessarie per costruire il primo agente senza
collocare logica agentica nel Runtime Core:

- Runtime Core, lifecycle, configurazione, Registry ed Event Bus condiviso;
- Provider Runtime con completion, streaming, output strutturato, tool calling,
  error semantics, resilienza e osservabilità;
- Gestor per discovery e risoluzione spiegabile delle capability;
- Plugin System trusted in-process e un workspace provider Laravel;
- Context Engine per snapshot, retrieval e bundle con provenance e budget;
- Benchmark Layer con fixture provider e task di sviluppo riproducibili.

Mancano ancora un contratto eseguibile per i tool, un permission model, lo
stato di una sessione, un piano verificabile e il coordinamento bounded del
ciclo modello–tool. La Milestone 7 introduce questi confini senza trasformare
il Runtime Core in un agente e senza affidare al modello autorizzazioni o
selezione implicita delle risorse.

---

# Obiettivo

Consegnare un primo agente autonomo, workspace-aware e provider-agnostic capace
di pianificare un task, costruire il contesto, richiedere tool, ottenere
autorizzazioni ed eseguire passi entro limiti dichiarati.

Il percorso logico è:

```text
request
   |
   v
session -> plan -> context bundle -> model turn
                                     |
                           +---------+---------+
                           |                   |
                           v                   v
                      final answer        tool calls
                                               |
                                               v
                                  prepare -> authorize -> execute
                                               |
                                               v
                                          model result
```

Il sistema deve rendere ispezionabili piano, stato, limiti, decisioni di
autorizzazione ed esiti tecnici, senza pubblicare prompt, contenuto del
workspace, argomenti sensibili o output dei tool negli eventi osservazionali.

---

# Confini della milestone

## Incluso

- contratti pubblici separati per Tool System e Agent System;
- catalogo tool thread-safe con descriptor e JSON Schema provider-neutral;
- preparazione e validazione delle invocazioni prima di ogni effetto;
- permission model default-deny con policy e approvazione opzionale;
- azioni tipizzate per lettura, modifica, esecuzione, rete e disclosure;
- sessioni in-memory con snapshot immutabili e transizioni validate;
- piani espliciti, task step e risultati terminali;
- budget per turni modello, tool call, tempo e dimensione degli output;
- loop tool call/result attraverso il Provider Runtime condiviso;
- provider, modello, workspace, agente e policy scelti esplicitamente;
- Context Engine usato per costruire evidenza workspace con provenance;
- un set minimo di tool workspace per dimostrare lettura e modifica controllata;
- un reference agent general-purpose registrato nel composition root;
- integrazione Gestor, eventi redatti, test concorrenti e benchmark.

## Escluso

- memoria di lungo periodo, profilo utente o persistenza delle conversazioni;
- esecuzione distribuita, code durable e ripresa dopo restart del processo;
- delega multi-agent, swarm, agenti ricorsivi o consenso tra agenti;
- selezione automatica di provider, modello o tool basata su ranking nascosto;
- privilegi impliciti derivati dal prompt o dalla risposta del modello;
- esecuzione di codice non fidato in sandbox o isolamento di processo dei plugin;
- shell completa, Docker, Composer, Artisan, PHPUnit e Git write come requisito
  del primo reference agent;
- credenziali, secret manager o inoltro automatico di variabili d'ambiente;
- watcher filesystem e reindicizzazione continua;
- marketplace, packaging esterno e tool di terze parti non fidati;
- UI/CLI completa per l'approvazione interattiva.

Sandbox forte, distribuzione di estensioni e CLI completa appartengono alla
Milestone 8. Il permission model di questa milestone limita ciò che Maestro
decide di invocare; non trasforma codice Go trusted in-process in codice
isolato.

---

# Modello architetturale

## Due servizi, ownership separate

Il Tool System e l'Agent System hanno responsabilità differenti:

```text
                          Maestro Runtime
                +---------------+---------------+
                |               |               |
                v               v               v
          Provider Runtime  Context Engine    Gestor
                ^               ^               ^
                |               |               |
                +-------- Agent Runtime --------+
                              |
                              v
                         Tool Runtime
                              |
                    prepare / authorize / execute
```

Il package pubblico `pkg/tool` definisce descriptor, invocation, action,
permission, result, registry ed executor. L'implementazione concreta vive in
`internal/tool`.

Il package pubblico `pkg/agent` definisce agent identity, request, limits,
session, plan, step, run result e runtime. L'implementazione concreta vive in
`internal/agent`.

Il composition root espone entrambi in modo additivo tramite `Runtime.Tools()`
e `Runtime.Agents()`. `pkg/runtime.Runtime` e `runtime.Context` non vengono
ampliati: componenti esistenti non acquisiscono implicitamente accesso agli
strumenti o all'esecuzione agentica.

## Gli agenti non possiedono i sottosistemi

L'Agent Runtime coordina servizi già autorevoli:

- Provider Runtime possiede registrazione, routing e chiamate ai modelli;
- Context Engine possiede snapshot, retrieval e bundle;
- Tool Runtime possiede catalogo, autorizzazione ed execution boundary;
- Gestor descrive capability e dependency plan, ma non esegue codice;
- Runtime Core possiede componenti, lifecycle, stato ed Event Bus.

Una sessione conserva riferimenti e snapshot necessari al run, non copie di
registry, provider, documenti o stato lifecycle.

---

# Tool System

## Descriptor

Ogni tool dichiara almeno:

- ID stabile e namespaced;
- nome compatibile con il contratto tool calling dei provider;
- versione;
- descrizione destinata al modello;
- schema JSON degli argomenti;
- limiti dichiarativi e categorie di effetto possibili.

ID interno e nome esposto al modello restano distinti. Il nome provider deve
essere univoco nel set consegnato a una singola completion. Collisioni non
vengono risolte con rinomina o ordine di registrazione impliciti.

Il descriptor è metadata. Non costituisce autorizzazione e non prova che una
specifica invocazione sia sicura.

## Prepare, authorize, execute

Ogni invocazione attraversa tre confini:

1. `Prepare` valida gli argomenti, risolve il workspace e normalizza risorse e
   azioni senza produrre effetti;
2. l'`Authorizer` valuta le azioni esatte preparate e produce una decisione;
3. `Execute` riceve soltanto una prepared invocation autorizzata e produce un
   risultato limitato.

```text
untrusted model arguments
          |
          v
       Prepare ----invalid----> typed failure
          |
          v
      Actions[]
          |
          v
      Authorize ---deny-------> denied tool result
          |
          v
       Execute ----failure----> typed tool result
          |
          v
        Result
```

Preparazione e autorizzazione non vengono eseguite sotto i lock del catalogo.
Prepared invocation e action sono value object immutabili dal punto di vista
del chiamante. I tool restano codice trusted; l'executor protegge il percorso
normale, mentre la sandbox contro implementazioni malevole resta fuori scope.

## Risultati e limiti

Un risultato distingue:

- successo applicativo;
- rifiuto di autorizzazione;
- input non valido;
- failure del tool;
- cancellazione;
- output troncato.

Il contenuto destinato al modello è separato dai metadata osservazionali.
Dimensione massima, numero di elementi e politica di troncamento sono
dichiarati dal run o dal tool e applicati prima di aggiungere il risultato alla
memoria della sessione.

Gli errori preservano cause e sentinel tramite `errors.Is`/`errors.As`, ma
error string, stdout, contenuti file e argomenti non entrano automaticamente in
eventi o log.

## Catalogo e concorrenza

Il catalogo:

- registra tool prima dell'avvio operativo del run;
- rifiuta ID, nome, versione, schema, nil e typed nil invalidi;
- rifiuta duplicati senza sostituzione implicita;
- restituisce listing ordinati e copie difensive;
- risolve per ID esatto;
- non mantiene lock durante `Prepare`, autorizzazione o `Execute`.

La baseline non introduce unload o hot replacement.

---

# Permission model

## Azioni, non nomi di tool

L'autorizzazione opera sulle azioni concrete prodotte da `Prepare`, non sul
solo nome del tool. Due invocazioni dello stesso tool possono quindi ottenere
decisioni diverse.

Le categorie iniziali sono:

| Effetto | Esempio |
|---|---|
| `workspace.inspect` | leggere un path logico o elencare una directory |
| `workspace.mutate` | creare o modificare un file |
| `process.execute` | avviare un processo con argv e cwd dichiarati |
| `network.access` | contattare una destinazione esplicita |
| `model.invoke` | usare provider e modello esatti |
| `model.disclose` | inviare evidenza workspace al modello |

Ogni action contiene effect, resource normalizzata, workspace opzionale e
metadata redatti necessari alla policy. I contenuti completi non sono richiesti
per decidere.

## Policy, approvazione e decisione

Una policy produce una delle seguenti decisioni:

- `allow`: esecuzione consentita per lo scope dichiarato;
- `deny`: nessun effetto;
- `prompt`: richiesta a un `Approver` configurato dal chiamante.

In assenza di una regola applicabile la decisione è `deny`. In assenza di un
Approver, `prompt` non diventa `allow`. Il modello, il tool e l'agente non
possono creare grant.

I grant possono valere per una sola invocation o per il run corrente e sono
legati al fingerprint delle action normalizzate. Non sopravvivono al processo
e non vengono estesi a risorse più ampie per prefisso implicito.

Un deny viene rappresentato come risultato tipizzato e può essere restituito
al modello perché formuli un'alternativa. Ripetizioni restano soggette ai
budget del run. Cancellazione o failure dell'Approver terminano l'invocazione
senza effetti.

## Confine di sicurezza

Il permission model controlla anche l'invocazione del modello e la disclosure
di contenuto workspace, così un futuro provider remoto non eredita
automaticamente le policy pensate per un provider locale.

Le policy non sostituiscono i controlli locali dei tool: containment dei path,
symlink, limiti, argv, environment e output devono essere validati anche
dall'implementazione proprietaria dell'effetto.

---

# Sessione, memoria e piano

## Sessione

Una sessione identifica un singolo run e contiene:

- agent ID, provider ID e model ID esatti;
- workspace ID e generazione osservata;
- richiesta utente;
- limiti e policy selezionata;
- piano e stato dei task step;
- messaggi modello e risultati tool bounded;
- contatori di turni, chiamate, token noti e durata;
- esito terminale.

Gli snapshot pubblici sono immutabili e difensivi. La mutazione appartiene
all'aggregato interno della sessione, che valida ogni transizione e pubblica
snapshot completi. La memoria iniziale è in-process e vive al massimo quanto il
run; non è una memoria utente di lungo periodo.

## Piano

Un piano è una lista ordinata e limitata di step con ID, obiettivo, stato e
dipendenze esplicite. Gli stati minimi sono `pending`, `running`, `completed`,
`failed`, `blocked` e `skipped`.

La baseline usa una fase di planning separata e validata prima del loop
operativo. Un planner può usare structured output del Provider Runtime, ma il
risultato viene parsato e validato dall'Agent Runtime; testo JSON prodotto dal
modello non diventa stato autorevole senza validazione.

Il piano è guida osservabile, non una capability di autorizzazione. Un tool non
diventa consentito perché compare nel piano. Eventuali revisioni del piano sono
versionate, limitate e registrate come transizioni di sessione.

## Limiti e terminali

Ogni request dichiara limiti positivi per:

- durata totale;
- turni modello;
- tool call totali e per turno;
- revisioni del piano;
- byte per risultato e memoria complessiva;
- token input/output quando osservabili.

Il Runtime applica hard ceiling indipendenti dalle istruzioni del prompt. Il
run termina con un reason tipizzato: completato, limite raggiunto, cancellato,
permission denied non recuperabile, provider failure, tool failure o piano
bloccato. Non esiste un loop illimitato come default.

---

# Orchestrazione provider e tool calling

Il provider e il modello sono input obbligatori del run. Gestor può verificare
che le capability richieste siano dichiarate o disponibili, ma non sceglie un
candidato tramite ordine lessicografico o fallback nascosto.

L'Agent Runtime:

1. valida request, limiti e riferimenti;
2. autorizza model invocation e disclosure applicabili;
3. costruisce o riceve il context bundle;
4. crea un piano validato;
5. traduce i descriptor tool nel contratto `provider.Tool`;
6. invoca il Provider Runtime condiviso;
7. valida ogni tool call completa e la associa a un call ID stabile;
8. prepara, autorizza ed esegue le chiamate entro i limiti;
9. aggiunge messaggi `tool` correlati e prosegue;
10. valida il terminale e chiude la sessione una sola volta.

Chiamate multiple in un turno vengono eseguite nell'ordine dichiarato nella
baseline. Non vengono parallelizzate implicitamente: ordine degli effetti,
approvazioni e risultati restano quindi comprensibili. Streaming può essere
usato per output e delta tool, ma una call viene preparata soltanto dopo aver
ricostruito arguments JSON completi.

Retry del Provider Runtime e retry semantici dell'agente restano concetti
distinti. Il run non ripete automaticamente tool mutanti dopo una failure
ambigua.

---

# Workspace awareness

Il run riceve un workspace ID registrato; il modello non sceglie root assolute.
Il Context Engine costruisce il bundle da una query e da un budget espliciti.
Le sezioni vengono delimitate e mantengono path logico, range, generation e
provenance.

I reference tool operano soltanto su path logici risolti rispetto alla root del
workspace associato alla sessione. La baseline include almeno:

- listing limitato;
- lettura limitata;
- ricerca testuale limitata;
- scrittura o patch controllata con precondizione sul contenuto osservato.

Una mutazione non modifica retroattivamente il bundle già inviato. Dopo un
effetto workspace, l'Agent Runtime marca il contesto stale e richiede un refresh
esplicito al checkpoint successivo prima di costruire nuova evidenza. Failure
di refresh non pubblicano snapshot parziali e non vengono nascoste al run.

La precondizione content-addressed riduce overwrite su contenuto cambiato tra
preparazione ed esecuzione. Containment e symlink vengono verificati al confine
del tool anche se il Context Engine aveva già validato lo snapshot.

---

# Integrazione con Gestor e plugin

Gestor indicizza descriptor per agenti e tool senza eseguirli. Capability
iniziali previste:

```text
agent.run
agent.planning
agent.workspace-aware
tool.workspace.inspect
tool.workspace.mutate
```

Il Tool Runtime resta il registry autorevole delle istanze tool. Una source
Gestor interna traduce snapshot di agenti e tool in descriptor immutabili. I
plugin futuri potranno registrare tool trusted prima dello start attraverso un
contratto additivo; packaging e trust di terze parti restano nella Milestone 8.

Laravel continua a fornire un `WorkspaceProvider` generico. L'Agent Runtime non
importa conoscenza Laravel e non invoca plugin per nome.

---

# Eventi, privacy e osservabilità

Gli eventi condividono l'Event Bus sincrono e descrivono almeno:

- session start e terminale;
- plan created/revised;
- model turn start/terminal;
- tool invocation prepared/authorized/terminal;
- transition di task step;
- limite raggiunto.

I payload possono includere run ID, agent ID, tool ID, provider ID redatto o
codificato secondo policy, contatori, durata, effect code, decisione e reason
code. Non includono richiesta utente, prompt, messaggi, context section, path
assoluti, arguments JSON, output tool, stdout/stderr, secret o error string.

Gli eventi osservazionali sono best-effort al boundary del servizio: errori e
panic degli handler non annullano uno stato già committato. Nessun callback
esterno viene eseguito mantenendo lock di sessione, catalogo o registry.

La sessione completa contiene dati sensibili ed è restituita soltanto al
chiamante autorizzato attraverso snapshot bounded; non viene usata come payload
dell'Event Bus.

---

# Invarianti principali

- il Runtime Core non contiene logica di planning o tool execution;
- Tool Runtime e Agent Runtime hanno registry e ownership separati;
- il modello propone azioni ma non autorizza effetti;
- ogni effetto usa una prepared invocation validata e una decisione esplicita;
- una decisione assente non equivale mai ad allow;
- provider, modello, workspace, agente e limiti non vengono scelti implicitamente;
- nessun tool o callback esterno viene invocato sotto lock del registry;
- una sessione ha un solo terminale e transizioni validate;
- ogni loop è limitato da hard ceiling indipendenti dal prompt;
- tool mutanti non vengono ritentati implicitamente;
- context snapshot e workspace corrente non vengono confusi;
- eventi e log non contengono prompt, contenuti, arguments o output sensibili;
- un permission model in-process non viene descritto come sandbox di sicurezza.

---

# Strategia di verifica

La suite ordinaria non richiede rete, modelli o processi esterni. Usa provider,
planner, tool, approver, workspace e clock fittizi e deve coprire:

- validazione, typed nil, duplicati e snapshot difensivi;
- prepare/authorize/execute e assenza di effetti su deny/cancel;
- policy exact-match, default-deny, prompt senza approver e grant scoped;
- path traversal, symlink, race sul contenuto e limiti degli output;
- session state machine, plan revision e singolo terminale;
- turn, tool, token, byte e deadline budget;
- tool call singole, multiple, malformate e duplicate;
- correlazione call/result e ricostruzione dei delta streaming;
- provider failure, retry già consumati e cancellazione;
- refresh del Context Engine dopo mutazione;
- resolution Gestor senza esecuzione;
- eventi redatti e observer bloccanti o in panic;
- run end-to-end dal composition root con reference agent e workspace Laravel;
- race detector sui package coinvolti e benchmark deterministici del loop.

Il gate finale richiede suite completa, `go test -race`, `go vet ./...`, audit
di compatibilità pubblica, `git diff --check` e documentazione allineata.

---

# Decisioni da formalizzare nella Fase 1

- separazione tra `pkg/tool` e `pkg/agent`;
- composition root additivo senza modifica di `pkg/runtime.Runtime`;
- prepared invocation come confine tra input non fidato ed effetto;
- permission model default-deny e indipendente dal modello;
- provider, modello e workspace espliciti;
- sessioni bounded in-memory e piani validati;
- loop sequenziale con terminale unico e hard ceiling;
- Context Engine come fonte di evidenza, non memoria dell'agente;
- eventi redatti separati dalla memoria sensibile della sessione;
- baseline trusted in-process distinta da sandbox e plugin trust.

Queste decisioni confluiranno in ADR-0025. Il piano operativo è descritto in
`agent-system-development-plan.md`.
