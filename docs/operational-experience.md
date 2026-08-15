# Maestro v0.1 Operational Experience

Stato: Contratto operativo pubblico sperimentale v0.1.0

Data: 2026-08-15

---

# Confine

La Fase 3 rende controllabile dal terminale il percorso applicativo già
consegnato. Non aggiunge tool, authority, sandbox, rollback o selezione
automatica. Permission e permit restano responsabilità del Tool Runtime;
l'Approver restituisce soltanto una scelta utente validata.

# Input e modalità

Con un'istruzione posizionale, stdin rimane disponibile per eventuali
approval. Se stdin è un TTY e manca l'argomento, Maestro legge una sola riga di
istruzione e usa lo stesso terminale per le approval successive.

Con stdin non interattivo, l'istruzione viene letta fino a EOF entro 1 MiB. In
questa modalità una policy `prompt` fallisce chiusa con permission denied,
anche se il pipe contiene testo che assomiglia a una risposta di approval.
L'automazione deve usare policy `allow` esplicite e bounded.

# Approval

L'approval resta un contratto sperimentale disponibile per integrazioni 0.x,
ma non fa parte del percorso ufficiale read-only della v0.1.0. La
configurazione distribuita non registra tool mutanti e imposta
`workspace_mutate: deny`; il quick start non presenta quindi prompt mutativi.

Prima di chiedere una decisione, Maestro mostra su stderr:

- subject `model` o `tool`;
- provider/modello oppure tool ID;
- effect e numero delle action preparate;
- resource logica e workspace quando applicabili;
- per una disclosure, soltanto workspace, sezioni, token e byte.

Non mostra instruction, prompt, contenuti del bundle, arguments, output tool,
fingerprint della disclosure, API key o root assoluta del workspace.

Le scelte sono:

```text
d / deny   nega; è anche il default su riga vuota
o / once   approva la permission corrente una sola volta
r / run    approva la stessa permission fingerprint per il run corrente
```

Il grant `run` non è una wildcard: una variazione di action, resource,
workspace, modello o disclosure produce una fingerprint diversa e richiede una
nuova decisione. Nessun grant sopravvive al run o al processo. EOF, input
troppo lungo o invalido e assenza di TTY negano. Cancellazione o deadline
interrompono l'attesa senza concedere authority.

# Progresso e risultato

stderr contiene una vista redatta e line-oriented:

```text
limits\t...
progress\trun=... state=...
plan\trun=... version=...
step\trun=... id=... state=...
permission\trun=... decision=... actions=...
progress\trun=... model_turns=... tool_calls=... tokens=.../...
terminal\trun=... reason=... model_turns=... tool_calls=... duration_ms=...
```

Il renderer si iscrive soltanto agli `EventPayload` allowlist di Agent e Tool
Runtime. Errori e panic del writer sono best-effort e non possono modificare lo
stato del run.

Su successo stdout contiene run ID, terminale, contatori e una sezione
`result`; il contenuto finale del modello è l'unico contenuto applicativo
stampato intenzionalmente. Un run fallito non scrive risultato su stdout e
stampa su stderr un reason code sintetico, non la catena d'errore provider.

# Cancellazione e limiti

SIGINT/SIGTERM cancellano il context del comando e producono exit code 130.
L'attesa di approval e le chiamate runtime osservano lo stesso context. Lo
shutdown applicativo ha un bound di 30 secondi. La cancellazione non promette
rollback degli effetti già iniziati o completati.

Prima dell'esecuzione vengono mostrati tutti gli hard limit configurati:
durata, turni modello, tool call totali/per turno, step/revisioni, token e byte.
Il terminale `limit_exceeded` resta autorevole quando uno di questi bound viene
raggiunto.

# Non garanzie

L'approval non è una sandbox e non limita i privilegi del processo. La v0.1.0
esegue codice trusted in-process e i workspace tool operano con i privilegi
dell'utente locale. Non esiste auto-approval, `--yes`, memoria di grant tra run
o rollback generale.

La v0.1.0 non dichiara supportati `workspace.write`, `workspace.patch`, un
reference agent mutante o llama.cpp. La presenza del relativo codice e dei test
deterministici non costituisce una promessa operativa live.
