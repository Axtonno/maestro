# Maestro Laravel Plugin

Versione: 0.2.0

Stato: Implementato

Ultimo aggiornamento: 2026-08-11

---

# Scopo

Il plugin Laravel è il primo adapter framework-aware di Maestro e valida il
Plugin Runtime con un caso d'uso reale senza introdurre logica Laravel nel core.

La facade pubblica vive in `pkg/plugin/laravel`; detection e stato interno
restano in `internal/plugin/laravel`.

---

# Utilizzo

```go
rt := maestro.New()

err := rt.Plugins().RegisterLoader(
    laravel.ID,
    laravel.NewLoader(laravel.Config{Root: "/workspace/app"}),
)

_, err = rt.Plugins().Load(ctx, laravel.ID)
err = rt.Start(ctx)
```

`laravel.New` permette anche la registrazione diretta di un'istanza.

Il plugin dichiara `plugin.CapabilityWorkspaceDetection`, con ID stabile
`plugin.workspace-detection`. Gestor la indicizza attraverso il Registry
globale dei componenti senza eseguire la detection.

---

# Detection

Durante `Initialize` il plugin verifica che la root configurata contenga:

* un file regolare `artisan`;
* un `composer.json` valido;
* una dipendenza non vuota `laravel/framework` nella sezione `require`.

Il manifest Composer viene letto con un limite di 1 MiB. Il constraint trovato,
per esempio `^12.0`, è esposto tramite `FrameworkVersion`. La root viene
normalizzata come percorso assoluto.

La capability `Health` ripete la detection e segnala se il workspace non è più
riconoscibile come applicazione Laravel.

`Root` è immutabile dopo la costruzione. `FrameworkVersion` è vuota prima di
`Initialize` e viene sostituita atomicamente soltanto dopo una detection
riuscita. `Health` non modifica lo snapshot; health o initialize fallite
conservano l'ultima versione valida. Letture, initialize e health concorrenti
sono thread-safe.

Una root inesistente non rende invalida la configurazione in costruzione:
l'esistenza e i marker appartengono alla detection lifecycle e producono
`ErrNotDetected` durante `Initialize`.

---

# Errori

Gli errori pubblici distinguono:

* `ErrInvalidConfig`: root vuota o non risolvibile;
* `ErrNotDetected`: marker Laravel mancanti;
* `ErrInvalidComposerManifest`: JSON invalido o oltre il limite.

Gli errori restano ispezionabili con `errors.Is`.

---

# Limiti

Questa prima versione esegue discovery e health del framework. Comandi Artisan,
analisi delle route, service container, Eloquent, Blade e integrazioni con il
Context Engine verranno aggiunti come capability del plugin, non del core.

Non viene ancora esposto un contratto workspace framework-neutral: i metodi
concreti `Root` e `FrameworkVersion` restano il confine Laravel finché il Context
Engine non definirà requisiti comuni verificabili.
