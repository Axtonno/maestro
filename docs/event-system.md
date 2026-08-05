# Maestro Event System

Versione: 0.1.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-05

---

# Scopo

L'Event System permette ai componenti del Runtime di comunicare senza creare
dipendenze dirette tra loro.

Il contratto pubblico vive in `pkg/runtime`. L'implementazione concreta rimane
in `internal/runtime` ed è condivisa dal Runtime e dal `Context` consegnato ai
componenti.

---

# Contratto pubblico

Un evento espone:

* `Name() string`, usato come topic;
* `Payload() any`, che contiene i dati applicativi.

L'`EventBus` espone:

* `Publish(Event) error`;
* `Subscribe(topic string, Handler) error`;
* `Unsubscribe(topic string) error`.

I topic vuoti o composti soltanto da spazi non sono validi. La corrispondenza
dei topic è esatta e case-sensitive.

---

# Semantica di consegna

La consegna è sincrona.

`Publish` invoca tutti gli handler associati al topic nell'ordine in cui sono
stati registrati e ritorna soltanto quando l'intera consegna è terminata.

Se nessun handler è registrato, la pubblicazione termina senza errore.

Ogni pubblicazione usa uno snapshot degli handler. Una sottoscrizione o una
rimozione effettuata durante un callback sarà quindi visibile dalla
pubblicazione successiva, senza alterare quella già in corso.

`Unsubscribe` rimuove tutti gli handler del topic ed è idempotente per ogni
topic valido.

---

# Concorrenza e ownership

L'Event Bus protegge internamente la collezione dei subscriber ed è sicuro per
l'uso concorrente.

Gli handler vengono eseguiti senza mantenere lock interni. Possono quindi
pubblicare altri eventi, aggiungere sottoscrizioni o rimuovere un topic senza
causare deadlock nel bus.

Il bus non copia né modifica evento e payload. La sicurezza concorrente dei
dati contenuti nel payload appartiene al publisher e ai subscriber.

Gli handler sono eseguiti nel goroutine del publisher. Un handler lento applica
backpressure al publisher; l'eventuale esecuzione asincrona è responsabilità
esplicita del componente che ne ha bisogno.

Il bus non intercetta i panic degli handler. Il boundary che esegue codice non
fidato dovrà introdurre la propria strategia di isolamento.

---

# Errori

Gli errori pubblici dell'Event System sono:

* `ErrInvalidEvent`, per eventi nil o privi di un topic valido;
* `ErrInvalidSubscription`, per topic non validi o handler nil.

Gli errori vengono arricchiti con il contesto dell'operazione e restano
ispezionabili tramite `errors.Is`.

---

# Estensioni escluse dalla prima versione

La prima implementazione non introduce:

* wildcard o gerarchie di topic;
* code persistenti;
* retry;
* priorità degli handler;
* dispatch asincrono implicito;
* consegna distribuita.

Queste funzionalità potranno essere aggiunte soltanto in presenza di requisiti
concreti, senza appesantire il core preventivamente.
