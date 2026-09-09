# Architettura pubblica

Maestro separa la superficie utente dai componenti sperimentali. Il percorso
supportato attraversa confini espliciti e fallisce chiuso quando configurazione,
target, modello o autorizzazione non coincidono con il contratto.

```text
CLI
 ├─ Direct Chat ───────> Ollama locale
 └─ Controlled Mutation
      ├─ target esplicito
      ├─ proposta validata
      ├─ preview + allow-once
      └─ stale check + scrittura atomica
```

## Principi

- esecuzione locale e disclosure minima;
- capability esplicite, senza fallback che amplino l’autorità;
- separazione tra proposta del modello e commit dell’host;
- limiti e autorizzazioni verificati nel codice;
- eventi operativi redatti, senza prompt o contenuto dei file;
- componenti sperimentali non inclusi automaticamente nel prodotto.

Le API Go e la configurazione restano sperimentali durante la serie `0.x`.
Per il comportamento supportato vedere [Capacità correnti](current-capabilities.md)
e [Security Model](security-model.md).
