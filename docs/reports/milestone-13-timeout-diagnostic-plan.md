# Milestone 13 — Timeout Diagnostic Plan

Data: 2026-08-21

Stato: completato; esito in `milestone-13-timeout-diagnostic.md`

## Scopo

Distinguere inferenza molto lenta, saturazione del contesto e mancata
convergenza del modello senza alterare o sostituire i risultati di Batch 1.
Tutte le prove di questo documento sono esplorative, non ufficiali e restano
fuori da completion rate, qualità e minimo delle 22 run.

## Invarianti

- Batch 1 resta immutabile e le sue run non vengono ripetute ufficialmente;
- artifact, provider e modello restano identificati esplicitamente;
- ogni profilo diagnostico usa un nuovo ID e non viene presentato come profilo
  supportato;
- tool e policy restano list/read/search con mutation deny;
- snapshot pre/post e anti-leak restano obbligatori;
- nessun risultato diagnostico modifica retroattivamente prompt, timeout,
  classificazioni o denominatori di Batch 1.

## Prove

1. raccogliere tempi Ollama di prompt evaluation e token generation quando
   disponibili, separandoli dalla durata end-to-end Maestro;
2. registrare dimensione del contesto, input/output token, tool result byte e
   numero di turni/call senza conservare contenuti nel report pubblico;
3. rieseguire una copia diagnostica di un task fallito con provider timeout
   ampliato da 5 a 10 minuti e limite run separato, per distinguere lento da
   non convergente;
4. ricostruire dagli eventi redatti perché il modello termina o resta in
   inferenza dopo la prima read, senza introdurre retry;
5. eseguire un task semplice sul secondo progetto con lo stesso profilo
   diagnostico per separare effetto progetto, task e host.

## Identità delle prove

Le prove usano ID `D13-*`, configurazioni e prompt con SHA-256 propri e una
directory evidenze distinta. Non usano gli ID `FV-*` come run ufficiali e non
occupano una ripetizione della matrice.

## Gate diagnostico

L'indagine termina con una delle classi:

- `slow_but_convergent`: completa entro il timeout ampliato con metriche
  Ollama coerenti;
- `context_pressure`: evidenza positiva che prompt evaluation o contesto
  domina il costo;
- `generation_pressure`: token generation domina il costo;
- `non_convergent`: non completa entro il bound ampliato;
- `unresolved`: osservabilità insufficiente per attribuire la causa.

Solo dopo il verdetto si decide se congelare un Batch 2. Un Batch 2 deve avere
nuovo `profile_id`, motivazione, configurazione e hash; non sostituisce Batch 1.
