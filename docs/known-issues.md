# Maestro v0.2.0 Known Issues and Limitations

## Modello e prestazioni

- `llama3.1:8b` è generativo: formulazione, token e latenza variano fra run.
- Su Intel Core i5-8365U CPU-only le prove possono richiedere da pochi secondi
  a diversi minuti, secondo lo stato del modello e dell'host.
- Una pseudo-tool-call JSON valida, anche incorporata in testo esplicativo e
  anche se nomina un tool inesistente, non viene accettata come risposta
  finale: Maestro richiede un'invocazione dichiarata nel turno successivo,
  entro gli stessi hard limit. Per una richiesta stretta
  `Read <logical-path> ...`, il reference agent esegue invece una read
  verificata tramite Tool Runtime prima della prima inferenza. Arguments
  invalidi di un tool read-only sono recuperabili in modo redatto; failure di
  esecuzione, tool mutanti e altri errori restano terminali e non esiste retry
  implicito generalizzato.
- Il reason sintetico finale di alcuni hard limit è `execution_failed`, mentre
  l'evento terminale precedente resta `limit_exceeded` e autorevole.
- Sui workspace reali `llama3.1:8b` può alternare una run valida a
  `tool_failure` per arguments malformati o a `provider_failure` quando una
  singola inferenza supera il timeout; il runtime resta fail-closed e non
  applica retry impliciti.
- La Field Validation della Milestone 13 ha chiuso la matrice a 5/22 per stop
  rule con decisione `adoption_no_go_on_reference_profile`: 3/5 run ufficiali
  sono terminate in `provider_failure` e le due completion sono state
  classificate `partial` e `incorrect`. Aumentare il timeout ha migliorato la
  completion, non la qualità multi-file.
- La progressive choreography development-only impedisce finalizzazioni senza
  evidenza, ma `llama3.1:8b` non ne segue la progressione. `qwen3.5:9b` supera
  tool calling diretto e smoke provider, ma non converge sul fixture sintetico
  osservato ed è `candidate_rejected`. Nessuno dei due esiti amplia il percorso
  supportato o produce `v0.2.1`.
- Il confronto direct/chat della Milestone 13 mostra una risposta single-file
  corretta quando il file è allegato direttamente, ma latenze superiori a sei
  minuti e un timeout nel percorso Maestro pre-caricato. La Milestone 14 ha
  implementato `maestro chat` come superficie development-only separata e ha
  superato la matrice deterministica; il preflight live ha però trovato Ollama
  non attivo, quindi C0-C4 sono `not_run` e l'esito è
  `direct_chat_deferred`. Non costituisce supporto v0.2.0.

## Supporto ristretto

- Solo Linux `amd64` è qualificato.
- Solo Ollama con `llama3.1:8b` è qualificato per il reference agent.
- `embeddinggemma:latest` è la fixture embedding, ma il quick start usa
  retrieval lessicale.
- llama.cpp è sperimentale/non supportato; le prove router mode sull'hardware
  target hanno causato OOM e non costituiscono una matrice valida. Il preflight
  conclusivo non dispone di server, endpoint o profilo single-model e chiude
  la Milestone 3 con esito non supportato, non con un PASS.
- Modelli `rnj-1:8b-instruct-q4_K_M`, `ibm/granite4.1:8b`, `qwen3:8b` e
  `qwen2.5-coder:7b` non sono fixture del percorso supportato.

## Autorità e isolamento

- Non esiste sandbox o isolamento di processo.
- I plugin e tool built-in sono trusted in-process e usano i privilegi
  dell'utente.
- Tool mutanti e approval mutativa sono implementati ma non supportati; la
  configurazione ufficiale non li registra.
- Non esistono rollback generale, memoria persistente o recovery dopo restart.

## Controlled Mutation candidato

- `configs/maestro.mutating.example.yaml` è un profilo di qualificazione
  opt-in, non una configurazione supportata dalla v0.2.0.
- Soltanto `workspace.patch` su un file PHP esistente sotto `app/` appartiene
  al candidato; `workspace.write`, creazione, delete, rename e multi-file non
  sono qualificati.
- Il commit atomico qualificabile è implementato per Linux. Il fallback su
  altre piattaforme fallisce chiuso.
- Non esiste rollback dopo il rename. Un failure di sync directory,
  cancellazione o failure di reindex può lasciare il file applicato con stato
  durable/fresh incompleto; la run termina senza testo finale.
- Il profilo Ollama `ibm/granite4.1:8b` e il lower bound hardware restano
  candidati non supportati. La Milestone 11 ha superato matrice deterministica
  15/15 e preflight, ma Gate A è fallito al primo tentativo per arguments patch
  non esatti. Gate B/C non sono stati eseguiti per fail-fast; ADR-0032 registra
  `mutation_deferred`.

## Prodotto ed ecosistema

- CLI, config schema e API Go sono sperimentali durante la serie 0.x.
- Non esistono installer di sistema, auto-update o service unit.
- Maestro non installa Ollama, modelli, PHP, Composer o dipendenze Laravel.
- Nessun packaging generalizzato di plugin/tool di terze parti.
- Nessun multi-agent, remote execution, shell, Git o Docker completi.
- Nessuna selezione automatica di provider o modello.

Vedere `compatibility.md` per la matrice autorevole e `troubleshooting.md` per
le azioni operative. Il report decisionale completo è in
`reports/milestone-13-field-validation.md`; il confronto per modalità è in
`reports/milestone-13-direct-chat-diagnostic.md`.
