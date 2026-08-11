# Milestone 5 — Report Fase 5

Fase: Osservabilità, hardening e gate finale

Stato: Completata

Data: 2026-08-11

---

# Obiettivo

Chiudere la Milestone 5 verificando cardinalità, ordine, failure isolation e
payload degli eventi plugin, ripetendo l'audit di compatibilità ed eseguendo i
gate repository-wide.

---

# Eventi plugin

Restano pubblici tre topic stabili:

```text
plugin.loader.registered
plugin.registered
plugin.loaded
```

Una registrazione loader riuscita emette il primo evento. Un `Load` riuscito
emette `registered` e poi `loaded`, per una sequenza totale:

```text
loader.registered -> registered -> loaded
```

Registrazione diretta emette soltanto `registered`. Operazioni fallite non
emettono eventi di successo per stadi non completati.

La matrice negativa copre:

- errore factory;
- cancellazione dopo la factory;
- manifest incompatibile;
- rifiuto del registrar;
- registrazione diretta duplicata.

In tutti i casi la cardinalità dei success event successivi è zero.

---

# Isolamento e backpressure

La pubblicazione plugin è ora best-effort rispetto al risultato dell'operazione
già completata:

- un errore restituito dall'Event Bus viene ignorato;
- un panic del publisher o di un subscriber viene recuperato al confine Plugin
  Runtime;
- stato, catalogo e registry già committati restano autorevoli;
- un observer non trasforma un successo in un errore apparente.

La consegna resta sincrona secondo ADR-0005. Un subscriber lento applica
backpressure: l'operazione ritorna soltanto dopo il callback. I test dimostrano
che, mentre il callback è bloccato, lo stato è già visibile e nessun lock del
Plugin Runtime impedisce letture o re-entrancy.

L'isolamento avviene al boundary Plugin Runtime e non cambia la semantica
generale dell'Event Bus, che continua a non recuperare autonomamente i panic.

---

# Payload e trust

`plugin.EventPayload` contiene soltanto:

- l'ID stabile del plugin;
- il riferimento `Plugin` per gli eventi successivi all'istanza, nil per la
  registrazione del loader.

Non vengono copiati nel payload configurazione, credenziali, error string,
contenuti o file del workspace. Il riferimento Plugin è intenzionalmente un
oggetto trusted in-process e non è un payload telemetrico serializzabile né un
confine di redazione per processi remoti. Adapter di logging o telemetria non
devono serializzarlo implicitamente.

Questa semantica conserva la compatibilità pubblica della baseline senza
presentare il riferimento in-process come barriera di sicurezza.

---

# Audit di compatibilità

L'audit finale in `docs/plugin-api-compatibility-audit.md` conferma:

- nessuna firma rimossa o modificata;
- `plugin.Runtime` invariato;
- manifest e API runtime ancora alla versione `1`;
- sentinel e topic invariati;
- nuova capability workspace aggiunta tramite costante;
- versione Laravel aggiornata a `0.2.0`;
- isolamento eventi implementato internamente senza cambiare il contratto;
- nessuna nuova dipendenza nel modulo Go.

---

# Verifica

Comandi del gate:

```text
GOCACHE=/tmp/maestro-m5p5-repeat go test -count=20 . ./internal/plugin ./internal/plugin/laravel ./pkg/plugin/laravel
GOCACHE=/tmp/maestro-m5p5-test go test ./...
GOCACHE=/tmp/maestro-m5p5-race go test -race ./...
GOCACHE=/tmp/maestro-m5p5-vet go vet ./...
git diff --check
```

Esito: tutti i comandi superati.

---

# Gate di uscita

Superato:

- sequenza e cardinalità degli eventi verificate;
- nessun success event su failure o cancellazione;
- callback re-entrant e bloccanti fuori lock;
- errori e panic di osservazione isolati dal risultato;
- payload e trust boundary documentati senza overclaim di redazione;
- audit API completato;
- suite ripetuta, completa, race detector e vet verdi;
- documentazione e report allineati.

La Fase 5 e il gate finale della Milestone 5 sono completati.
