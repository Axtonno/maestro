# Security Policy

## Supported version

La serie supportata iniziale è `0.1.x`. Le capacità dichiarate sperimentali o
non supportate nella compatibility matrix non ricevono una promessa operativa,
ma i problemi che compromettono il confine read-only o la sicurezza del core
sono comunque rilevanti.

## Reporting a vulnerability

Usare preferibilmente il canale privato “Report a vulnerability” nella sezione
Security del repository GitHub. Se il canale non è disponibile, aprire una
issue priva di dettagli sensibili chiedendo un contatto privato al maintainer.

Non pubblicare API key, workspace reali, prompt riservati, exploit completi o
altro materiale che possa danneggiare utenti. Includere privatamente versione,
commit, piattaforma, impatto, prerequisiti e una riproduzione minima redatta.

## Scope notes

Maestro v0.1.0 è trusted in-process e non dichiara una sandbox. L'esecuzione di
estensioni non attendibili con i privilegi dell'utente non è una vulnerabilità
rispetto al modello dichiarato. Sono invece in scope, tra gli altri:

- evasione del containment dei workspace tool built-in;
- mutazioni possibili dal profilo ufficiale read-only;
- bypass del permission model o dei hard limit;
- esposizione involontaria di secret in eventi o diagnostiche;
- archive ufficiali non corrispondenti a manifest o checksum.

Il modello completo è in `docs/security-model.md`.
