# Context Engine Public API Compatibility Audit

Versione: 0.1.0

Stato: Fase 1 completata

Data: 2026-08-11

---

# Scopo

Registrare la baseline pubblica introdotta dalla Milestone 6 e verificare che
non modifichi i contratti già stabili di Runtime, Gestor, Provider e Plugin.

---

# Esito

Il nuovo package `pkg/contextengine` è additivo. Nessuna firma esistente è
stata modificata e il package non è ancora collegato al composition root.

Le dipendenze pubbliche sono limitate alla libreria standard. I contratti non
importano adapter provider, plugin Laravel o implementazioni interne.

---

# Inventario API

## Identità

| Tipo | Semantica | Validazione |
|---|---|---|
| `WorkspaceID` | identità richiesta dal chiamante | valore esatto non vuoto |
| `SourceID` | source registrabile | ID namespaced |
| `AnalyzerID` | analyzer registrabile | ID namespaced |
| `EstimatorID` | estimator registrabile | ID namespaced |
| `DocumentPath` | path logico nel workspace | relativo, normalizzato, no traversal |
| `Digest` | SHA-256 del testo normalizzato | 64 caratteri hex lowercase |
| `Language` | hint neutrale | identificatore lowercase |

## Workspace e source

| Contratto | Ownership |
|---|---|
| `ScanPolicy` | limiti e filtri richiesti dal chiamante |
| `Workspace` | valore immutabile con root, source, policy e metadata |
| `WorkspaceProvider` | componente/plugin che descrive un workspace |
| `Source` | lettura cancellabile di documenti |
| `ScanResult` | documenti e diagnostiche prodotti dalla source |

`Workspace.Policy` e `Workspace.Metadata` restituiscono copie difensive. La
root assoluta è necessaria all'I/O ma non è un path documento e non è destinata
agli eventi.

## Documento, analisi e snapshot

| Contratto | Ownership |
|---|---|
| `Document` | testo immutabile e digest content-addressed |
| `SourceRange` | intervallo byte half-open nel testo UTF-8 |
| `Diagnostic` | codice sicuro e severità senza error string esterna |
| `Symbol`, `Relation`, `Chunk` | evidenza strutturale neutrale |
| `Analysis` | risultato versionato di un analyzer su un digest |
| `Snapshot` | vista ordinata e immutabile di una generazione |

Costruttori pubblici validano riferimenti e duplicati. Accessor di slice e mappe
restituiscono copie; i valori interni sono composti soltanto da scalari o tipi
immutabili.

## Retrieval e bundle

| Contratto | Ownership |
|---|---|
| `RetrievalQuery` | metodi, filtri, top-k e target embedding espliciti |
| `RetrievalResult` | provenance, score e reason code |
| `Budget` | massimo, riserva e margine di sicurezza |
| `TokenEstimator` | metodo cancellabile e versionato |
| `ContextSection` | testo selezionato con origine e costo |
| `ContextBundle` | snapshot immutabile delle sezioni budgetate |
| `BuildRequest` | richiesta validata al builder |

Semantic retrieval richiede `EmbeddingTarget`; un target non è accettato senza
il metodo semantic. Il costruttore del bundle compone `ErrBudgetExceeded` e
`ErrInvalidBundle`, preservando entrambe le classificazioni con `errors.Is`.

## Engine

`Engine` espone registrazione pre-uso di source, analyzer ed estimator, indexing,
lettura snapshot, retrieval e build. Il contratto non include lifecycle
duplicato, provider registry, persistence API o tool execution.

---

# Matrice di compatibilità

| Package esistente | Modifica | Esito |
|---|---|---|
| `pkg/runtime` | nessuna | Compatibile |
| `pkg/gestor` | nessuna | Compatibile |
| `pkg/provider` | nessuna | Compatibile |
| `pkg/plugin` | nessuna | Compatibile |
| `pkg/plugin/laravel` | nessuna nella Fase 1 | Compatibile |
| composition root `maestro` | nessuna nella Fase 1 | Compatibile |

L'accessor del composition root e il contratto Laravel verranno aggiunti nella
Fase 6 soltanto dopo che l'implementazione concreta dimostrerà il consumer.

---

# Errori pubblici

I sentinel distinguono input invalidi, registrazione, lookup, ambiguità,
supporto, limiti, failure delle estensioni, embedding e budget. Le
implementazioni devono usare wrapping `%w` e preservare `context.Canceled` e
`context.DeadlineExceeded`.

Gli errori non autorizzano la copia di path assoluti, testo o messaggi esterni
negli eventi. Le stringhe restituite direttamente al chiamante possono
descrivere l'input locale necessario alla diagnosi.

---

# Rischi evolutivi controllati

- i costruttori evitano struct literal dipendenti dalla rappresentazione
  privata per workspace, documenti, analisi, snapshot, query e bundle;
- nuovi campi opzionali possono essere aggiunti tramite options struct;
- enum possono essere estesi in modo additivo, ma metodi sconosciuti restano
  invalidi finché non implementati;
- `Engine` è un contratto nuovo: prima dell'integrazione finale sarà auditato
  di nuovo per evitare metodi non necessari;
- `ScanPolicy` è una struct pubblica e richiede compatibilità additiva;
- `ContextSection` e `RetrievalResult` sono value DTO e richiedono audit prima
  di modifiche semantiche.

---

# Verifica

- assertion di compilazione per tutte le interfacce;
- validazione di ID, root, path, policy e metadata;
- copie difensive per workspace, query, snapshot e bundle;
- digest deterministico dei documenti;
- ordinamento snapshot e lookup;
- riferimenti strutturali invalidi;
- target semantic obbligatorio;
- composizione dei sentinel sul budget.

La suite `go test ./pkg/contextengine` è verde.
