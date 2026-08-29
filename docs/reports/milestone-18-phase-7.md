# Milestone 18 — Fase 7: GitHub Release e verifica post-download

Data: 2026-08-29

Stato: **COMPLETATA — PASS PUBBLICO**

## Pubblicazione

La GitHub Release pubblica è:

`https://github.com/Axtonno/maestro/releases/tag/v0.3.0`

Il tag annotato remoto `v0.3.0` punta al commit release
`3f4c7d4b4fd2e380644cf250ce9e8fec2311af53`. La Release non è draft né
prerelease e contiene esclusivamente:

- `maestro-v0.3.0-linux-amd64.tar.gz`;
- `maestro-v0.3.0-linux-amd64.tar.gz.sha256`.

Nessun asset preesistente è stato sovrascritto. I primi tentativi di creazione
sono stati rifiutati dall'API prima della write per problemi di serializzazione
JSON. La Release è stata quindi creata una sola volta. Il primo URL di upload
è stato respinto localmente prima della richiesta; la Release risultava
pubblica e con zero asset. Dopo la verifica esplicita di identità e assenza di
asset, il workflow ha caricato la coppia qualificata senza sostituzioni.

## Verifica dal canale pubblico

Entrambi gli asset sono stati scaricati dagli URL pubblici in una directory
nuova sotto `/tmp`; le copie locali sotto `dist/` non sono state usate per i
gate post-download.

| Gate | Esito |
|---|---|
| visibilità URL e Release | PASS |
| tag remoto annotato e target | PASS |
| nomi e numero asset | PASS — esattamente 2 |
| dimensione archive | PASS — `3775317` byte |
| SHA-256 archive | PASS — `6c8f0e883ec8f8c05571fc2e7bc1f4ecac608c2bd7e338395ae0a4253fff1aaf` |
| checksum asset | PASS |
| estrazione pulita | PASS |
| manifest versione/stato/commit | PASS — `v0.3.0` / `release` / `3f4c7d4…` |
| SHA-256 binario | PASS — `378a0533083b9a00be6c0212ca52001cebc5f77b476a20038bc8e08d1fc3d42d` |
| profilo e modello/digest | PASS |
| `maestro version`, root help, chat help | PASS |
| installazione in prefix separato | PASS |

## Gate finale

Verdetto Fase 7: **PASS PUBBLICO**. Tag, Release, archive e checksum pubblici
corrispondono agli artifact qualificati. La Fase 7 autorizza la chiusura della
Milestone 18 con verdetto `v0.3.0_released_and_verified`.
