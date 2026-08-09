# Milestone 4 — Report Fase 5

Fase: Composition root, osservabilità e gate finale

Stato: Completata

Data: 2026-08-09

---

# Obiettivo

Comporre Gestor nel Runtime pubblico di Maestro, collegare sorgenti,
invalidazione, refresh iniziale ed eventi redatti e verificare l'intero percorso
in-memory senza modificare ownership, lifecycle o contratti del Runtime Core.

---

# File e contratti introdotti

`pkg/gestor/contracts.go` aggiunge `Service`, facade minimale che compone i
contratti `Registry` e `Resolver` già approvati.

`pkg/gestor/event.go` introduce i topic pubblici stabili:

- `gestor.refresh.started`;
- `gestor.refresh.completed`;
- `gestor.refresh.failed`;
- `gestor.resolution.completed`;
- `gestor.resolution.failed`.

Il relativo `EventPayload` contiene soltanto metadata redatti e categorie di
errore stabili.

`internal/gestor/service.go` delega Registry e Resolver e aggiunge la
pubblicazione best-effort degli eventi. `internal/runtime/runtime.go` compone
Registry Gestor, sorgenti Runtime/Provider, Resolver, vista del grafo e Service,
pubblica lo snapshot iniziale e installa l'invalidatore condiviso.

`maestro.Runtime` e l'interfaccia additiva `internal/runtime.Runtime` espongono
`Gestor() gestor.Service`. `pkg/runtime.Runtime` non cambia.

I test aggiunti sono in:

- `internal/gestor/service_test.go`;
- `maestro_test.go` per il percorso pubblico end-to-end.

---

# Bootstrap e policy di refresh

Ogni `maestro.New()` crea servizi Gestor isolati e registra automaticamente:

1. `runtime.components` sul Registry autorevole dei componenti;
2. `provider.capabilities` sul Provider Runtime autorevole.

Il bootstrap esegue un refresh deterministico e pubblica uno snapshot current a
generazione 1 anche con zero descriptor. Le registrazioni successive rendono lo
snapshot stale senza avviare discovery o I/O. Il chiamante richiede
esplicitamente un nuovo `Refresh`; per risolvere componenti il Runtime deve
avere costruito il dependency graph, normalmente tramite `Start`.

La sorgente Provider built-in interroga adapter e instance. Non inferisce model
ID dal default, dal nome o dal catalogo. Target model ulteriori richiedono una
`Source` esplicita con ID esatti.

---

# Osservabilità e redazione

Gli eventi sono pubblicati sullo stesso Event Bus esposto dal Runtime. Nessun
evento viene emesso mantenendo lock Gestor o Runtime.

Il payload può contenere:

- capability ID;
- generazione e conteggi dello snapshot;
- target kind e scope, ma non target ID o model ID;
- reason della risoluzione e numero di dipendenze;
- categoria di fallimento tipizzata.

Non esistono campi per error string, cause remote, source detail, prompt,
risposte, embedding, configurazioni o credenziali. La pubblicazione è
best-effort: errori e panic degli observer non cambiano il risultato già
prodotto. Un observer sincrono lento può aggiungere latenza, ma può leggere lo
snapshot già pubblicato e non blocca i lock Gestor.

---

# Invarianti implementati

- Runtime Core, Provider Runtime e Plugin Runtime conservano ownership e API.
- Componenti e plugin condividono ancora un solo Registry e dependency graph.
- Il bootstrap non esegue I/O live con cataloghi vuoti.
- Registrazioni riuscite invalidano dopo il rilascio dei lock; quelle fallite
  non cambiano lo snapshot.
- `Refresh` resta all-or-nothing e non viene avviato implicitamente dalle
  registrazioni.
- Gestor non esegue capability, lifecycle, probe o codice del candidato durante
  `Resolve`.
- Eventi e observer non possono trasformare un fallimento in successo o
  corrompere uno snapshot pubblicato.
- Plugin discovery attraversa soltanto la sorgente componenti globale.
- Il contratto `pkg/runtime.Runtime` resta compatibile e invariato.

---

# Test eseguiti

Comandi del gate:

```text
GOCACHE=/tmp/maestro-go-build go test ./...
GOCACHE=/tmp/maestro-go-build go test -race ./...
GOCACHE=/tmp/maestro-go-build go vet ./...
git diff --check
```

Esito: tutti i comandi superati.

La suite aggiuntiva copre:

- costruttore Service e collaboratori mancanti;
- eventi refresh started/completed/failed;
- eventi resolution completed/failed;
- classificazione stabile degli errori;
- assenza di error message e target identity nei payload;
- errori, panic e re-entry degli observer;
- observer lento mentre lo snapshot current rimane leggibile;
- bootstrap pubblico con zero componenti e provider;
- registrazioni multiple e invalidazione coordinata;
- refresh e resolve dal composition root;
- provider adapter/instance con availability distinta;
- component resolution con dipendenza richiesta e opzionale mancante;
- plugin visibile una sola volta;
- eventi sul bus condiviso;
- source failure senza pubblicazione parziale;
- lifecycle Runtime e shutdown invariati;
- race detector repository-wide.

---

# Decisioni architetturali

- Gestor è esposto dal composition root Maestro e non aggiunto a
  `pkg/runtime.Runtime`, evitando di ampliare il contratto del Core.
- `Service` incorpora Registry e Resolver invece di duplicarne i metodi o
  nascondere refresh e stale state.
- Il refresh iniziale è l'unico refresh automatico; ogni refresh successivo è
  esplicito.
- La registrazione invalida soltanto e non effettua introspection.
- Gli eventi usano l'Event Bus condiviso e una forma pubblica redatta per
  costruzione.
- Errori e panic degli observer sono isolati al confine del servizio; il bus
  resta sincrono e mantiene le semantiche esistenti.
- La policy built-in non inventa model target; l'estensione passa da `Source`.

---

# Limitazioni e attività rinviate

- Non esistono unregister, hot reload o persistenza degli snapshot.
- Non esistono discovery remota, marketplace o probe periodici.
- Non esistono ranking, fallback globali o preferenze implicite.
- Gli observer sincroni possono aggiungere latenza alla chiamata osservata.
- Model target built-in non sono configurati implicitamente; una sorgente
  esplicita deve dichiararli.
- Gestor restituisce metadata e dependency plan, non handle eseguibili.

Nessuna di queste limitazioni blocca il contratto consegnato dalla Milestone 4.

---

# Gate di uscita

Superato:

- composition root pubblico e contratti coerenti;
- bootstrap, invalidazione, refresh e resolve end-to-end verificati;
- eventi success/failure redatti e fuori lock;
- observer failure/panic isolati;
- lifecycle e Runtime specializzati invariati;
- cinque report di fase presenti;
- suite, race detector, vet e diff check repository-wide verdi.

La Fase 5 è completata e chiude la Milestone 4 — Gestor.
