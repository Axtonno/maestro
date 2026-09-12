# Roadmap pubblica

Questa pagina comunica direzioni di prodotto, non date, sequenze operative o
promesse di rilascio. Una capacità diventa supportata soltanto quando appare
nella [matrice delle capacità correnti](current-capabilities.md).

## Oggi

Maestro supporta Direct Chat e Controlled Mutation nel perimetro dichiarato
per v0.5.0 su Linux `amd64` con Ollama locale.

Il percorso di onboarding post-v0.5 ha superato un trial su Windows 11 → WSL2
con filesystem Linux e Ollama interno. Il candidate resta non pubblicato e il
risultato non equivale a supporto Windows nativo.

## In valutazione

- distribuzione e qualifica su Windows e macOS; dopo il trial WSL2, il target
  Windows ancora in valutazione è quello nativo;
- miglioramenti dell’esperienza editor mantenendo preview e approvazione nella
  superficie controllata.

## Non promesso

- modifiche multi-file;
- selezione autonoma di file o intervalli;
- agent autonomi e tool calling come superficie di prodotto;
- esecuzione automatica di shell, Git, build o test;
- supporto di provider e modelli non qualificati;
- un profilo single-model senza una nuova qualifica: il candidato già valutato
  ha conservato i gate mutativi, ma non ha raggiunto la qualità Direct Chat.

Le priorità possono cambiare sulla base di sicurezza, affidabilità e riscontri
degli utenti. I piani tattici e i report intermedi non fanno parte della
documentazione pubblica.
