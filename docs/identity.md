# Maestro Identity

Versione: 0.5.0

Stato: Current

Ultimo aggiornamento: 2026-09-08

## Definizione

Maestro è una workstation AI locale per lo sviluppo software, costruita su un
runtime modulare e capability-based.

“Workstation” descrive ciò che l'utente può fare oggi: interrogare un workspace
locale e applicare una modifica piccola, esplicita e approvata. “Runtime”
descrive le fondamenta che separano provider, modelli, workspace, policy e
capacità e che rendono possibile l'evoluzione del prodotto.

Maestro non è un modello, un IDE o un agente autonomo general-purpose.

## Il nucleo operativo attuale

La release v0.5.0 fornisce due capacità:

- Direct Chat: una domanda senza contesto workspace oppure con un singolo file
  scelto dall'utente;
- Controlled Mutation: sostituzione di un intervallo di righe PHP sotto
  `app/`, con modello dedicato, preview, allow-once, stale check e apply
  atomico.

Questo nucleo è stato verificato dall'asset pubblico su WSL2/GPU e su Linux
nativo CPU-only. È un prodotto ristretto ma operativo, non soltanto una
proposta architetturale.

## Il problema che risolve oggi

Un modello locale da solo non definisce contesto, authority o confini della
scrittura. Maestro rende espliciti:

- quale workspace è autorizzato;
- quale file e quale intervallo sono coinvolti;
- quale modello serve ogni capacità;
- cosa viene mostrato prima di una scrittura;
- quando un cambiamento stale o non approvato deve fallire senza effetti.

L'intelligenza resta nel modello; fiducia, composizione e controllo
dell'effetto appartengono all'orchestrazione.

## Principi di prodotto

- locale per default;
- authority minima ed esplicita;
- nessun fallback silenzioso;
- una preview comprensibile prima di ogni write;
- failure chiusi e osservabili;
- claim limitati alle prove realmente superate;
- architettura estensibile senza presentare il futuro come già disponibile.

## Cosa Maestro non è ancora

Agent autonomi, retrieval multi-file, tool calling general-purpose, altri
provider/modelli e workflow multi-step sono direzioni successive. Il codice o
il design di queste capacità nel repository non equivale a supporto di
prodotto.

La separazione autorevole fra oggi e futuro è in
[Compatibility Matrix](compatibility.md) e [Roadmap](roadmap.md).

## Missione

Offrire agli sviluppatori un ambiente AI locale controllabile, verificabile e
progressivamente estensibile, nel quale il modello non riceva più authority di
quella necessaria al compito.

## Motto

> The intelligence is in the orchestration.
