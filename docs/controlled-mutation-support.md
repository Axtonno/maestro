# Controlled Mutation: perimetro supportato

Controlled Mutation è il write-mode opt-in di Maestro v0.5.0. Non è
un'autorizzazione generale a modificare il workspace: trasforma una selezione
esplicita dell'utente in una sola sostituzione verificabile.

La [pagina delle capacità correnti](current-capabilities.md) separa il nucleo
supportato dalle superfici sperimentali e future.

## Contratto supportato

Una mutation è supportata soltanto quando sono vere tutte queste condizioni:

| Dimensione | Perimetro |
| --- | --- |
| Comando | `maestro workspace replace` |
| Target | Un file PHP regolare e non symlink sotto `app/` |
| Selezione | Un intervallo inclusivo `start:end` indicato dall'utente |
| Operazione | Sostituzione esatta del solo intervallo |
| Configurazione | Profilo strict v4 distribuito con v0.5.0 |
| Provider | Ollama locale su loopback |
| Modello | `qwen2.5-coder:14b` con digest qualificato |
| Interazione | TTY reale, preview completa e allow-once esplicito |
| Commit | Verifica stale e sostituzione atomica |

Il modello Direct Chat non viene usato come fallback per una mutation. Un
modello, digest, prompt o schema diverso fa fallire il preflight.

## Sequenza di controllo

```text
selezione esplicita
  -> lettura contained
  -> proposta strutturata
  -> validazione target e replacement
  -> preview completa + fingerprint
  -> allow-once / deny
  -> verifica che la sorgente non sia cambiata
  -> apply atomico oppure failure senza write
```

L'approvazione vale per quella preview e quella sorgente. Non è persistente e
non autorizza operazioni successive.

## Garanzie del percorso

- il path deve restare sotto `workspace.root/app`;
- coordinate e contenuto selezionato partecipano al fingerprint;
- il diff è mostrato prima dell'approvazione;
- input vuoto, EOF, valore sconosciuto e deny non scrivono;
- una sorgente cambiata dopo la preview produce `stale_source`;
- output non conforme, astensione o errore provider non scrivono;
- l'apply usa una sostituzione atomica e restituisce digest post-write;
- chat e mutation usano profili e modelli distinti, senza fallback incrociato.

Queste sono garanzie di contenimento e controllo dell'effetto, non una prova
che ogni proposta del modello sia semanticamente corretta. L'utente deve
leggere il diff.

## Fuori perimetro oggi

- scelta autonoma di file o righe;
- insert, delete o patch non riconducibili alla selezione fornita;
- più file, più intervalli o modifiche concatenate;
- file fuori `app/`, linguaggi diversi da PHP, directory o symlink;
- esecuzione senza TTY o approvazione automatica;
- shell, Git, Composer, test, Docker o comandi suggeriti dal modello;
- rollback automatico dopo un apply già completato;
- agent autonomi, retrieval o tool calling come fallback;
- provider remoti, altri modelli o digest non qualificati.

## Piattaforme osservate

Il percorso v0.5.0 è stato qualificato sull'artifact Linux `amd64`:

- Windows → WSL2 → filesystem Linux con RTX 5070 da 12 GB;
- Ubuntu 24.04.4 Linux nativo CPU-only su ThinkPad T490s, Intel i5-8365U.

La seconda prova è evidenza di adozione sul campo per quella macchina, non un
requisito minimo né una promessa universale su ogni PC Linux. Latenza e memoria
dipendono dall'hardware e dalla residenza dei modelli.

Non sono supportati Windows nativo, multi-file, agent autonomi, provider o
modelli alternativi. Ogni scrittura richiede approvazione allow-once.

## Uso consigliato

1. partire da un repository Git pulito;
2. eseguire `doctor --mode all`;
3. selezionare il minimo intervallo necessario;
4. leggere l'intera preview, non solo la riga aggiunta;
5. negare se target o diff non sono esatti;
6. verificare `git diff` e test del progetto dopo l'apply.

Le evidenze della qualifica sono nel
[report M38](reports/milestone-38-final.md). Il modello di sicurezza completo è
in [Security Model](security-model.md).
