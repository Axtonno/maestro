# Maestro Architecture

Versione: 0.1.0

Stato: Draft

Ultimo aggiornamento: 2026-08-06

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Perché esiste questo documento?

Questo documento descrive l'architettura logica di Maestro.

Definisce i componenti fondamentali, le loro responsabilità e le relazioni tra essi.

Non descrive implementazioni specifiche, ma il contratto architetturale che ogni implementazione dovrà rispettare.

---

# Obiettivo

Costruire un runtime modulare, estensibile e provider-agnostic capace di orchestrare strumenti, provider, plugin e agenti AI dedicati allo sviluppo software.

---

# Architettura generale

```
                    +----------------+
                    |      CLI       |
                    +----------------+
                            |
                            v
                    +----------------+
                    |    Runtime     |
                    +----------------+
                            |
        +---------+---------+---------+---------+
        |         |         |         |         |
        v         v         v         v         v
   Provider    Gestor   Context    Plugins    Tools
                Layer     Engine
        |                   |
        |                   |
        +---------+---------+
                  |
                  v
               Agent
```

Il Runtime rappresenta il punto di ingresso dell'intero sistema.

Ogni componente comunica attraverso interfacce ben definite.

---

# Componenti principali

## Runtime

Responsabilità:

- bootstrap del sistema;
- configurazione;
- lifecycle;
- orchestrazione;
- dependency injection;
- event bus.

Il Runtime non contiene logica di dominio.

L'Event Bus interno permette la comunicazione disaccoppiata tra componenti.
La consegna è sincrona, ordinata per sottoscrizione e thread-safe. Il Runtime e
il `Context` dei componenti espongono la stessa istanza del bus.

Il Runtime compone inoltre un Provider Runtime condiviso. Applicazioni e
componenti usano la stessa istanza per registrare, risolvere e invocare provider
senza conoscere le implementazioni concrete.

Il composition root pubblico compone anche un Plugin Runtime. La registrazione
di un plugin confluisce nel Registry dei componenti, così dependency graph,
stato e lifecycle rimangono unici.

La configurazione viene iniettata nel composition root e consegnata ai
componenti come snapshot a chiavi esatte. Il core non decide come leggere file,
variabili d'ambiente o secret esterni.

Invarianti del Runtime interno

Le implementazioni contenute in `internal/runtime` nascondono la propria rappresentazione e consentono modifiche soltanto attraverso operazioni controllate.

Ogni tipo protegge gli invarianti del proprio livello di responsabilità. Gli invarianti locali appartengono al tipo proprietario, mentre quelli che coinvolgono più entità sono coordinati dal relativo aggregato.

Questa convenzione permette di modificare e ottimizzare l'implementazione interna senza ampliare i contratti pubblici o distribuire la responsabilità degli invarianti tra più componenti.

---

## Provider Layer

Responsabilità:

- comunicazione con i modelli;
- rappresentazione delle richieste conversazionali;
- streaming;
- embedding.

Il contratto del layer è capability-based. L'identità del provider è separata
dalle capability di completion, streaming, embedding e model listing.

Il Provider Runtime mantiene un registry thread-safe, applica una selezione
esplicita del provider predefinito e inoltra le operazioni senza mantenere lock
durante l'esecuzione di codice esterno.

Implementazioni disponibili:

- Ollama
- llama.cpp

Implementazioni previste:

- LM Studio
- OpenAI
- Anthropic

---

## Gestor

Gestor rappresenta il registro centrale delle capability.

Responsabilità:

- registrazione dei componenti;
- discovery;
- dependency resolution;
- capability lookup.

Gestor non esegue codice.

Coordina il sistema.

---

## Plugin System

Responsabilità:

- registrazione plugin;
- lifecycle plugin;
- risoluzione dei plugin registrati;
- catalogo e caricamento dei loader;
- validazione della compatibilità;
- pubblicazione degli eventi plugin;
- integrazione con il dependency graph globale.

I plugin sono componenti Go fidati caricati in-process. Un catalogo di loader
fornisce discovery deterministica; il Plugin Runtime valida ID e manifest, li
indicizza e pubblica gli eventi, mentre il Runtime Core orchestra il lifecycle.
Distribuzione di artefatti esterni e isolamento sono livelli successivi.

Esempi:

- Laravel (implementato)
- Symfony
- Django
- React

---

## Context Engine

Responsabilità:

- analisi workspace;
- indicizzazione;
- AST;
- gestione della memoria;
- costruzione del contesto.

Il Context Engine rappresenta il "cervello" del runtime.

---

## Tool System

Responsabilità:

- filesystem;
- git;
- terminale;
- docker;
- composer;
- artisan;
- phpunit.

Ogni tool implementa una capability.

---

## Agent System

Responsabilità:

- pianificazione;
- esecuzione task;
- utilizzo dei tool;
- gestione delle autorizzazioni.

Gli agenti non comunicano direttamente con i provider.

Utilizzano il Runtime.

---

# Flusso di una richiesta

1. La CLI riceve una richiesta.
2. Il Runtime crea il contesto.
3. Gestor individua le capability necessarie.
4. Il Context Engine prepara il workspace.
5. Il Provider comunica con il modello.
6. L'Agent coordina gli strumenti.
7. Il Runtime restituisce il risultato.

---

# Dipendenze

Le dipendenze devono sempre puntare verso il centro dell'architettura.

Mai il contrario.

Il Runtime conosce tutti.

I componenti non devono conoscersi direttamente.

---

# Estensibilità

Ogni componente deve poter essere sostituito senza modificare il Runtime.

Nuovi provider.

Nuovi plugin.

Nuovi tool.

Nuovi agenti.

Devono poter essere aggiunti come implementazioni.

---

# Decisioni

- Runtime minimale.
- Provider completamente astratti.
- Gestor come registry delle capability.
- Plugin indipendenti.
- Context Engine separato dal Provider.
- Tool completamente modulari.

---

# Documenti dipendenti

- ADR
- Specifiche tecniche
- Implementazioni
