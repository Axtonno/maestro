# Milestone 12 — Phase 6 Report

Data: 2026-08-21

Stato: **COMPLETATA**

## Catena di release

| Elemento | Identità |
|---|---|
| Freeze documentale | `fac2ae347d9fd6e03e9faef466d11bafa961370c` |
| Commit artifact | `5b05237362370fa79f133e159105a6a99050e81a` |
| Versione | `v0.2.0` |
| Stato manifest | `release` |
| Piattaforma | Linux `amd64` |
| Archive | `maestro-v0.2.0-linux-amd64.tar.gz` |
| SHA-256 | `c2d2a6f35178e91ad0c62d3c27f4ff2c33eedb46fd5fb327535890638e963758` |
| Tag annotato | `v0.2.0` |

Il commit artifact è pulito, distinto e discendente dal freeze documentale.
L'archive finale è distinto da `v0.2.0-rc.1`; non deriva da rename o
sovrascrittura del release candidate.

## Documentazione pubblica

Prima della build finale sono stati congelati README, changelog, note di
release, installazione, quick start, compatibility matrix, security model,
known issues, troubleshooting e contratto API. L'audit non rileva marker
`Unreleased`, stato candidate o pubblicazione subordinata residui.

Il support claim resta limitato a Linux `amd64`, Ollama, `llama3.1:8b`,
reference agent Laravel e list/read/search con `workspace_mutate: deny`.
Controlled Mutation, Granite, llama.cpp, sandbox, shell, Git, multi-agent e
tool di terze parti restano non supportati.

## Packaging e installazione finale

Il gate ha costruito due volte l'archive con input normalizzati. Archive e
checksum sono byte-identici. La build conservata riproduce lo stesso SHA-256.

L'installazione in una directory esterna al checkout conferma:

- nome archive, versione, commit, piattaforma e stato manifest esatti;
- `maestro version`, root help e `agents` funzionanti;
- `doctor` 9/9 contro Ollama locale;
- presenza del modello supportato tramite `models`;
- licenze, attribution, fixture e documentazione pubblica complete;
- assenza di profilo mutante, documentazione interna, path del checkout,
  credenziali-shaped, symlink o token irrisolti.

## Conferma live

Il quick start sull'esatto artifact finale termina `completed` in 241082 ms,
con un turno modello e una read. La risposta identifica semanticamente il
servizio e il metodo attesi. La response completa non è pubblicata nel report.

La fixture installata coincide byte-per-byte con la baseline dopo la run.
Nessuna approval o capability mutativa è stata osservata.

## Gate deterministici finali

| Controllo | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| syntax check script di packaging | PASS |
| `git diff --check` | PASS |
| worktree del commit artifact | pulito |

## Tag

`v0.2.0` è un tag annotato. Il commit risolto dal tag, quello dichiarato nel
manifest e quello stampato dal binario coincidono esattamente con
`5b05237362370fa79f133e159105a6a99050e81a`.

## Gate

**PASS.** Documentazione, artifact finale, installazione, conferma live,
anti-leak, suite e identità del tag sono verificati. La pubblicazione remota
non fa parte di questo gate e non è stata eseguita.
