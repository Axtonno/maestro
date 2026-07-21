# Maestro Design Decisions

Versione: 0.1.0

Stato: Living Document

Ultimo aggiornamento: 2026-07-20

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Perché esiste questo documento?

Questo documento raccoglie i criteri permanenti con cui vengono prese le decisioni progettuali.

Non descrive decisioni specifiche (che appartengono agli ADR), ma il metodo con cui esse vengono valutate.

---

# Obiettivo

Garantire coerenza nelle scelte progettuali durante l'evoluzione del progetto.

---

# Quando aggiungere un componente al core?

Una funzionalità appartiene al core solo se:

- è indispensabile per il funzionamento del runtime;
- è indipendente da provider e framework;
- non può essere implementata come plugin.

In caso di dubbio, la funzionalità deve vivere fuori dal core.

---

# Quando creare un plugin?

Una funzionalità deve diventare plugin quando:

- introduce dipendenze esterne;
- è legata a un framework;
- è legata a un linguaggio;
- è opzionale.

---

# Quando introdurre una nuova interfaccia?

Una nuova interfaccia è giustificata quando esistono almeno due implementazioni plausibili.

Le interfacce non devono essere create in anticipo senza una reale esigenza.

---

# Quando aggiungere una dipendenza?

Prima di introdurre una libreria occorre verificare:

- valore aggiunto;
- maturità del progetto;
- manutenzione;
- licenza;
- impatto sull'architettura.

Ogni dipendenza deve avere una motivazione documentata.

---

# Quando rompere la retrocompatibilità?

Solo se:

- semplifica significativamente il progetto;
- riduce il debito tecnico;
- migliora l'architettura.

La motivazione deve essere registrata in un ADR.

---

# Come affrontare una nuova funzionalità?

Ogni proposta segue questo flusso:

1. Analisi del problema.
2. Verifica della filosofia.
3. Verifica dei principi.
4. Analisi architetturale.
5. Aggiornamento della documentazione.
6. Implementazione.
7. Test.
8. Commit.

---

# Decisioni

- La semplicità prevale sulla completezza.
- Le interfacce sono più importanti delle implementazioni.
- Ogni nuova dipendenza richiede una motivazione esplicita.
- Il core deve rimanere minimale.

---

# Documenti dipendenti

- architecture.md
- adr/*
