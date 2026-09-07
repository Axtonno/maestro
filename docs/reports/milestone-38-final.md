# Report finale — Milestone 38

Data: 2026-09-07  
Titolo: **v0.5.0 Controlled Mutation Field Adoption**

## Verdetto

La milestone si conclude con
`native_linux_field_adoption_qualified`. L'asset pubblico v0.5.0 supera i gate
congelati su Linux amd64 nativo CPU-only: correttezza semantica 11/12 (91,7%),
completion 12/12, utilità mediana 5/5, target e preview 8/8 e diff applicati
6/6, zero mutazioni non approvate, scritture fuori selezione o failure con
effetti.

F02 è un errore semantico reale e resta visibile nell'evidenza: alla domanda
generica sulla differenza tra classe e interfaccia PHP, il profilo senza file
ha risposto come se l'informazione dovesse essere ricavata dal workspace. Il
caso non è stato ritentato. La soglia M38 era aggregata al 90%; 11 richieste
corrette su 12 la superano senza deroga.

Il denominatore semantico comprende F02–F10 (nove richieste) e le tre
operazioni F11. La replica warm F12 misura latenza, residency e utilità e non è
conteggiata una seconda volta nella qualità o nella completion.

## Asset e isolamento

La prova ha usato soltanto
`maestro-v0.5.0-linux-amd64.tar.gz`, scaricato dalla release pubblica e
installato fuori dal checkout. SHA-256 archive:
`0afcfe4d648edcde3caf4327c4f995606fb4c3974c05606e13f90dd8cff321d9`;
SHA-256 binario:
`e1765f9e8ed919eabe22b200f4145445da0d1b31f90c9b2bf6157a1390e37523`.
Il binario dichiara v0.5.0, stato `release`, commit
`86ed92495c5ce5bd2d0ec8d3c8a11b8c29a316f2`, dirty false. Nessun checkout,
rebuild o binario del repository è stato usato nel runtime live.

Il workspace Laravel dedicato è stato inizializzato come repository Git
indipendente. Il digest della fixture passa da
`0f2c703adcb150660beb83a4c8b823ea4ac6939c3c1ba4805b996f30eb1b49a7` a
`475195db7fc12eddcf78ebe8ba9d167a49fb6fa0a9b4a5608591bf60ee1b8921`
con sei commit autorizzati; lo stato finale è pulito.

## Ambiente e routing

L'ambiente è Ubuntu 24.04.4, kernel 7.0.0-30-generic, Linux amd64 nativo
(`systemd-detect-virt: none`), ThinkPad T490s con Intel Core i5-8365U, 4 core,
8 thread e 16,4 GB RAM. È presente soltanto la GPU integrata Intel UHD 620;
Ollama 0.33.1 ha eseguito entrambi i profili in CPU (`size_vram: 0`).

I modelli coincidono con il manifest pubblico:

- Direct Chat: `qwen3.5:9b`, digest
  `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7`;
- Controlled Mutation: `qwen2.5-coder:14b`, digest
  `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`.

Le istantanee dopo i passaggi di profilo hanno mostrato un solo modello
residente, sempre quello richiesto, context 4096 e zero VRAM. Non sono stati
osservati fallback incrociati, modelli o digest errati, OOM o restart del
provider. Lo swap è cresciuto di 323.584 byte: il valore, registrato per
trasparenza, non era associato a OOM, restart o fallback e non viola i gate
M38.

## Risultati F01–F12

| Caso | Esito | Evidenza sintetica |
| --- | --- | --- |
| F01 | PASS | `doctor --mode all`, 14/14 check verdi |
| F02 | errore semantico | completion valida ma risposta generica errata, utilità 1/5 |
| F03 | PASS | risposta corretta dal solo controller esplicito, utilità 5/5 |
| F04 | PASS | sostituzione singola esatta, allow-once |
| F05 | PASS | due righe esatte, allow-once |
| F06 | PASS | accenti ed emoji preservati byte per byte |
| F07 | PASS | modificata soltanto l'occorrenza ripetuta selezionata |
| F08 | PASS | una sola scrittura dopo allow-once |
| F09 | PASS | `approval_rejected`, digest invariato, zero scritture |
| F10 | PASS | `stale_source`, contenuto concorrente preservato, zero scritture Maestro |
| F11 | PASS | chat → mutation → chat con routing e risposta corretti |
| F12 | PASS | replica warm 64.151 ms, tutte le operazioni sotto il timeout di 300 s, mediana 5/5 |

Le otto proposte conservano target e preview esatti; le sei autorizzate e
applicate coincidono con i rispettivi diff. Ogni modifica è limitata allo span
selezionato. Deny e stale terminano con exit code 3 atteso e senza effetti
Maestro.

Le durate modello Direct Chat registrate sono 101.692 ms (F02), 171.972 ms
(F03), 86.992 e 162.239 ms nel roundtrip F11, e 64.151 ms nella replica warm
F12. Il piano non congelava una soglia p50/p95 distinta: il confronto viene
quindi riportato come osservazione, senza trasformarlo in uno SLA.

## Difetto documentale osservato

La guida inclusa nell'archive contiene, prima della sezione v0.5.0 corretta,
un paragrafo legacy che afferma erroneamente che l'archive includa soltanto il
profilo chat v3 e cita un file non presente. L'archive contiene invece
`configs/maestro.v0.5.0-candidate.yaml`, coerente con manifest e sezione v4.
Il difetto non altera binario o prova, ma rende ambiguo il percorso di prima
installazione. La documentazione sorgente è stata corretta per le future
pubblicazioni; l'asset v0.5.0 rimane immutato.

## Chiusura

M38 estende l'evidenza di adozione dall'ambiente WSL2 M37 a una macchina Linux
amd64 nativa CPU-only concreta. Non autorizza altri modelli/provider,
linguaggi, directory, multi-file, agent o una promessa universale su ogni
hardware Linux. L'evidenza strutturata completa è in
`milestone-38-live-runs.json`; l'ambiente è in
`milestone-38-environment.yaml`.
