# Context Engine — Structured Analysis

Versione: 0.1.0

Stato: Fase 3 implementata

Data: 2026-08-11

---

# Scopo

Descrivere registry, selezione e output degli analyzer del Context Engine e
l'analyzer AST Go fornito come implementazione di riferimento.

---

# Registry

Ogni analyzer dichiara un `AnalyzerID` namespaced e una versione esatta. La
registrazione rifiuta nil, typed nil, ID/versioni invalidi e duplicati.

ID, versione, `Supports` e `Analyze` appartengono a codice trusted in-process.
Le callback operative vengono eseguite senza lock dell'engine. Panic durante
support check o analisi diventano `ErrAnalyzerFailure` e non pubblicano uno
snapshot.

---

# Selezione

`WorkspaceOptions.Analyzers` permette di dichiarare una composizione ordinata
di analyzer. Gli ID devono essere unici e registrati.

Quando la lista è vuota, l'engine valuta gli analyzer registrati in ordine di
ID. Zero matching lascia il documento testuale senza analisi; un solo matching
viene eseguito; più matching producono `ErrAmbiguous`.

Quando la lista è esplicita, tutti gli analyzer selezionati che supportano il
documento vengono eseguiti. La lista esplicita è quindi anche il contratto di
composizione; l'ordine di registrazione non diventa ranking.

Documenti `application/octet-stream` non vengono passati agli analyzer.

---

# Output validato

Ogni `Analysis` deve conservare:

- path e digest esatti del documento;
- ID e versione esatti dell'analyzer;
- simboli con ID unici e intervalli byte validi;
- relazioni con sorgente simbolo valida;
- chunk unici e contenuti nel documento;
- diagnostiche sicure riferite allo stesso path.

Un output con identità, versione o riferimenti incoerenti produce
`ErrInvalidAnalysis`. Tutti gli output vengono rivalidati dal costruttore dello
snapshot prima del commit.

---

# Analyzer Go

`context.go-ast@1` usa soltanto `go/parser`, `go/ast` e `go/token` della
libreria standard. Supporta documenti `text/x-go` con language hint `go`.

Produce:

- simbolo package;
- import;
- type e field di struct;
- const e variable;
- function e method;
- relazioni `contains` e `imports`;
- un chunk per dichiarazione top-level.

Gli intervalli sono byte half-open nel contenuto normalizzato. Sorgenti
incomplete vengono analizzate con `parser.AllErrors`: quando esiste un AST
parziale, simboli e chunk validi vengono conservati insieme alla diagnostica
`go_parse_error`. Il testo resta sempre nello snapshot.

Il modello non espone nodi AST Go e non pretende di rappresentare type checking,
call graph o semantica completa.

---

# Failure e cancellazione

Errori restituiti dagli analyzer sono wrappati con `ErrAnalyzerFailure` e
preservano la causa. `context.Canceled` e `context.DeadlineExceeded` restano
autorevoli.

Un errore operativo, panic, output invalido o cancellazione impedisce la
pubblicazione completa. Un parse error rappresentato come diagnostica valida è
invece un risultato locale e può essere pubblicato.

---

# Estensioni future

Analyzer PHP, Laravel o altri linguaggi possono implementare lo stesso
contratto in package o plugin dedicati. Il Runtime Core e l'indice filesystem
non importano parser language-specific.
