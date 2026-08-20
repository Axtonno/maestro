# Milestone 10 — Report Fase 4

Data: 2026-08-20

Stato: **COMPLETATA — commit filesystem atomico e fault matrix verde**

## Risultato

`workspace.patch` non tronca più il file esistente. Su Linux prepara un file
temporaneo non seguibile nella stessa directory, ne preserva i permessi,
sincronizza il contenuto, rivalida la precondizione e sostituisce il target con
un `renameat` atomico relativo al descriptor della directory.

## Sequenza di commit

```text
open parent -> open/read target -> validate -> create temp -> write -> chmod
-> fsync temp -> reopen/read target -> revalidate -> renameat -> fsync parent
```

Il secondo open verifica inode, digest, singola occorrenza e contenuto proposto
immediatamente prima del rename. Il parent viene aperto tramite `os.Root`,
confrontato con `Lstat` prima e dopo l'open e poi mantenuto tramite descriptor.
Target e temporaneo usano `O_NOFOLLOW`; il temporaneo usa anche `O_EXCL` e un
nome random locale alla directory.

## Punto di commit ed esiti

- prima di `renameat`: qualsiasi failure lascia il target byte-identico e
  tenta la rimozione del temporaneo;
- `renameat` riuscito: la patch è applicata e non viene eseguito rollback;
- `fsync` della directory fallito: risultato `post_commit_sync_failed` con
  `applied: true` e `durable: false`;
- cancellazione osservata dopo il rename: risultato `post_commit_canceled` con
  stato applicato e durability registrata;
- precondizione o inode cambiati: `precondition_failed`, nessun overwrite.

`workspace.write` resta sperimentale e fuori scope; la sua implementazione non
viene qualificata dalla strategia atomica di `workspace.patch`.

## Fault injection

La seam interna copre:

- open parent e target;
- read iniziale e di recheck;
- creazione, write e chmod del temporaneo;
- sync del file;
- rename;
- sync della directory;
- cleanup;
- sostituzione concorrente del target;
- cancellazione immediatamente dopo il rename.

I test concorrenti leggono ripetutamente il target durante il commit e
osservano soltanto il contenuto completo precedente o successivo. Il successo
usa un inode nuovo e conserva i permessi `0640` della fixture.

## Gate

| Verifica | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |
| Cross-compile test package Darwin fallback | PASS |
| Fault pre-commit lascia target invariato | PASS |
| Lettura concorrente senza contenuto parziale | PASS |

La Fase 4 è completata. La Fase 5 può rendere espliciti gli stati
proposal/approval/apply/reindex e impedire retry o final success dopo un esito
mutativo non riuscito.
