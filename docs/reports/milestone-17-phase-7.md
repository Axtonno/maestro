# Milestone 17 — Fase 7: packaging candidate e qualifica finale

Data: 2026-08-29

Stato: **COMPLETATA — PASS, MATRICE FINALE LIVE VERDE**

## Candidate congelato

| Campo | Valore |
|---|---|
| versione | `v0.3.0-pc.1` |
| stato manifest | `packaging-candidate` |
| commit sorgente | `70a9630203ccf82a4d8858a9e47b48f5333b9cbd` |
| archive | `maestro-v0.3.0-pc.1-linux-amd64.tar.gz` |
| dimensione | `3776699` byte |
| SHA-256 archive | `82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61` |
| SHA-256 binario nell’archive | `dee9d5113ccf2db0573b03e8a3851f600d7bc789964793ebae14376f9c849a66` |
| profilo | `configs/maestro.chat.example.yaml` |
| SHA-256 profilo | `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee` |
| modello | `qwen3.5:9b` |
| digest modello richiesto | `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| context / thinking / temperatura | `4096` / `false` / `0` |
| streaming | abilitato, opt-in |
| file / output massimi | 1 MiB / 1 MiB |
| SHA-256 servizio/prompt | `7fd79e1fafb70d0b7726ecca0909f92592f8706df890a9b6fb263c9d5b8575c1` |
| SHA-256 fixture `routes/api.php` | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |
| toolchain / target | Go 1.24.5 / Linux `amd64`, CGO disabilitato |

Archive e checksum sono conservati localmente sotto `dist/`, directory esclusa
da Git. La piattaforma finale deve ricevere entrambi byte-identici; non deve
ricostruire il binario.

## Coerenza con F6.4

Codice Direct Chat, prompt e fixture non differiscono dal freeze F6.4. Il
profilo installabile conserva esattamente modello, timeout, streaming,
`num_ctx`, thinking, temperatura e limiti qualificati. Rimuove i blocchi
agentici development-only e usa come root relativa la copia della medesima
fixture inclusa nell’archive; non introduce tuning o una nuova capability.

## Gate locali

| Gate | Esito |
|---|---|
| `go test -count=3 ./...` | PASS |
| `go test -tags maestro_development -count=3 ./...` | PASS |
| `go test -race ./...` | PASS |
| `go vet ./...` | PASS |
| `bash -n scripts/*.sh` | PASS |
| doppio packaging dal medesimo commit | PASS — byte-identico |
| checksum e path archive | PASS |
| allowlist e file obbligatori | PASS |
| assenza config agentiche/mutative e materiale development-only | PASS |
| token documentali renderizzati | PASS |
| installazione fuori checkout, version e help | PASS |
| doctor chat offline fail-closed sul solo provider | PASS |
| containment offline prima del provider | PASS |
| scansione path checkout, credential-shaped, symlink e fixture | PASS |

Il gate locale non avvia Ollama, non scarica modelli e non viene interpretato
come prova live.

## Protocollo finale immutabile

Eseguire sulla WSL2/Ubuntu 24.04/RTX 5070, Ollama 0.33.1, senza rebuild,
tuning, sostituzione di modello o retry selettivi:

1. verificare SHA-256 dell’archive e confrontare versione, commit, stato,
   modello, digest e parametri del manifest;
2. estrarre in una directory nuova fuori da ogni checkout e usare soltanto i
   file dell’archive;
3. verificare `maestro version`, root help e `chat --help`;
4. verificare il digest modello e ottenere doctor chat 5/5;
5. eseguire una domanda no-file e richiedere l’esplicita assenza di evidenza di
   progetto;
6. eseguire su `routes/api.php` la domanda F6.4 su endpoint/controller/action
   una volta complete e una volta `--stream`; entrambe devono riportare
   `POST /orders` e `OrderController::store`, senza endpoint aggiunti, con
   terminale `completed`, finish `stop`, stderr vuoto ed equivalenza semantica;
7. ripetere i negativi traversal e symlink evasivo: exit 2,
   `file_not_allowed`, stdout vuoto e nessuna chiamata provider;
8. interrompere una generation con SIGINT: exit 130, `canceled`, stdout vuoto;
9. usare una copia temporanea del profilo con soli timeout provider/chat
   ridotti per il test negativo: exit 4, `deadline_exceeded`, stdout vuoto;
10. confrontare il digest ricorsivo della fixture prima/dopo e scandire stderr
    ed evidenza redatta per escludere domanda, contenuto, response, root fisica
    e canary/secret; eliminare gli output raw locali dopo la valutazione.

Il profilo temporaneo del punto 9 non viene usato per run di successo e non
modifica l’archive. Ogni failure materiale arresta la serie. `not_run`, `skip`,
`unknown` e risultati di un archive diverso non valgono come PASS.

## Matrice finale live

| Gate | Esito | Evidenza redatta |
|---|---|---|
| identità archive/manifest/binario | PASS | archive `82bfb33f…`, binario `dee9d511…`, versione e commit esatti |
| modello e digest | PASS | `qwen3.5:9b`, digest `6488c96f…`, Ollama 0.33.1 |
| installazione pulita | PASS | estrazione nuova sotto `/tmp`, fuori dal checkout, nessun rebuild |
| version e help | PASS | versione/commit, root help e chat help, exit 0 |
| doctor chat | PASS 5/5 | config, workspace, composition, model e generation |
| no-file | PASS | assenza di evidenza dichiarata, zero fatti di progetto inventati |
| single-file complete | PASS | `POST /orders`, `OrderController::store`, terminale `completed`, finish `stop` |
| single-file stream ed equivalenza | PASS | risposta semanticamente identica a complete, terminale e usage coerenti |
| traversal e symlink | PASS 2/2 | exit 2, `file_not_allowed`, stdout vuoto |
| cancellazione | PASS | SIGINT, exit 130, `canceled`, stdout vuoto |
| deadline | PASS | timeout temporaneo, exit 4, `deadline_exceeded`, stdout vuoto |
| immutabilità | PASS | digest ricorsivo pre/post identico |
| anti-leak | PASS | canary, secret e root fisica assenti dai failure sink |

## Evidenza operativa finale

La serie autorevole è stata eseguita sulla WSL2/Ubuntu 24.04/RTX 5070 usando
soltanto i file estratti dall'archive. SHA-256 verificati prima della prima
generation:

- archive: `82bfb33f3fd9af911e3b2b1e89f9920177b281046da21b186512e577e114fb61`;
- binario: `dee9d5113ccf2db0573b03e8a3851f600d7bc789964793ebae14376f9c849a66`;
- profilo: `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee`.

Le tre generation positive hanno exit 0, stderr vuoto, terminale `completed`,
finish `stop` e `truncated=false`. Latenze end-to-end osservate: no-file
6.588 ms, complete 627 ms, stream 552 ms. Il digest ricorsivo locale della
fixture è `42642ed5ec5f814fcd82e4c60adb17a03ccbb8346500b0357265c4ecd60112de`
prima e dopo la matrice.

Il primo tentativo del solo test deadline non ha raggiunto il provider perché
la copia temporanea del profilo era stata collocata fuori da `configs/`,
alterando la risoluzione relativa della workspace root. Ha terminato
`file_not_allowed` prima della generation e non costituisce un failure del
candidate. L'harness è stato corretto esclusivamente nella posizione della
copia e l'intera matrice, non il solo gate, è stata ripetuta sul medesimo
archive senza tuning; la seconda serie completa è quella autorevole sopra.

Gli output raw locali sono stati eliminati dopo valutazione e scansione.

## Gate

Il packaging locale e la matrice finale sullo stesso archive sono **PASS**.
La Fase 7 è completata e autorizza il verdetto
`direct_chat_product_baseline`, limitato al perimetro qualificato. La creazione
di tag, release candidate/release artifact e pubblicazione sono azioni
separate: non sono state eseguite durante questa qualifica.
