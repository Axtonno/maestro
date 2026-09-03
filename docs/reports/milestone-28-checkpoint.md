# Milestone 28 – Checkpoint deterministico

Data: 2026-09-03

Stato: **IN CORSO** — `deterministic_compiler_implemented_not_yet_qualified`

È stato introdotto il contratto JSON v1 strict e il compilatore provider-neutral. Il compilatore ricalcola il digest della read autorevole, impone path identico e singola occorrenza esatta, rifiuta no-op e target sensibili, deriva contenuto post-modifica e fingerprint canonico senza effettuare I/O.

La nuova superficie candidata `workspace.replace` prepara preview, diff completa e digest pre/post dallo stesso candidato immutabile. L'esecuzione riusa containment, controllo anti-stale e commit atomico già presenti nel trusted filesystem layer; `workspace.patch` storico non viene promosso né reinterpretato.

La matrice unitaria copre proposta positiva deterministica, campi mancanti/sconosciuti/duplicati, trailing output, versione e operazione errate, mismatch di path, traversal, secret, digest stale, old text assente/ambiguo e no-op. Un test d'integrazione verifica che Prepare non muti il workspace e mostri il fingerprint vincolante.

Un toolchain Go 1.24.5 ufficiale scaricato in area temporanea e verificato con SHA-256 ha compilato il codice. I test del compilatore e il test d'integrazione di Prepare sono PASS; `go vet` sui package `internal/mutation` e `internal/tool` e `git diff --check` sono PASS. La suite Windows completa non è verde per failure baseline dipendenti dalla piattaforma (atomic replace Linux-only, symlink/Mkfifo/permission mode, fixture LF) e per il blocco degli eseguibili prodotti in `%TEMP%` da parte della policy applicativa; non sono failure attribuite come gate M28. Restano aperti suite/race Linux, matrice di approval e fault injection sulla nuova superficie, confronto dei due trasporti, qualificazione semantica/live e handoff v0.5.0.
