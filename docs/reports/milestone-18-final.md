# Milestone 18 — Report finale

Data: 2026-08-29

Stato: **COMPLETATA**

Verdetto: **`v0.3.0_released_and_verified`**

Maestro v0.3.0 è pubblicato come baseline Direct Chat locale read-only per
zero o un singolo file esplicito. La milestone non ha introdotto cambi a
codice Direct Chat, prompt, schema, configurazione qualificata, modello,
digest, limiti, fixture o contratto CLI.

## Catena conclusiva

| Identità | Valore |
|---|---|
| Release | `https://github.com/Axtonno/maestro/releases/tag/v0.3.0` |
| tag | `v0.3.0`, annotato |
| commit release | `3f4c7d4b4fd2e380644cf250ce9e8fec2311af53` |
| archive | `maestro-v0.3.0-linux-amd64.tar.gz` |
| SHA-256 archive | `6c8f0e883ec8f8c05571fc2e7bc1f4ecac608c2bd7e338395ae0a4253fff1aaf` |
| dimensione | `3775317` byte |
| SHA-256 binario | `378a0533083b9a00be6c0212ca52001cebc5f77b476a20038bc8e08d1fc3d42d` |
| modello/digest | `qwen3.5:9b` / `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |

Le sette fasi sono PASS. L'RC è riproducibile, installabile e verde sulla
WSL2/Ubuntu 24.04/RTX 5070; l'artifact finale è byte-riproducibile; tag,
manifest, binario e versione condividono la medesima identità; i due asset
pubblici sono stati riscaricati e verificati da zero.

Restano non supportati agent, retrieval, multi-file, tool calling, plugin di
terze parti, Controlled Mutation, write/patch/approval, modelli differenti e
provider remoti. Nessuna di queste capability è implicata dal verdetto.

La Milestone 18 è chiusa. Ogni evoluzione funzionale successiva richiede una
nuova milestone e una nuova qualifica; gli asset v0.3.0 non vengono
sovrascritti.
