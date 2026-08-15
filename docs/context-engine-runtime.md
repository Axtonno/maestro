# Context Engine Runtime Integration

Versione: 1.0.0

Stato: Implementato

Data: 2026-08-11

---

# Composition root

`maestro.New` compone una sola istanza del Context Engine e la espone con
`Runtime.ContextEngine`. L'istanza condivide il Provider Runtime per gli
embedding e l'Event Bus del Runtime per l'osservabilità. Non è un componente
lifecycle e non duplica registry, Gestor o configurazione provider.

Il contratto di basso livello `pkg/runtime.Runtime` resta invariato. L'accessor
appartiene alla facade `maestro.Runtime` insieme a `Gestor` e `Plugins`.

# Workspace provider Laravel

Laravel `0.3.1` implementa `contextengine.WorkspaceProvider` e dichiara la
capability `context.workspace-provider`. Il workspace generico contiene ID,
root normalizzata, source filesystem, policy di scansione e metadata limitati.

Gestor indicizza e risolve la capability, ma non invoca `Workspace`, non
indicizza file e non esegue analyzer. Il chiamante risolve il plugin e passa
esplicitamente il workspace al Context Engine.

# Eventi

| Topic | Quando |
|---|---|
| `context.index.started` | prima dell'indexing |
| `context.index.completed` | dopo la pubblicazione dello snapshot |
| `context.index.failed` | su failure o cancellazione |
| `context.build.started` | prima del build |
| `context.build.completed` | dopo la costruzione del bundle |
| `context.build.failed` | su failure o cancellazione |
| `context.cache.observed` | dopo index o build riusciti |

Il payload contiene soltanto workspace ID, generazione, conteggi, statistiche
aggregate della cache e un codice di failure. Non contiene root, document path,
query, testo, embedding, provider, modello o error string esterne.

La pubblicazione è sincrona e best-effort. Avviene fuori dai lock operativi;
errori e panic degli observer non cambiano un'operazione già committata. Un
subscriber lento applica la backpressure dell'Event Bus condiviso.

# Percorso pubblico

```go
plugin, err := laravel.New(root)
if err != nil {
    return err
}

runtime := maestro.New()
if err := runtime.Plugins().Register(plugin); err != nil {
    return err
}
if err := runtime.Start(ctx); err != nil {
    return err
}

workspace, err := plugin.Workspace(ctx)
if err != nil {
    return err
}
snapshot, err := runtime.ContextEngine().Index(ctx, workspace)
if err != nil {
    return err
}
_ = snapshot
```

# Invarianti operativi

- indexing, analyzer, provider ed eventi non sono invocati sotto lock globali;
- success e failure event sono mutuamente esclusivi per operazione;
- una cancellazione non pubblica snapshot o cache entry parziali;
- il bundle viene restituito soltanto al chiamante esplicito;
- memoria, tool permissions, watcher e persistenza restano fuori scope.
