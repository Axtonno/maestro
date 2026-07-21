# Maestro Principles

Versione: 0.1.0

Stato: Draft

Ultimo aggiornamento: 2026-07-20

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Perché esiste questo documento?

La filosofia descrive come pensa Maestro.

Questo documento traduce quella filosofia in principi concreti che guidano la progettazione e l'implementazione del software.

Ogni decisione architetturale dovrà essere coerente con questi principi.

---

# Obiettivo

Definire le regole progettuali che permettono al progetto di rimanere coerente nel tempo.

I principi rappresentano il collegamento tra la filosofia del progetto e la sua architettura.

---

# Principio 1 — Core minimale

Il core del runtime deve contenere solamente ciò che è essenziale.

Ogni funzionalità che può vivere come plugin non deve essere implementata nel core.

---

# Principio 2 — Provider agnostic

Il runtime non deve conoscere le particolarità di Ollama, llama.cpp, OpenAI o altri provider.

Ogni provider implementa una stessa interfaccia.

---

# Principio 3 — Plugin first

Framework, linguaggi e integrazioni appartengono ai plugin.

Il core non contiene logica specifica per Laravel, Symfony, Python o altri ecosistemi.

---

# Principio 4 — Capability over implementation

Il runtime non ricerca implementazioni.

Ricerca capability.

Sarà compito di Gestor individuare quale componente è in grado di soddisfare una determinata richiesta.

---

# Principio 5 — Single Responsibility

Ogni componente possiede una sola responsabilità chiaramente definita.

---

# Principio 6 — Estendibilità

Ogni modulo deve poter essere esteso senza modificare il codice del runtime.

---

# Principio 7 — Hardware awareness

Il runtime deve conoscere l'ambiente nel quale opera per adattare il proprio comportamento.

---

# Principio 8 — Documentazione prima del codice

Ogni modifica significativa dell'architettura deve essere descritta nella documentazione prima dell'implementazione.

---

# Principio 9 — Compatibilità

Le modifiche future devono preservare la compatibilità delle interfacce pubbliche quando possibile.

---

# Principio 10 — Qualità prima della quantità

È preferibile implementare poche funzionalità ben progettate piuttosto che molte funzionalità difficili da mantenere.

---

# Decisioni

- Il runtime rimarrà piccolo.
- Il sistema sarà guidato dalle capability.
- Le implementazioni saranno sostituibili.
- L'estensione avverrà tramite plugin.
- L'architettura verrà progettata prima dell'implementazione.

---

# Documenti dipendenti

- vision.md
- architecture.md
- specifications/*
