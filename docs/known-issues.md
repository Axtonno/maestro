# Maestro v0.1.0 Known Issues and Limitations

## Modello e prestazioni

- `llama3.1:8b` è generativo: formulazione, token e latenza variano fra run.
- Su Intel Core i5-8365U CPU-only le prove hanno richiesto da circa un minuto a
  oltre cinque minuti.
- Una tool call malformata può terminare `tool_failure`; non esiste correzione o
  retry implicito generalizzato.
- Il reason sintetico finale di alcuni hard limit è `execution_failed`, mentre
  l'evento terminale precedente resta `limit_exceeded` e autorevole.

## Supporto ristretto

- Solo Linux `amd64` è qualificato.
- Solo Ollama con `llama3.1:8b` è qualificato per il reference agent.
- `embeddinggemma:latest` è la fixture embedding, ma il quick start usa
  retrieval lessicale.
- llama.cpp è sperimentale/non supportato; le prove router mode sull'hardware
  target hanno causato OOM e non costituiscono una matrice valida.
- Modelli `rnj-1:8b-instruct-q4_K_M`, `ibm/granite4.1:8b`, `qwen3:8b` e
  `qwen2.5-coder:7b` non sono fixture del percorso supportato.

## Autorità e isolamento

- Non esiste sandbox o isolamento di processo.
- I plugin e tool built-in sono trusted in-process e usano i privilegi
  dell'utente.
- Tool mutanti e approval mutativa sono implementati ma non supportati nella
  v0.1.0; la configurazione ufficiale non li registra.
- Non esistono rollback generale, memoria persistente o recovery dopo restart.

## Prodotto ed ecosistema

- CLI, config schema e API Go sono sperimentali durante la serie 0.x.
- Non esistono installer di sistema, auto-update o service unit.
- Maestro non installa Ollama, modelli, PHP, Composer o dipendenze Laravel.
- Nessun packaging generalizzato di plugin/tool di terze parti.
- Nessun multi-agent, remote execution, shell, Git o Docker completi.
- Nessuna selezione automatica di provider o modello.

Vedere `compatibility.md` per la matrice autorevole e `troubleshooting.md` per
le azioni operative.
