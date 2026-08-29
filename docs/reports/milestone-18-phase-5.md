# Milestone 18 — Fase 5: gate live RC sulla RTX 5070

Data: 2026-08-29

Stato: **BLOCKED — `release_environment_blocked`**

## Decisione

La matrice live non è stata eseguita e non viene reinterpretata come PASS. La
macchina corrente è `antonio-cafeo-ThinkPad-T490s`, Linux `x86_64` Ubuntu
24.04, senza `nvidia-smi`, senza bridge Windows/WSL disponibile e senza Ollama
raggiungibile su `127.0.0.1:11434`.

Il piano richiede la stessa WSL2/Ubuntu 24.04/RTX 5070 della qualifica M17 con
Ollama 0.33.1 e il modello/digest congelato già disponibili. Non autorizza ad
avviare provider, scaricare modelli, modificare cataloghi o sostituire la
piattaforma. Fasi 6 e 7 restano quindi vietate.

## RC da trasferire senza rebuild

| Campo | Valore |
|---|---|
| archive | `maestro-v0.3.0-rc.1-linux-amd64.tar.gz` |
| checksum | `maestro-v0.3.0-rc.1-linux-amd64.tar.gz.sha256` |
| SHA-256 | `b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a` |
| dimensione | `3775354` byte |
| versione/stato | `v0.3.0-rc.1` / `release-candidate` |
| commit | `f33ce456cd65c24abcd5561d7140438ff08e64f1` |
| SHA-256 binario | `59e50848fdb6e4a6d85bccea2fe4aadb98aaa128fd692dcdbf467738d4c1a607` |
| profilo | SHA-256 `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee` |
| fixture route | SHA-256 `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |

Il checksum locale è stato verificato nuovamente dopo il preflight ambientale.
Questi sono gli unici asset ammessi; la macchina live non deve ricostruirli.

## Matrice live

| Gate | Esito |
|---|---|
| identità archive/binario/config | `NOT_RUN` |
| Ollama 0.33.1, modello e digest | `NOT_RUN` |
| installazione nuova fuori checkout | `NOT_RUN` |
| version e help | `NOT_RUN` |
| doctor chat 5/5 | `NOT_RUN` |
| no-file | `NOT_RUN` |
| single-file complete | `NOT_RUN` |
| single-file stream/equivalenza | `NOT_RUN` |
| traversal e symlink evasivo | `NOT_RUN` |
| SIGINT | `NOT_RUN` |
| deadline | `NOT_RUN` |
| fixture invariata e anti-leak | `NOT_RUN` |

## Protocollo di ripresa

Sulla piattaforma qualificata:

1. trasferire archive e checksum byte-identici e verificare SHA-256 prima
   dell'estrazione;
2. estrarre in una directory nuova fuori checkout e usare soltanto quei file;
3. confrontare manifest, `maestro version`, binario, profilo, modello, digest,
   context 4096, thinking false e temperatura zero con la tabella sopra;
4. verificare Ollama 0.33.1 e il digest modello senza pull, update o load
   amministrativo implicito;
5. richiedere doctor chat 5/5;
6. eseguire no-file, single-file complete e `--stream` sulla fixture inclusa;
   complete e stream devono riportare `POST /orders` e
   `OrderController::store`, terminale `completed`, finish `stop`, stderr
   vuoto ed equivalenza semantica;
7. eseguire traversal e symlink evasivo: exit 2, `file_not_allowed`, stdout
   vuoto e nessuna chiamata provider;
8. eseguire SIGINT: exit 130, `canceled`, stdout vuoto;
9. eseguire il negativo deadline con copia del profilo collocata sotto
   `configs/` e soli timeout ridotti: exit 4, `deadline_exceeded`, stdout
   vuoto; la copia non viene usata per run positive e non modifica l'archive;
10. confrontare il digest ricorsivo fixture pre/post e scandire l'evidenza
    redatta per escludere domanda, contenuto, response, root fisica, canary e
    secret; eliminare gli output raw dopo la valutazione.

Qualsiasi errore harness richiede la ripetizione dell'intera matrice sullo
stesso RC. Una failure del prodotto respinge `rc.1`; tuning, retry selettivo,
rebuild o sostituzione modello non sono ammessi.

## Gate

Verdetto provvisorio: **`release_environment_blocked`**. Per riprendere è
necessario rendere disponibile la piattaforma qualificata e restituire la
matrice completa redatta sullo stesso SHA-256. Fino ad allora non vengono
costruiti artifact finali, creati tag, pushati commit o pubblicati asset.
