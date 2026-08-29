# Milestone 18 — Fase 5: gate live RC sulla RTX 5070

Data: 2026-08-29

Stato: **COMPLETATA — PASS LIVE**

## Identità immutabile

La matrice autorevole è stata eseguita sulla WSL2/Ubuntu 24.04/RTX 5070
usando esclusivamente l'archive RC già congelato, copiato in una directory
nuova sotto `/tmp` e verificato prima dell'estrazione. Non è stato eseguito
alcun rebuild, pull, tuning o cambio modello.

| Campo | Valore |
|---|---|
| archive | `maestro-v0.3.0-rc.1-linux-amd64.tar.gz` |
| SHA-256 archive | `b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a` |
| dimensione | `3775354` byte |
| commit | `f33ce456cd65c24abcd5561d7140438ff08e64f1` |
| SHA-256 binario | `59e50848fdb6e4a6d85bccea2fe4aadb98aaa128fd692dcdbf467738d4c1a607` |
| SHA-256 profilo | `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee` |
| SHA-256 fixture route | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |
| provider | Ollama 0.33.1 su loopback |
| modello/digest | `qwen3.5:9b` / `6488c96fa5fa…` |

## Matrice live

| Gate | Esito |
|---|---|
| identità archive/manifest/binario/config | PASS |
| RTX 5070, Ollama 0.33.1, modello e digest | PASS |
| installazione nuova fuori checkout | PASS |
| version e help | PASS |
| doctor chat | PASS 5/5 |
| no-file | PASS — assenza di evidenza dichiarata |
| single-file complete | PASS — `POST /orders`, `OrderController::store` |
| single-file stream/equivalenza | PASS — medesimi fatti e terminali |
| traversal | PASS — exit 2, `file_not_allowed`, stdout vuoto |
| symlink evasivo | PASS — exit 2, `file_not_allowed`, stdout vuoto |
| SIGINT | PASS — exit 130, `canceled`, stdout vuoto |
| deadline | PASS — exit 4, `deadline_exceeded`, stdout vuoto |
| fixture invariata | PASS |
| anti-leak | PASS |

Le tre generation positive hanno terminato con exit 0, stderr vuoto, stato
`completed` e finish `stop`. Il digest ricorsivo redatto della fixture è
rimasto identico prima e dopo:
`e3a1370e032ed2a8742e3a2772bc2e5035540d975e42c95a11fb53bd07de18da`.
Gli output raw temporanei sono stati eliminati automaticamente al termine.

## Gate

Verdetto Fase 5: **PASS LIVE**. Il precedente
`release_environment_blocked` è risolto. La Fase 6 è autorizzata a produrre
l'artifact finale dal commit release pulito; la RTX 5070 non è richiesta di
nuovo se il delta resta limitato ai metadata e documenti approvati.
