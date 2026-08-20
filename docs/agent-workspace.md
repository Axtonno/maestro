# Agent Workspace Integration

Versione: 0.1.0

Stato: Implementato

Data: 2026-08-12

---

# Workspace binding

Una `RunRequest` può includere un `Workspace` immutabile il cui ID deve
coincidere con il target esplicito della request. L'Agent Runtime associa quel
workspace al `RunID` soltanto per la durata del coordinatore; il registry
condiviso dai reference tool risolve la root localmente.

Il modello riceve path logici, evidenza e provenance del Context Engine. Non
riceve né sceglie la root assoluta. Il binding viene rimosso a fine run e un
secondo binding sullo stesso run viene rifiutato.

# Reference workspace tool

`internal/tool.NewWorkspaceTools` costruisce cinque tool trusted in-process:

| Tool ID | Nome provider | Effetto |
|---|---|---|
| `workspace.list` | `workspace_list` | `workspace.inspect` |
| `workspace.read` | `workspace_read` | `workspace.inspect` |
| `workspace.search` | `workspace_search` | `workspace.inspect` |
| `workspace.write` | `workspace_write` | `workspace.mutate` |
| `workspace.patch` | `workspace_patch` | `workspace.mutate` |

Arguments JSON vengono decodificati in modo strict e normalizzati durante
`Prepare`; le action usano path logici e Workspace ID esatti. Listing e search
sono ordinati/bounded, read restituisce contenuto e SHA-256, mentre write e
patch restituiscono il nuovo digest.

# Containment e symlink

I path assoluti, `..`, backslash e forme non normalizzate vengono rifiutati.
La root deve essere una directory fisica. Ogni componente esistente viene
controllato con `Lstat` e qualsiasi symlink, anche interno, viene rifiutato.

Le operazioni usano `os.Root`, che mantiene il boundary del filesystem anche
in presenza di rename concorrenti. I controlli vengono ripetuti durante
`Execute`; una permission su un path non si estende per prefisso ad altri
path.

# Mutazioni e precondizioni

`workspace.write` richiede un digest SHA-256 esatto oppure `absent` per una
creazione `O_EXCL`, ma resta fuori dal profilo Controlled Mutation.
`workspace.patch` richiede digest esatto e una sola occorrenza del testo da
sostituire. Su Linux scrive un temporaneo nella stessa directory, preserva i
permessi, esegue `fsync`, rivalida inode e contenuto e usa `renameat` come unico
punto di commit atomico seguito dal sync della directory. Mismatch o
occorrenza ambigua producono `ResultFailed/precondition_failed` senza
overwrite; un failure dopo il rename dichiara esplicitamente `applied: true`.

Contenuto, dimensioni, UTF-8, item e output rispettano sia la ScanPolicy del
workspace sia gli ExecutionLimits del Tool Runtime.

# Freshness

Il Tool Runtime notifica l'inizio di `Execute` fuori lock e senza output.
Quando una action `workspace.mutate` inizia, la sessione diventa stale anche se
l'esito successivo è fallito o ambiguo. Un deny di policy non avvia Execute e
non marca stale.

Dopo una mutazione conclusa, il loop esegue un checkpoint esplicito:

```text
mutation start -> stale -> Context Engine.Index -> Context Engine.Build -> fresh
```

La nuova generazione deve essere strettamente maggiore e coincidere tra
snapshot indicizzato e bundle. Se refresh o build falliscono, la sessione
rimane stale sulla precedente generazione e nessuno snapshot parziale diventa
corrente.

Una run può iniziare al massimo un tentativo mutativo. Un esito non successful
è terminale e non viene restituito al modello per un retry; un secondo tentativo
nella stessa run viene rifiutato prima di `Invoke`. Soltanto apply riuscito,
`Index`, `Build` e `markFresh` riusciti consentono un nuovo turno e quindi una
risposta finale. Gli esiti pubblicano stati redatti `apply` e `reindex`, inclusi
commit applicato, durability e nuova generazione, senza path o contenuto.

# Framework neutrality

I tool dipendono esclusivamente da `contextengine.Workspace`. Il percorso di
test usa anche il `WorkspaceProvider` del plugin Laravel, senza importare
conoscenza Laravel nell'Agent Runtime o nel Tool Runtime.
