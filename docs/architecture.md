# Maestro Architecture

Versione: 0.1.0

Stato: Draft

Ultimo aggiornamento: 2026-07-20

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

Invarianti del Runtime interno

Le implementazioni contenute in `internal/runtime` nascondono la propria rappresentazione e consentono modifiche soltanto attraverso operazioni controllate.

Ogni tipo protegge gli invarianti del proprio livello di responsabilità. Gli invarianti locali appartengono al tipo proprietario, mentre quelli che coinvolgono più entità sono coordinati dal relativo aggregato.

Questa convenzione permette di modificare e ottimizzare l'implementazione interna senza ampliare i contratti pubblici o distribuire la responsabilità degli invarianti tra più componenti.

---

## Provider Layer

Responsabilità:

- comunicazione con i modelli;
- gestione delle conversazioni;
- streaming;
- embedding.

Implementazioni previste:

- Ollama
- llama.cpp
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

- caricamento plugin;
- registrazione plugin;
- lifecycle plugin.

Esempi:

- Laravel
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
