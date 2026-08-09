# Milestone 4 — Report Fase 3

Fase: Discovery sources

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Popolare lo Snapshot Registry di Gestor attraverso il Registry globale dei
componenti e il Provider Runtime autorevoli, preservando dichiarazione,
availability, identità e cancellazione senza duplicare cataloghi o introdurre
probe durante la futura risoluzione.

---

# File e contratti introdotti

`internal/gestor/runtime_source.go` introduce:

- `RuntimeComponentSource`;
- collaboratore interno minimale `Components()`;
- mapping dei lifecycle Runtime negli ID Gestor `runtime.*`;
- supporto alle capability custom già namespaced;
- descriptor component-scoped con availability `unknown`.

`internal/gestor/provider_source.go` introduce:

- `ProviderCapabilitySource`;
- collaboratore interno additivo `Registered()` + `Capabilities()`;
- introspection adapter e instance per ogni provider registrato;
- introspection di model target espliciti e exact;
- mapping completo delle undici capability Provider note;
- traduzione conservativa di support e availability.

`internal/provider/runtime.go` aggiunge al concrete Runtime interno:

- `Registered()`, listing ordinato con copia difensiva;
- `SetRegistrationInvalidator(func())`, hook interno eseguito fuori lock dopo
  registrazioni riuscite.

`internal/runtime/runtime.go` coordina lo stesso invalidatore per registrazioni
di componenti, plugin e provider. Nessuno di questi metodi estende le
interfacce pubbliche `runtime.Runtime`, `provider.Runtime` o `plugin.Runtime`.

Le fixture e i test sono in:

- `internal/gestor/runtime_source_test.go`;
- `internal/gestor/provider_source_test.go`;
- test d'integrazione aggiuntivi dei Runtime Core e Provider.

---

# Runtime component source

La sorgente legge una copia della lista dal Registry autorevole, acquisisce i
metadata e ordina per component ID. Non conserva componenti e non esegue
lifecycle.

Regole di mapping:

- `configure` → `runtime.configure`;
- `initialize` → `runtime.initialize`;
- `start` → `runtime.start`;
- `stop` → `runtime.stop`;
- `reload` → `runtime.reload`;
- `health` → `runtime.health`;
- capability custom già namespaced → ID invariato.

Ogni descriptor usa `kind=component`, `scope=component`, source
`runtime.components` e availability `unknown`. Componenti senza capability non
producono descriptor.

I plugin non hanno una sorgente dedicata. Un plugin registrato dal Plugin
Runtime entra nel Registry globale dei componenti ed è visibile una sola volta
attraverso `RuntimeComponentSource`.

---

# Provider capability source

La sorgente legge gli ID registrati dal concrete Provider Runtime in ordine
lessicografico. Per ogni provider interroga:

1. target `adapter`;
2. target `instance`;
3. eventuali target `model` configurati, ordinati per model ID esatto.

I model target vengono copiati e validati dal costruttore. La sorgente non usa
il modello predefinito, non deduce modelli dal nome e non esegue model discovery
implicita.

Ogni report deve mantenere identità, target, model ID, numero e ordine canonico
delle capability. Le capability `unsupported` con availability `unavailable`
vengono omesse perché non costituiscono una dichiarazione Gestor. Le capability
`supported` producono descriptor mantenendo senza promozioni:

- `unknown` → `unknown`;
- `available` → `available`;
- `unavailable` → `unavailable`.

Gli errori Provider e la cancellazione vengono restituiti con la causa
originale. Quando la sorgente opera dentro lo Snapshot Registry, un errore
impedisce la pubblicazione dell'intero refresh.

---

# Invarianti implementati

- Runtime Registry e Provider Runtime restano le sole fonti autorevoli delle
  rispettive registrazioni.
- La discovery non conserva istanze in Gestor.
- Plugin e normali componenti condividono un unico percorso di indicizzazione.
- Descriptor Runtime dichiarati non vengono promossi ad available.
- Supporto Provider e availability restano assi distinti.
- Model ID e scope vengono conservati esattamente dal report richiesto.
- Provider, componenti, capability e model target duplicati o invalidi vengono
  rifiutati.
- L'ordine dei descriptor è deterministico indipendentemente dall'ordine dei
  cataloghi autorevoli.
- `Discover` verifica il context tra target e componenti e non pubblica stato.
- Errori o cancellazione di una sorgente mantengono l'ultimo snapshot valido.
- Registrazioni riuscite invalidano Gestor fuori dai lock; registrazioni
  fallite non invalidano.
- Nessuna interfaccia pubblica Runtime, Provider o Plugin è cambiata.

---

# Test eseguiti

Comandi del gate:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
```

Esito: tutti i comandi superati.

La suite copre:

- componenti con zero, una e più capability;
- mapping delle sei capability lifecycle;
- capability plugin namespaced;
- plugin registrato e scoperto una sola volta dal catalogo globale;
- dati Runtime nil, duplicati o invalidi;
- cancellazione e discovery Runtime concorrente;
- mapping di tutte le undici capability Provider note;
- adapter supported con availability unknown;
- instance available e unavailable;
- model available, unavailable e unknown;
- fixture Qwen `qwen2.5-coder:7b` con tool calling dichiarato ma unknown;
- fixture Llama `llama3.1:8b` con tool calling available;
- omissione delle capability unsupported;
- report Provider incoerenti, errori e cancellazione a metà introspection;
- conservazione dello snapshot completo su errore Provider;
- integrazione con il concrete Provider Runtime reale in-memory;
- composizione delle due sorgenti nello Snapshot Registry;
- listing Provider ordinato e con copia difensiva;
- invalidazione dopo registrazione di componenti, plugin e provider;
- callback di invalidazione reentrant a prova dell'esecuzione fuori lock;
- discovery concorrente e race detector repository-wide.

---

# Decisioni architetturali

- Le sorgenti dipendono da interfacce strutturali interne minime, non da
  implementazioni concrete.
- `Registered()` e l'hook invalidation sono additivi sul concrete Provider
  Runtime e restano invisibili attraverso `pkg/provider.Runtime`.
- I model target sono input esatti della sorgente, non configurazione pubblica
  o discovery automatica.
- Capability Runtime custom sono accettate soltanto se già namespaced e valide.
- Gestor non verifica nuovamente le interfacce lifecycle e non esegue codice
  risolto.
- Il composition root installerà sorgenti e invalidatore nella Fase 5; questa
  fase consegna i punti di integrazione e ne verifica il comportamento isolato.

---

# Limitazioni e attività rinviate

- Non esiste ancora Resolver: filtri, ambiguity, preferenze e dependency graph
  appartengono alla Fase 4.
- La vista read-only del dependency graph non è introdotta in questa fase.
- Il composition root non espone ancora Gestor e non registra automaticamente
  le sorgenti; appartiene alla Fase 5.
- La policy con cui la configurazione applicativa sceglierà i model target sarà
  coordinata nella Fase 5; la sorgente richiede già valori exact.
- Eventi, observer e refresh iniziale appartengono alla Fase 5.
- Non sono introdotti ranking, fallback, model discovery implicita o I/O in
  `Resolve`.

---

# Gate di uscita

Superato:

- entrambe le sorgenti producono descriptor ordinati e validi;
- declared, unknown, available e unavailable sono distinti dai test;
- plugin indicizzato senza duplicazione;
- errori e cancellazione impediscono snapshot parziali;
- invalidazione dopo registrazioni verificata fuori lock;
- interfacce pubbliche invariate;
- suite, race detector e vet repository-wide verdi.

La Fase 3 è completata. La Fase 4 — Resolver e dependency graph è pronta.
