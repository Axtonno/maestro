# Milestone 18 — Fase 4: installazione pulita e audit RC

Data: 2026-08-29

Stato: **COMPLETATA — PASS OFFLINE**

## Artifact verificato

La fase ha usato esclusivamente la coppia persistente:

- `maestro-v0.3.0-rc.1-linux-amd64.tar.gz`;
- `maestro-v0.3.0-rc.1-linux-amd64.tar.gz.sha256`;
- SHA-256 `b034828a07f33a2643556123c00917ff563d83f1976dab968542712f0df7be3a`;
- commit manifest/binario `f33ce456cd65c24abcd5561d7140438ff08e64f1`;
- stato `release-candidate`.

Archive e checksum sono stati copiati in una directory nuova sotto `/tmp` e
verificati prima dell'estrazione. Non è stato usato alcun file del checkout,
né è stato ricostruito il binario.

## Gate di installazione

| Gate | Esito |
|---|---|
| checksum dalla directory pulita | PASS |
| estrazione nuova fuori checkout | PASS |
| manifest/versione/commit/stato | PASS |
| root help | PASS |
| `chat --help` | PASS |
| installazione binario in prefix separato | PASS |
| version/help da working directory vuota | PASS |
| doctor chat offline | PASS fail-closed |
| containment traversal pre-provider | PASS |
| fixture immutabile | PASS |
| scansione path checkout e credential-shaped | PASS |

Il doctor offline termina exit 1 con config, workspace e composition `pass`,
model `fail required_capability_unavailable` e generation
`skip model_unavailable`. È il comportamento atteso quando il provider non è
disponibile e non viene reinterpretato come gate live.

Il traversal `../outside.php` termina exit 2 con il solo messaggio redatto
`chat failed: file_not_allowed`, prima di qualsiasi chiamata provider. Il
digest ricorsivo della fixture è rimasto identico prima e dopo:
`ae8483e599d7495b10333d00980951680800632ea7b437425d022cd841841fe7`.

## Confine live

La macchina corrente è Ubuntu 24.04 sul ThinkPad, non espone `nvidia-smi` e
non ha Ollama raggiungibile su loopback. La ripetizione live del quick start
non appartiene al PASS offline di questa fase e resta integralmente assegnata
alla Fase 5 sulla WSL2/Ubuntu 24.04/RTX 5070 qualificata.

## Gate finale

Verdetto Fase 4: **PASS OFFLINE**. L'RC è installabile e supera audit,
containment e immutabilità senza provider. La Fase 5 è autorizzata a usare
esattamente questo archive, senza rebuild o tuning. Nessun tag, push o asset
remoto è stato creato.
