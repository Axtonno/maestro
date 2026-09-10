# Milestone 40 validation

Data: 2026-09-10  
Verdetto: `vscode_controlled_surface_prototype_validated`

La superficie controllata VS Code è stata verificata senza ampliare l'autorità
del runtime. Il verdetto riguarda soltanto il prototipo locale: non autorizza
pubblicazione Marketplace, inclusione negli archive Maestro o modifica del
support claim v0.5.0.

## Input congelati

- Maestro `v0.5.0`, release commit
  `86ed92495c5ce5bd2d0ec8d3c8a11b8c29a316f2`, binary SHA-256
  `e1765f9e8ed919eabe22b200f4145445da0d1b31f90c9b2bf6157a1390e37523`;
- archive Maestro SHA-256
  `0afcfe4d648edcde3caf4327c4f995606fb4c3974c05606e13f90dd8cff321d9`;
- Ollama `0.33.1`, archive SHA-256
  `88e0d36bd90121595e5516c84f6ab61b546368fbd2d825b4aae70999c949649d`;
- `qwen3.5:9b` digest
  `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`;
- `qwen2.5-coder:14b` digest
  `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`;
- VS Code Linux x64 `1.137.0`, commit
  `645f29cc3176500b4b5762ba887cf2a7f0ffdf2c`;
- Node.js Linux x64 `v24.20.0`.

Il server Ollama portabile è stato isolato su `127.0.0.1:11435` perché la
porta standard era già occupata. La copia temporanea del profilo pubblico ha
modificato soltanto `provider.base_url`; modelli, digest, prompt, schema,
workspace e controlli di generazione sono rimasti identici. L'esecuzione è
stata CPU-only.

## Matrice

| Caso | Esito | Evidenza essenziale |
| --- | --- | --- |
| V01 manifest e activation | PASS | workspace extension, trust e virtual workspace disabilitati |
| V02 registrazione comandi | PASS | quattro comandi registrati nell'Extension Development Host |
| V03 shell token escaping | PASS | unit test con apici e metacaratteri |
| V04 workspace containment | PASS | path esterno respinto |
| V05 coordinate selezione | PASS | conversione 0-based/exclusive in 1-based/inclusive |
| V06 scope mutativo | PASS | soltanto PHP sotto `app/` |
| V07 dirty e multi-selection | PASS | nessun terminale lanciato nell'host reale |
| V08 nessuna write o auto-approval | PASS | contratto statico e documento invariato dall'estensione |
| V09 Extension Host | PASS | VS Code reale, 4 test d'integrazione passati |
| V10 identità pubblica | PASS | release v0.5.0 e SHA-256 binario atteso |
| V11 doctor all | PASS | 14/14 controlli passati |
| V12 Direct Chat | PASS | file esatto, primo tentativo a setup completo, `finish_reason=stop` |
| V13 mutation deny | PASS | exit 3, `approval_rejected`, SHA-256 e Git diff invariati |
| V14 mutation allow | PASS | allow-once nel TTY, un file, una riga rimossa e una aggiunta |

La fixture era un repository Git pulito. Il caso mutativo ha usato
`app/Http/Controllers/OrderController.php:22`; preview e apply hanno sostituito
soltanto lo status `201` con `202`. Il file iniziale aveva SHA-256
`4826abe9c6c5d701133817a9dcb565f0b84f760da57e1b518d430b601520b1bd`;
dopo il deny era identico e dopo l'allow aveva SHA-256
`30097526418b813fef8ba74ece5825637881b83814ba0340b78f5f349c040215`.
Entrambe le preview hanno prodotto il medesimo diff SHA-256
`714e5a81f3cc7c11623c25ecef0694db513cecef7b4c9f89dfceb67da67d5f88`.

Un probe di setup eseguito prima che il secondo modello fosse disponibile è
terminato `provider_unavailable` senza effetti. Non è stato contato come caso
V12: la matrice live è iniziata soltanto dopo aver verificato la presenza e i
digest di entrambi i modelli richiesti.

## Comandi di regressione

```sh
npm test
MAESTRO_VSCODE_EXECUTABLE=/path/to/VSCode-linux-x64/code \
  npm run test:integration
```

Sono inoltre passati `go test ./...`, `go test -race ./...` e `go vet ./...`
sull'albero Maestro corrente.
