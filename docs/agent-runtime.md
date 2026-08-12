# Agent Runtime

Versione: 0.1.0

Stato: Implementato

Data: 2026-08-12

---

# Loop operativo

L'Agent Runtime esegue ogni step ready del piano in sequenza. Per lo step
corrente costruisce una conversazione con ruoli system, user, assistant e tool,
autorizza la chiamata modello, invoca il Provider Runtime esatto e interpreta
una risposta finale oppure le tool call.

Le tool call multiple restano nell'ordine dichiarato dal provider e non
vengono parallelizzate. Al termine testuale lo step diventa completed; quando
tutti gli step sono completed o skipped, la sessione termina con `completed` e
l'ultimo contenuto testuale diventa il `RunResult`.

# Tool adapter

Solo i descriptor registrati nel Tool Runtime e inclusi esplicitamente nella
`RunRequest` vengono convertiti in `provider.Tool`. Il mapping usa il nome
provider-compatible del descriptor e conserva internamente l'ID autorevole.

Per ogni call il runtime:

1. valida cardinalità, nome, arguments JSON e call ID;
2. genera un ID correlabile se il provider non ne fornisce uno;
3. rifiuta ID duplicati nell'intero run;
4. costruisce `tool.Invocation` ed `ExecutionRequest`;
5. attraversa il solo percorso pubblico `Tool Runtime.Invoke`;
6. serializza il risultato in un messaggio tool JSON tipizzato.

Il messaggio include outcome, reason, content, media type, item count,
truncation e deny disposition. Un deny recoverable torna al modello; un deny
terminale chiude il run. Non esiste retry implicito di tool o effetti.

# Provider e streaming

`RunRequest.Streaming()` seleziona `Provider Runtime.Stream`; la modalità
predefinita usa `Complete`. L'assembler raggruppa i delta tool per indice,
preserva l'ordine, richiede ID/nome coerenti e concatena gli arguments soltanto
entro i ceiling di byte e call. EOF chiude lo stream; failure mid-stream scarta
la risposta parziale e diventa provider failure.

La risposta assemblata attraversa la stessa validazione usata dalla modalità
non-stream. Finish reason sconosciuti, tool call malformate o cardinalità
eccedente non raggiungono il Tool Runtime.

# Budget e terminali

Un turno viene consumato prima della chiamata provider; usage e byte vengono
registrati prima dell'iterazione successiva. Ogni tool call consuma il budget
prima di `Invoke`, mentre ogni risultato consuma byte prima di tornare al
modello. Sono applicati sia `MaxToolCalls` sia `MaxToolCallsPerTurn`.

Cancellation e deadline attraversano provider, approval e tool execution.
Errori vengono classificati nei terminali permission, limit, provider, tool,
canceled o deadline senza trasformare una failure in una nuova iterazione.

# Confini

Il loop usa esclusivamente contratti provider-neutral e tool-neutral. Non
seleziona automaticamente provider/modello, non invoca adapter concreti e non
introduce esecuzione parallela, retry semantici o memoria persistente.
