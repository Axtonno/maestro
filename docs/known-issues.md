# Maestro v0.3.0 Known Issues and Limitations

## Qualità e prestazioni

- `qwen3.5:9b` è generativo: formulazione, token e latenza possono variare.
- La matrice qualitativa F6.4 ha ottenuto 4/5. Nel caso fallito il modello ha
  motivato un refactoring con conseguenze non dimostrate dal file; proposte e
  inferenze devono quindi essere riesaminate dall’utente.
- Temperatura zero riduce il drift fra complete e stream ma non rende la
  correttezza semanticamente deterministica.
- `num_ctx_effective` e `thinking_effective` possono apparire `unknown` quando
  l’adapter non restituisce un’attestazione per-run; doctor verifica comunque
  che i controlli richiesti siano supportabili prima della completion.
- Latenza e memoria dipendono dal provider, dall’hardware e dalla residenza del
  modello. Maestro non gestisce automaticamente load/unload.

## Supporto ristretto

- Solo Linux `amd64`, Ollama locale e l’identità modello/digest documentata
  appartengono alla matrice qualificata.
- È disponibile zero o un file esplicito; directory, glob, multi-file,
  retrieval e selezione automatica non sono supportati.
- Lo streaming è aggregato: non stampa chunk progressivi, preservando output
  atomico in caso di errore o cancellazione.
- Endpoint remoti, llama.cpp e altri modelli non sono qualificati.

## Autorità e isolamento

- Maestro non è una sandbox e usa i privilegi dell’utente locale.
- Il file selezionato e la domanda vengono inviati al provider configurato.
- Prompt injection nel sorgente viene trattata come contenuto non attendibile,
  ma non esiste una difesa perfetta contro output ingannevole.
- Tool, agent, retrieval e mutazioni non fanno parte del percorso Direct Chat;
  il codice sperimentale presente nel repository non amplia il support claim.
- Non esistono rollback, memoria persistente o recovery dopo restart.

## Prodotto ed ecosistema

- CLI e schema v2 sono sperimentali durante la serie 0.x.
- Non esistono installer di sistema, auto-update o service unit.
- Maestro non installa Ollama, modelli, PHP, Composer o dipendenze Laravel.
- Nessun packaging generalizzato di plugin/tool di terze parti.
- Nessun multi-agent, remote execution, shell, Git o Docker completi.

Vedere `compatibility.md` per la matrice autorevole e `troubleshooting.md` per
le azioni operative.
