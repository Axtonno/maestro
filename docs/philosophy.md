# Maestro Philosophy

Versione: 0.1.0

Stato: Draft

Ultimo aggiornamento: 2026-07-20

Autori:
- Antonio Cafeo
- OpenAI ChatGPT

---

# Perché esiste questo documento?

Questo documento descrive il modo in cui Maestro affronta i problemi.

Non definisce componenti software.

Non descrive implementazioni.

Non contiene dettagli architetturali.

Definisce il modo di pensare che guida ogni decisione del progetto.

Ogni scelta tecnica dovrà essere coerente con questa filosofia.

---

# Obiettivo

Costruire un runtime che possa evolvere nel tempo senza perdere semplicità, coerenza e libertà.

La filosofia rappresenta il filtro attraverso il quale valutare ogni nuova idea.

Una funzionalità tecnicamente interessante non è automaticamente una buona funzionalità per Maestro.

---

# La filosofia di Maestro

## L'orchestrazione è più importante dell'intelligenza

Un modello AI è solamente uno dei componenti del sistema.

La qualità del risultato dipende principalmente da come vengono orchestrati:

- contesto;
- strumenti;
- plugin;
- provider;
- workspace;
- capacità disponibili.

Un runtime ben progettato permette anche a modelli relativamente piccoli di produrre risultati eccellenti.

---

## La semplicità è una scelta progettuale

La semplicità non significa avere meno funzionalità.

Significa costruire sistemi che risultino facili da comprendere, mantenere ed estendere.

Ogni componente deve poter essere spiegato in poche frasi.

Se un componente diventa difficile da spiegare, probabilmente sta assumendo troppe responsabilità.

---

## Il runtime non deve essere protagonista

L'obiettivo di Maestro non è sostituire gli strumenti esistenti.

Il suo compito è coordinarli.

Un buon runtime lavora in modo quasi invisibile.

Lo sviluppatore utilizza il proprio editor, il proprio framework e il proprio provider.

Maestro rende possibile la collaborazione tra queste componenti.

---

## La libertà dello sviluppatore viene prima

Lo sviluppatore mantiene sempre il controllo.

Può decidere:

- quale provider utilizzare;
- quale modello installare;
- quali plugin abilitare;
- quali strumenti concedere all'agente;
- quali limiti imporre.

Maestro non impone un ecosistema.

Lo rende possibile.

---

## L'hardware non è un limite

Ogni computer rappresenta un ambiente differente.

Una workstation professionale e un portatile di cinque anni devono poter utilizzare lo stesso runtime.

Cambieranno le implementazioni.

Non cambierà l'architettura.

---

## Le interfacce durano più delle implementazioni

I provider cambieranno.

I modelli cambieranno.

I framework cambieranno.

Le interfacce progettate correttamente possono invece rimanere stabili per molti anni.

Per questo motivo Maestro investe prima nella definizione delle interfacce e solo successivamente nelle implementazioni.

---

## Il codice segue la documentazione

La documentazione rappresenta la fonte di verità del progetto.

Prima si definisce una responsabilità.

Poi si progetta l'interfaccia.

Infine si implementa il codice.

Mai il contrario.

---

## Ogni componente deve poter essere sostituito

Nessuna implementazione è definitiva.

Ogni modulo deve poter essere sostituito senza compromettere il funzionamento del runtime.

La sostituibilità rappresenta uno dei principali indicatori della qualità architetturale.

---

## L'evoluzione deve essere incrementale

Maestro non nascerà completo.

Ogni nuova versione aggiungerà capacità senza compromettere ciò che è già stabile.

Le fondamenta vengono costruite una sola volta.

Per questo motivo devono essere solide.

---

# Come valutare una nuova funzionalità

Prima di implementare una nuova idea è necessario rispondere ad alcune domande.

- Rispetta la filosofia del progetto?
- Riduce oppure aumenta la complessità?
- Appartiene realmente al core?
- Può vivere come plugin?
- Mantiene la libertà dello sviluppatore?
- Rende il runtime più semplice oppure più difficile da comprendere?

Se la risposta a una di queste domande è negativa, la proposta deve essere rivalutata.

---

# Motto

> Build small.
>
> Orchestrate everything.

---

# Decisioni

- Il runtime privilegia l'orchestrazione rispetto alla complessità.
- Il codice segue sempre la documentazione.
- Le implementazioni sono temporanee.
- Le interfacce rappresentano il patrimonio più importante del progetto.
- Lo sviluppatore mantiene sempre il controllo.

---

# Documenti dipendenti

- principles.md
- vision.md
- architecture.md
