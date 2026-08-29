# Milestone 18 — Fase 3: release candidate riproducibile

Data: 2026-08-29

Stato: **COMPLETATA — PASS**

## Release candidate congelato

| Campo | Valore |
|---|---|
| versione | `v0.3.0-rc.1` |
| stato manifest | `release-candidate` |
| commit sorgente | `f33ce456cd65c24abcd5561d7140438ff08e64f1` |
| archive | `maestro-v0.3.0-rc.1-linux-amd64.tar.gz` |
| dimensione | `3775354` byte |
| SHA-256 archive | `b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a` |
| SHA-256 binario | `59e50848fdb6e4a6d85bccea2fe4aadb98aaa128fd692dcdbf467738d4c1a607` |
| SHA-256 profilo | `1c5bbe79edf125485d14518e58ff18c48156eaa0fb91faf82fcf3cd97375d0ee` |
| SHA-256 fixture route | `7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39` |
| modello/digest | `qwen3.5:9b` / `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| toolchain/target | Go 1.24.5 / Linux `amd64`, CGO disabilitato |

Archive e checksum sono conservati senza overwrite sotto `dist/`. L'RC è un
artifact distinto dal packaging candidate M17 e non è stato ottenuto tramite
rename o modifica dell'archive precedente.

## Riproducibilità e baseline

La verifica della Fase 2 ha costruito due archive byte-identici dal medesimo
commit sorgente; la build persistente della Fase 3 coincide con lo stesso
SHA-256. Per conservare il commit esatto mentre il branch aggiungeva soltanto
report di governance, il packaging persistente è stato eseguito da una copia
Git locale detached su `f33ce45`.

Il diff del commit sorgente rispetto al candidate M17 è esclusivamente
documentale. Codice Direct Chat, prompt, profilo, fixture, script, dipendenze,
modello, digest, context, thinking, temperatura, streaming e limiti sono
invariati. Il binario cambia identità rispetto a `pc.1` perché incorpora
versione `v0.3.0-rc.1` e commit RC, come richiesto dal contratto di release.

## Archive audit

| Gate | Esito |
|---|---|
| checksum sidecar | PASS |
| doppia build byte-identica | PASS |
| manifest/versione/commit/stato | PASS |
| allowlist e 44 entry attese | PASS |
| documenti obbligatori e release notes | PASS |
| token `@MAESTRO_*` risolti | PASS |
| profilo chat-only byte-identico | PASS |
| fixture route byte-identica | PASS |
| assenza profili agentici/mutativi e report | PASS |
| assenza symlink, VCS, `.env`, vendor e node_modules | PASS |
| scansione path checkout e credential-shaped | PASS |
| `maestro version` dall'estrazione | PASS |

Una prima espressione regolare usata sulla lista archive aveva una parentesi
mancante e non costituisce evidenza. L'audit è stato ripetuto materializzando
la lista e scandendo anche l'estrazione completa; la ripetizione valida è
interamente verde.

## Gate finale

Verdetto Fase 3: **PASS**. `v0.3.0-rc.1` è l'unico RC autorizzato per le Fasi
4 e 5. Non deve essere ricostruito, sostituito o sovrascritto. Nessun tag,
push o asset remoto è stato creato.
