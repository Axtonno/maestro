# Context Engine — Workspace Indexing

Versione: 0.1.0

Stato: Fase 2 implementata

Data: 2026-08-11

---

# Scopo

Descrivere la source filesystem e la pubblicazione degli snapshot introdotte
dalla Fase 2 della Milestone 6.

---

# Source filesystem

`context.filesystem` è registrata automaticamente dall'engine interno. La
source accetta soltanto root assolute, normalizzate, esistenti e fisicamente
directory. Una root symlink viene rifiutata.

La scansione usa path logici relativi con separatore `/`. `filepath.Rel` e la
validazione di `DocumentPath` impediscono path assoluti, traversal e separator
non portabili.

I symlink incontrati durante il walk non vengono seguiti né indicizzati. Prima
e dopo la lettura ogni file viene confrontato tramite identità, dimensione e
modification time; una sostituzione osservata invalida l'intero refresh.

---

# Policy

`ScanPolicy` governa:

- numero massimo di documenti;
- byte massimi per documento;
- byte massimi complessivi;
- pattern include ed exclude;
- inclusione di path nascosti;
- inclusione di contenuti binari opachi.

I limiti devono essere positivi e il limite per file non può superare quello
totale. Vengono applicati durante la scansione. Il superamento conserva
`ErrLimitExceeded` e impedisce la pubblicazione.

I pattern operano sui path logici. Pattern privi di `/`, come `*.go`, operano
anche sul basename. Il suffisso `/**` include o esclude ricorsivamente un
prefisso. La baseline non dichiara compatibilità completa con `.gitignore`.

I default escludono `.git`, `vendor` e `node_modules`, non includono file
nascosti o binari e limitano la scansione a 10.000 file, 1 MiB per file e
64 MiB complessivi.

---

# Testo e contenuto opaco

I file UTF-8 vengono normalizzati rimuovendo il BOM e convertendo CRLF e CR in
LF prima del digest. Media type e language hint sono deterministici per le
estensioni supportate; gli altri file UTF-8 usano `text/plain`.

Contenuti con NUL o UTF-8 invalido sono binari. Per default vengono ignorati.
Con `IncludeBinary` diventano documenti `application/octet-stream` senza
language hint. Possono partecipare allo snapshot e al digest, ma retrieval e
analyzer testuali devono ignorarli.

---

# Pubblicazione

L'engine risolve la source sotto read lock e la invoca senza lock interni. Il
risultato completo viene validato prima della pubblicazione.

Per ogni workspace:

- la prima pubblicazione usa generazione 1;
- ogni pubblicazione riuscita incrementa la generazione;
- failure, cancellazione o risultato invalido non sostituiscono lo snapshot;
- lettori concorrenti osservano un valore completo;
- operazioni concorrenti vengono committate nell'ordine di completamento.

La snapshot map è autorevole; la cache non è ancora introdotta. Non vengono
creati watcher o goroutine persistenti.

---

# Errori e cancellazione

Errori della source sono wrappati con `ErrSourceFailure`; cause, cancellazione
e deadline restano ispezionabili. Risultati sintatticamente invalidi producono
`ErrInvalidSnapshot` e non sono presentati come una scansione riuscita.

Il context viene verificato prima della risoluzione, durante il walk, durante
la lettura e prima del commit. Nessuna entry parziale diventa visibile.

---

# Limiti

- nessun watcher filesystem;
- nessuna persistenza;
- nessun parsing `.gitignore` completo;
- nessun follow symlink;
- nessuna analisi AST nella source;
- nessun retrieval di contenuto binario;
- mutation detection basata su metadata osservabili, non su snapshot
  filesystem transazionali.
