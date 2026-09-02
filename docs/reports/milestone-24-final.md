# Milestone 24 — Report finale

Data: 2026-09-02

Stato: **COMPLETATA**

Verdetto: **`v0.3.1_released_and_verified`**

Maestro v0.3.1 è pubblicato come patch release del percorso Direct Chat
single-file read-only. La milestone ha sostituito soltanto il limite
generativo 512 respinto da M23 con il profilo congelato 1024; compatibilità
v2, schema v3, diagnostica, heartbeat, residency, identità e confine di
autorità restano invariati.

## Catena conclusiva

| Identità | Valore |
|---|---|
| Release | `https://github.com/Axtonno/maestro/releases/tag/v0.3.1` |
| tag | `v0.3.1`, annotato |
| commit release | `bd0e902c8d7ef01c01117537fceed76845a33732` |
| archive | `maestro-v0.3.1-linux-amd64.tar.gz` |
| SHA-256 archive | `2420ba89ada7b0b9cf3de8bd62d7f97dc32868aa342e44e5c3dacbaa94b3a6b6` |
| dimensione archive | `3791820` byte |
| SHA-256 binario | `0d5e068019e5187c517f9ff0bc7966b5f3123be933b6d858f2f2fa16978c36ed` |
| modello/digest | `qwen3.5:9b` / `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| profilo | schema v3, context 4096, `num_predict: 1024`, residency `5m` |

La release è pubblica, non draft e non prerelease. GitHub registra la
pubblicazione il 2026-09-02 alle 17:53:55 UTC.

## Recupero del limite generativo

Le cinque osservazioni progettuali dell'asset v0.3.0 hanno prodotto da 71 a
944 token con terminale `stop`. Il candidate 1024 è stato congelato prima
della qualifica e non modificato durante le run.

| Task | Token v0.3.0 | Token candidate | Terminale candidate | Qualità candidate |
|---|---:|---:|---|---|
| Q17-1 | 721 | 535 | `stop` | PASS |
| Q17-2 | 71 | 110 | `stop` | PASS |
| Q17-3 | 441 | 422 | `stop` | PASS |
| Q17-4 | 796 | 564 | `stop` | PASS |
| Q17-5 | 944 | 537 | `stop` | FAIL |

Q17-5 contiene proposte oltre i fatti visibili sia nella baseline appaiata
sia nel candidate; non è una regressione introdotta dal limite. Il gate è
completion 5/5, qualità 4/5, terminali `length` zero, regressioni appaiate
zero, claim materialmente falsi nei PASS zero e workspace invariato. Q17-2
complete/stream è coerente e termina `stop` in entrambi i trasporti.

## Qualifica deterministica, LF e RC

Su una materializzazione Linux LF sono PASS:

- `go test ./...`;
- `go test -race ./...`;
- `go vet ./...`;
- sintassi degli script e `git diff --check`.

Il vero RC `v0.3.1-rc.2` incorpora il commit `6b879d8` e due build indipendenti
sono byte-identiche con SHA-256
`fb141965d972e321df2e7d4349c22abcb4ccb4d6054a458220668f8deab56867`.
Estratto fuori dal checkout, supera identità, doctor, no-file, complete,
stream, containment, redazione, parametri richiesti e immutabilità.

## Artifact finale e canale pubblico

Due build finali dal commit release sono byte-identiche. L'esatto archive
finale, distinto dall'RC, supera lo stesso gate live prima del tag. Dopo la
pubblicazione, archive e checksum sono stati riscaricati in una directory
nuova e coincidono byte per byte con gli asset locali.

L'installazione dal download pubblico conferma:

```text
version  v0.3.1
status   release
commit   bd0e902c8d7ef01c01117537fceed76845a33732
dirty    false
sha256   0d5e068019e5187c517f9ff0bc7966b5f3123be933b6d858f2f2fa16978c36ed
```

Doctor, no-file, complete, stream, containment, anti-leak,
`num_predict: 1024`, residency `5m` e immutabilità sono nuovamente PASS sul download
pubblico. Non è avvenuto alcun rebuild, tuning o cambio funzionale tra
qualifica, tag e pubblicazione.

## Confine conservato

Restano non supportati CPU, agent, retrieval, multi-file, tool calling,
provider remoti e Controlled Mutation. v0.3.1 non amplia l'autorità del
prodotto e non reinterpreta come supportate le capability presenti nel
repository.
