# Maestro Identity

Versione: 0.1.0

Stato: Draft

Ultimo aggiornamento: 2026-07-20

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---
# Perché esiste questo documento?
Questo documento definisce l'identità del progetto Maestro. Descrive il problema che il progetto intende risolvere, il suo ruolo nell'ecosistema e i principi che ne guidano l'evoluzione. Le decisioni riportate costituiscono il riferimento per tutte le successive scelte architetturali.

# Introduzione

L'intelligenza artificiale sta evolvendo rapidamente.

Ogni mese vengono introdotti nuovi modelli, nuovi strumenti e nuovi paradigmi di sviluppo. Tuttavia, per la maggior parte degli sviluppatori, costruire una AI locale realmente utile rimane un processo complesso, frammentato e fortemente dipendente dagli strumenti scelti.

Installare un modello è semplice.

Costruire un ecosistema che sappia orchestrare modelli, strumenti, framework e contesto è la vera sfida.

Maestro nasce per risolvere questo problema.

---

# Che cos'è Maestro?

Maestro è un runtime locale, estensibile e framework-aware progettato per orchestrare agenti AI dedicati allo sviluppo software.

Non è un Large Language Model.

Non è un chatbot.

Non è un'estensione per un editor.

Non è un IDE.

Maestro è il livello di orchestrazione che permette a tutte queste componenti di collaborare in modo coerente.

---

# Missione

Fornire agli sviluppatori un'infrastruttura aperta, modulare e indipendente dal provider, capace di trasformare un modello AI in un vero assistente di sviluppo.

L'obiettivo non è sostituire lo sviluppatore.

L'obiettivo è costruire il miglior ambiente possibile affinché lo sviluppatore possa lavorare insieme alla propria AI.

---

# Il ruolo di Maestro nell'ecosistema

Quando uno sviluppatore PHP pensa alla gestione delle dipendenze, pensa a Composer.

Composer non sviluppa l'applicazione.

Composer orchestra il suo ecosistema.

Maestro vuole rappresentare lo stesso concetto nel mondo dell'AI locale.

Il suo compito è orchestrare:

- Provider AI
- Plugin
- Workspace
- Tool
- Context Engine
- Agenti
- Framework
- Automazioni

Il runtime coordina queste componenti senza imporre un particolare modello, provider o ambiente di sviluppo.

---

# Il problema che risolve

La difficoltà nell'utilizzo di una AI locale non consiste nell'installare un modello.

La difficoltà consiste nel costruire un sistema capace di:

- comprendere il progetto;
- conoscere il framework utilizzato;
- utilizzare gli strumenti disponibili;
- adattarsi all'hardware dello sviluppatore;
- ottimizzare il contesto inviato al modello;
- eseguire attività complesse in autonomia.

Maestro vuole fornire le fondamenta di questo sistema.

Le ossa.

L'architettura.

Il resto rimane una scelta dello sviluppatore.

Ogni utilizzatore potrà decidere:

- quale modello utilizzare;
- quale provider installare;
- quali plugin abilitare;
- quali strumenti integrare;
- quali limiti imporre all'agente.

---

# Perché Maestro è diverso

Esistono già strumenti eccellenti come Continue, Aider, Claude Code e Codex CLI.

Questi strumenti risolvono brillantemente il problema dell'interazione tra sviluppatore e modello AI.

Maestro affronta un livello differente.

L'obiettivo non è collegare uno sviluppatore ad un modello.

L'obiettivo è costruire il runtime che rende possibile questa collaborazione.

L'intelligenza non risiede solamente nel modello.

Risiede anche nell'orchestrazione.

---

# Filosofia

Ogni componente deve avere una singola responsabilità.

Ogni componente deve poter essere sostituito.

Ogni provider deve poter essere cambiato.

Ogni plugin deve poter essere installato o rimosso senza modificare il core.

L'utente mantiene sempre il controllo del proprio ambiente.

Maestro non decide al posto dello sviluppatore.

Fornisce gli strumenti affinché lo sviluppatore possa costruire la propria AI locale.

---

# Valori

Il progetto si fonda su alcuni principi fondamentali.

## Apertura

Il runtime non deve dipendere da uno specifico provider.

## Modularità

Ogni componente deve poter evolvere indipendentemente.

## Portabilità

Il sistema deve adattarsi all'hardware disponibile.

## Trasparenza

Ogni decisione presa dal runtime deve essere comprensibile.

## Estensibilità

Nuovi provider, plugin e strumenti devono poter essere aggiunti senza modificare il core.

---

# Visione

L'obiettivo di lungo periodo è rendere Maestro il punto di riferimento per chi desidera costruire una AI locale dedicata allo sviluppo software.

Non il modello migliore.

Non il plugin migliore.

Ma l'infrastruttura migliore.

---

# Motto

> The intelligence is in the orchestration.

---

# Decisioni

- Il nome ufficiale del progetto è Maestro.
- Forge non rappresenta più il nome del progetto.
- Gestor sarà il nome di un componente interno dedicato alla gestione delle capability.
- Il runtime sarà completamente provider-agnostic.
- L'architettura sarà basata su Runtime, Plugin, Provider e Tool.

---

# Domande aperte

- Definizione formale del sistema di Capability.
- Definizione del ciclo di vita degli Agent.
- Gestione della memoria condivisa tra plugin.
- Sistema di eventi interno.

---

# Prossimi passi

- Scrivere principles.md.
- Definire vision.md.
- Progettare architecture.md.
