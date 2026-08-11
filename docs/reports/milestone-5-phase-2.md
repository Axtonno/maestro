# Milestone 5 — Report Fase 2

Fase: Catalogo, registry e caricamento

Stato: Completata

Data: 2026-08-11

---

# Obiettivo

Consolidare catalogo loader, registry plugin e percorso di `Load` sotto errori,
cancellazione e concorrenza, dimostrando che codice esterno ed eventi non
vengono invocati mantenendo lock del Plugin Runtime.

---

# Esito

L'implementazione baseline rispettava già gli invarianti della fase. Non sono
state necessarie modifiche al comportamento o alle firme pubbliche; sono stati
stabilizzati il contratto concorrente, la documentazione e la copertura di
regressione.

Sono ora dimostrati:

- catalogo e registry thread-safe;
- snapshot difensivi di `Available` e `Registered`;
- ordine sequenziale di registrazione preservato;
- ordine relativo intenzionalmente non specificato per chiamate concorrenti;
- factory eseguite senza lock del catalogo;
- registrar del Runtime Core invocato senza lock del registry plugin;
- eventi pubblicati dopo il rilascio dei lock e capaci di rientrare nel Plugin
  Runtime;
- nessuna visibilità nel registry dedicato prima del successo del registrar;
- una sola registrazione riuscita per load concorrenti dello stesso ID;
- load concorrenti di ID differenti tutti indipendenti;
- cause di loader e Runtime Core preservate con `errors.Is`;
- nessun indice plugin prodotto da failure, cancellazione, risultato nil, ID
  incoerente o manifest incompatibile.

---

# Semantica concorrente

Ogni chiamata `Load` è un tentativo indipendente. Chiamate concorrenti sullo
stesso ID possono invocare la factory più volte; il Registry globale garantisce
che una sola registrazione riesca. Le altre chiamate restituiscono sia
`plugin.ErrAlreadyRegistered` sia `runtime.ErrAlreadyRegistered`.

Non viene introdotto singleflight implicito: context, cancellazione e risultato
di tentativi distinti non vengono accorpati. Per questo il contratto `Loader`
documenta che la factory deve proteggere eventuale stato mutabile locale e non
deve acquisire risorse longeve.

L'ordine di `Available` e `Registered` corrisponde alle registrazioni riuscite.
Per operazioni concorrenti il relativo ordine dipende dal completamento e non è
una garanzia di ranking o selezione.

---

# Failure e atomicità

Il percorso di load distingue:

- ID, loader o context invalidi;
- loader non presente;
- errore della factory;
- risultato nil o con ID incoerente;
- manifest mancante o incompatibile;
- rifiuto del Runtime Core.

Factory fallita e risultato invalido includono `ErrLoadFailed`; i sentinel più
specifici e le cause originali restano ispezionabili. Errori di manifest e del
Runtime Core conservano il proprio dominio senza una tassonomia duplicata.

Il Plugin Runtime aggiorna l'indice dedicato soltanto dopo il successo del
registrar. Il catalogo non viene modificato da `Load`, anche in caso di errore.

---

# Test aggiunti

- registrazione concorrente di plugin con letture `Resolve` e `Registered`;
- registrazione concorrente di cento loader con snapshot `Available`;
- factory bloccante mentre il catalogo resta utilizzabile;
- registrar bloccante mentre il registry plugin resta leggibile;
- Event Bus re-entrant su tutti i topic plugin;
- sedici load concorrenti dello stesso ID con un vincitore;
- sessantaquattro load concorrenti di ID differenti;
- combinazioni `ErrLoadFailed` + causa o `ErrInvalidPlugin`;
- incompatibilità manifest e rifiuto del registrar durante `Load`;
- assenza di plugin indicizzati per ogni failure.

I test bloccanti usano deadline di due secondi per trasformare eventuali
deadlock in failure deterministiche.

---

# API e documentazione

`pkg/plugin.Runtime` mantiene le stesse firme. I commenti pubblici specificano:

- snapshot in ordine di registrazione riuscita;
- ordine relativo non specificato sotto concorrenza;
- tentativi `Load` indipendenti;
- possibile invocazione concorrente della stessa factory;
- una sola registrazione riuscita per ID.

`docs/plugin-runtime.md`, piano, roadmap, README e contesto sono allineati.

---

# Verifica

Comandi del gate:

```text
GOCACHE=/tmp/maestro-m5p2-test go test ./...
GOCACHE=/tmp/maestro-m5p2-race go test -race ./...
GOCACHE=/tmp/maestro-m5p2-vet go vet ./...
git diff --check
```

Esito: tutti i comandi superati.

---

# Gate di uscita

Superato:

- suite plugin e repository-wide verde;
- race detector verde;
- failure e concorrenza lasciano catalogo e registry coerenti;
- loader, registrar ed eventi dimostrati fuori lock;
- nessuna modifica breaking;
- semantica concorrente pubblica e deterministica nei confini dichiarati.

La Fase 2 è completata. La Fase 3 — Lifecycle, dependency graph e Gestor è
pronta.
