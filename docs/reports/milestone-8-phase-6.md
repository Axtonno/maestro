# Milestone 8 — Phase 6 Report

Data: 2026-08-15

Stato: **In corso — prima build finale rifiutata; hardening e nuovo gate
richiesti**

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

## Gate restante

Prima di chiudere la Fase 6 occorrono un nuovo commit sorgente, suite estesa,
nuova build riproducibile, installazione pulita e almeno due quick start
consecutivi corretti sul nuovo archive. Soltanto quel nuovo artifact potrà
essere associato al tag `v0.1.0`; questa prima build non deve essere rinominata
o pubblicata.
