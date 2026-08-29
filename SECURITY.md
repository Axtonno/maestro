# Security Policy

## Supported version

La release pubblicata corrente è `0.2.0`. Il packaging candidate `0.3.0`
riceve copertura per il confine Direct Chat durante la qualificazione, ma non
diventa release supportata prima del verdetto finale. Le capacità dichiarate
non supportate nella compatibility matrix non ricevono una promessa operativa;
i problemi che compromettono containment, redazione o read-only restano
comunque rilevanti.

## Reporting a vulnerability

Usare preferibilmente il canale privato “Report a vulnerability” nella sezione
Security del repository GitHub. Se il canale non è disponibile, aprire una
issue priva di dettagli sensibili chiedendo un contatto privato al maintainer.

Non pubblicare API key, workspace reali, prompt riservati, exploit completi o
altro materiale che possa danneggiare utenti. Includere privatamente versione,
commit, piattaforma, impatto, prerequisiti e una riproduzione minima redatta.

## Scope notes

Maestro è trusted in-process e non dichiara una sandbox. L'esecuzione di
estensioni non attendibili con i privilegi dell'utente non è una vulnerabilità
rispetto al modello dichiarato. Sono invece in scope, tra gli altri:

- evasione del containment del file Direct Chat o dei workspace tool storici;
- mutazioni possibili dal profilo Direct Chat o da un profilo ufficiale read-only;
- bypass dei limiti, del confine tool-free o dell'output atomico;
- esposizione involontaria di secret in eventi o diagnostiche;
- archive ufficiali non corrispondenti a manifest o checksum.

Il modello completo è in `docs/security-model.md`.
