# Milestone 8 — Phase 6 Report

Data: 2026-08-15

Stato: **Completata — artifact finale validato e tag `v0.1.0` creato**

## Baseline documentale

Il commit `6e867c13297c438874e0ecc2e1f334ba19fc7ab6` congela README e
documentazione pubblica intorno al solo percorso supportato: Linux `amd64`,
Ollama, `llama3.1:8b`, fixture embedding `embeddinggemma:latest` e reference
agent Laravel read-only. Mutazioni e llama.cpp restano
sperimentali/non supportati; il processo è trusted in-process e senza sandbox.

La build ripetuta dal medesimo commit è byte-identica e il verifier del
packaging è verde.

| Campo | Prima build finale |
|---|---|
| Artifact | `maestro-v0.1.0-linux-amd64.tar.gz` |
| Commit | `6e867c13297c438874e0ecc2e1f334ba19fc7ab6` |
| SHA-256 | `5ad3e297e28033868488c42a3ff58e47a44d393f6c830cc33085a461cc564124` |
| Dimensione | 3.606.307 byte |
| Manifest status | `release` |
| Verdetto | Rifiutato, non promuovibile |

## Gate pulito e causa del rifiuto

L'archive è stato verificato ed estratto in una directory temporanea senza
checkout. `maestro version`, help, `doctor` (9/9), `models` e `agents` hanno
superato il preflight.

Il primo quick start ha terminato `completed`, exit 0, con due turni, una
lettura reale e risposta corretta su `OrderService::create`; il digest della
fixture è rimasto invariato. Il secondo ha terminato exit 0 con un turno e zero
tool call, stampando una struttura JSON che nominava `workspace_read` invece
di invocare il tool. Non ha quindi risposto alla richiesta. Anche in questo
caso la fixture è rimasta invariata, ma il criterio semantico è fallito.

Il risultato non viene mascherato con una ripetizione opportunistica: la build
è conservata localmente come evidenza rifiutata e non può ricevere il tag
`v0.1.0`.

## Correzione richiesta

Il loop tratta ora una risposta JSON strutturata che descrive un tool
dichiarato come protocollo incompleto. Richiede al provider una vera
invocazione nel turno successivo, consumando gli stessi hard limit e senza
eseguire effetti implicitamente. Il controllo non si attiva per semplice
menzione del nome del tool in una risposta finale.

I test di regressione dimostrano entrambi i confini:

- pseudo-call testuale → correzione → tool call reale → risposta finale;
- risposta finale in prosa che nomina il tool → accettata senza turno extra.

## Artifact finale

L'hardening e le regressioni sono fissati nel nuovo commit sorgente
`f882919798fa6073bc11c6af18a431bf249a7755`. Il packaging è stato eseguito
due volte dallo stesso commit: gli archive risultano byte-identici e il
verifier conferma identità, manifest, documenti obbligatori, configurazione
read-only, fixture, assenza del path di build e assenza di dati con forma di
credenziale.

| Campo | Release finale |
|---|---|
| Artifact | `maestro-v0.1.0-linux-amd64.tar.gz` |
| Commit | `f882919798fa6073bc11c6af18a431bf249a7755` |
| SHA-256 | `c785676a177165a2c11ff0fc744931ac8b5d923466155ec32365e7a0c03d271f` |
| Dimensione | 3.604.828 byte |
| Manifest status | `release` |
| Tag annotato | `v0.1.0` |

## Prova da ambiente pulito

Soltanto archive e checksum sono stati copiati in
`/tmp/maestro-v010-final.1SbBgX`; il checkout non era disponibile dalla
directory estratta. Checksum, estrazione, `maestro version`, help, licenza e
manifest sono verdi. `version` restituisce `v0.1.0` e il commit esatto.

Il binario estratto supera:

- `doctor`: 9 pass, 0 failure;
- `models`: `llama3.1:8b` ed `embeddinggemma:latest` presenti;
- `agents`: `agent.reference` presente;
- scansione archive: nessun secret, path del builder o tool mutante nel
  profilo ufficiale.

## Quick start consecutivi

La stessa istruzione pubblica è stata eseguita due volte sul medesimo archive.

| Run | Terminale | Exit | Turni/tool | Durata | Risultato |
|---|---|---:|---:|---:|---|
| `run-39cd435cd0dae65878fa051d42286333` | `completed` | 0 | 2 / 1 | 316.768 ms | `OrderService::create`, corretto |
| `run-9f5b16663b77c93756bf09c2bad545a4` | `completed` | 0 | 2 / 1 | 14.131 ms | `OrderService::create`, corretto |

In entrambe le run il primo turno ha descritto la chiamata invece di usare il
canale tool. La nuova correzione protocollare è stata quindi esercitata live:
il testo non è stato accettato come successo, il secondo turno ha prodotto una
vera `workspace.read` e soltanto dopo è stata emessa la risposta finale.

Il digest aggregato della fixture prima e dopo ogni scenario è rimasto
`f5318094c365a9a634d4a983e86691ff5e84f83a96ba5b7f844818759847a250`.

## Controlli operativi finali

- hard limit `model_turns: 1`: `limit_exceeded`, exit 1, un turno, nessun
  risultato applicativo e workspace invariato;
- SIGINT: `canceled`, exit 130, terminale dopo 1.997 ms e quindi entro il
  budget di shutdown di 30 secondi;
- stderr operativo non espone prompt, contenuti della fixture, argomenti tool,
  root fisica, fingerprint o secret; stdout contiene soltanto il risultato
  applicativo intenzionale nelle run riuscite;
- checksum dell'archive ancora valido dopo il gate; modello Ollama scaricato
  dalla memoria al termine.

## Gate repository-wide

Sul commit sorgente sono verdi:

```text
go test -count=3 ./...
go test -race ./...
go vet ./...
go mod verify
git diff --check
```

`BenchmarkAgentLoopDeterministic` registra 36.057 ns/op, 17.262 B/op e 223
allocazioni/op sull'host di validazione.

## Verdetto

**Fase 6 completata. GO alla v0.1.0.**

L'archive rifiutato resta separato e non viene pubblicato. L'unico artifact
finale è quello associato al commit e checksum indicati sopra; il tag
`v0.1.0` punta esattamente al commit incorporato dal binario.
