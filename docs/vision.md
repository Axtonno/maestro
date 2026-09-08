# Maestro Vision

Versione: 0.5.0

Stato: Current

Ultimo aggiornamento: 2026-09-08

## Punto di partenza

Maestro non parte più soltanto da un'architettura promettente. v0.5.0 ha un
nucleo operativo dimostrato: Direct Chat locale e Controlled Mutation
host-bound con approvazione umana.

La visione cresce da questo punto senza confondere tre livelli:

1. capacità disponibili nella release pubblica;
2. componenti sperimentali presenti nel repository;
3. direzioni future della roadmap.

## Oggi

Maestro funziona come una workstation locale delimitata:

- l'utente sceglie il workspace;
- la chat usa zero o un file esplicito;
- il write-mode usa un file PHP e un intervallo esplicito;
- il modello propone, ma non approva;
- Maestro mostra e verifica l'effetto prima di scrivere;
- ogni claim operativo è legato ad artifact, modello, digest e ambiente provati.

## Direzione

L'obiettivo di lungo periodo resta una workstation AI locale più completa,
capace di comporre contesto, provider, framework e strumenti con policy
verificabili.

Possibili estensioni includono:

- contesto e retrieval multi-file;
- altri linguaggi e framework;
- provider e profili hardware ulteriori;
- workflow multi-step e plugin;
- memoria e sessioni;
- agent con authority graduata.

Questi elementi non sono supportati finché una release e la relativa matrice
non li qualificano.

## Metodo di avanzamento

Ogni nuova capacità deve partire piccola, dichiarare il proprio confine,
fallire chiusa e produrre evidenza ripetibile. L'autonomia cresce soltanto dopo
il controllo dell'effetto, non prima.

La roadmap può cambiare; il support claim corrente è sempre definito dalla
[Compatibility Matrix](compatibility.md).

## Posizionamento

Maestro vuole diventare il livello locale che trasforma modelli generativi in
strumenti di sviluppo controllabili:

```text
sviluppatore
  -> workstation Maestro
  -> capacità e policy
  -> provider locale
  -> modello
```

Il valore non è nascondere il modello, ma rendere espliciti contesto,
authority, preview, fallimenti e responsabilità.

## Motto

> Build small. Orchestrate everything.
