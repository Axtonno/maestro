# Milestone 12 — Phase 1 Report

Data: 2026-08-21

Stato: **COMPLETATA**

## Esito

Il contratto della Milestone 12 è congelato come productization v0.2.0
esclusivamente read-only. Il support claim di partenza resta Linux `amd64`,
Ollama, `llama3.1:8b`, `embeddinggemma:latest`, reference agent Laravel e tool
list/read/search con `workspace_mutate: deny`.

ADR-0032 è vincolante: `workspace.write`, `workspace.patch`, approval mutativa,
profilo mutante e `ibm/granite4.1:8b` non possono entrare in configurazione
inclusa, quick start, archive o compatibility promise. Il file
`configs/maestro.mutating.example.yaml` resta una fixture sperimentale del solo
repository e non viene distribuito.

## Baseline verificata

Baseline: commit `2ddbb7bd850f25fb805775d82acaf57c831bd53d`.

| Controllo | Esito |
|---|---|
| `go test ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| doppio packaging v0.1.1 release | PASS, byte-identico |
| installazione e verifica archive v0.1.1 | PASS |

L'archive di baseline verificato è
`maestro-v0.1.1-linux-amd64.tar.gz`, SHA-256
`ff613958c5749e1d966f6b6b7eca92dfbe09419d7b2dff80beaa1dd453029734`. È
soltanto una prova del percorso di packaging esistente e non è un artifact
v0.2.0.

## Delta assegnati

| Delta | Fase owner |
|---|---:|
| documentazione, release notes e compatibility contract v0.2.0 | 2 |
| allowlist archive e test espliciti del confine read-only | 2 |
| packaging candidate riproducibile e installazione pulita | 3 |
| deny, EOF, no-TTY, segnali, deadline, hard limit e anti-leak | 4 |
| due quick start live consecutivi e RC immutabile | 5 |
| documentazione finale, artifact release e tag verificato | 6 |

## Decisioni di esecuzione

- le sei fasi sono sequenziali rispetto alla promozione;
- candidate, RC e release sono build distinte e immutabili;
- un failure live arresta la serie e richiede un nuovo candidate dopo la
  correzione;
- un prerequisito live assente viene registrato come `skipped`, non come PASS;
- il final artifact deve provenire da un commit successivo al congelamento
  della documentazione pubblica.

## Gate

**PASS.** La baseline è verde, il confine autorizzato è esplicito e ogni delta
ha una fase owner. La Fase 2 può iniziare; nessun artifact v0.2.0 è stato
prodotto e il support claim corrente non è cambiato.
