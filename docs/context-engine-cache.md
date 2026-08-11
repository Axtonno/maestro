# Context Engine — Incremental Cache

Versione: 0.1.0

Stato: Fase 5 implementata

Data: 2026-08-11

---

# Scopo

Descrivere la cache derivata in-memory usata per analisi, embedding e stime del
Context Engine.

---

# Ownership

La cache non è autorevole. Workspace e snapshot continuano a essere prodotti
dalla source e pubblicati dall'indice; una entry assente o eliminata modifica
soltanto il costo dell'operazione.

La cache non conserva snapshot, non impedisce refresh e non crea watcher. Non
serializza contenuti o vettori fuori processo.

---

# Policy

`CachePolicy` richiede limiti positivi per numero di entry e byte stimati. I
default sono 10.000 entry e 64 MiB.

Entry più grandi del limite totale non vengono inserite. Quando un inserimento
supera uno dei limiti, eviction LRU rimuove prima l'accesso meno recente; la
chiave lessicale è tie-break deterministico.

`CacheStats` espone contatori cumulativi `Hits`, `Misses`, `Evictions` e gauge
`Entries`, `Bytes`. Nessun contatore contiene path, testo o target provider.

---

# Chiavi

## Analysis

La chiave comprende:

- digest del contenuto;
- path logico;
- media type e language hint;
- analyzer ID e versione.

Il path fa parte della chiave perché un analyzer può produrre output dipendente
dal filename e perché la provenance deve restare esatta. Una generazione nuova
con la stessa chiave riusa l'analysis; rename, language o versione la invalida.

## Embedding

La chiave comprende provider, modello, dimensione osservata e SHA-256 del testo
esatto inviato. Query e candidati condividono entry soltanto quando il testo è
identico.

La dimensione autorevole per target viene osservata dalla prima risposta. Se
cambia, tutte le entry del target vengono eliminate e la richiesta corrente
fallisce; il retry riparte sulla nuova dimensione senza confrontare vettori
incompatibili.

## Estimator

La chiave comprende estimator ID, versione e SHA-256 del testo. Costi invalidi,
errori, panic o cancellazioni non entrano in cache.

---

# Immutabilità

Vettori vengono clonati in ingresso e uscita. Analysis e costi sono valori
immutabili del contratto pubblico. La cache non restituisce mappe o slice
interne condivise.

Entry vengono pubblicate soltanto dopo la validazione dell'artefatto:

- analysis dopo la costruzione riuscita dello snapshot;
- embedding dopo cardinalità, dimensione, finitezza e norma;
- costo dopo validazione dell'estimator.

---

# Concorrenza

La cache usa un lock separato e non mantiene lock durante analyzer, provider o
estimator. `get`, `put`, statistiche ed eviction sono sezioni critiche brevi.

Non esiste singleflight. Due richieste cold concorrenti mantengono context,
cancellazione e risultato indipendenti e possono invocare entrambe il producer.
Le entry valide risultanti sono equivalenti e l'ultima sostituisce la stessa
chiave senza cambiare il risultato funzionale.

---

# Equivalenza cold/warm

I test verificano:

- bundle e ranking identici;
- nessuna seconda chiamata provider su semantic retrieval warm;
- nessuna seconda stima sul bundle warm;
- riuso analysis su path stabile;
- invalidazione analysis su rename;
- purge su cambio dimensione embedding;
- eviction bounded e deterministica;
- due richieste cold realmente indipendenti.
