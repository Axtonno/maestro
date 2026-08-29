# Milestone 19 — Post-Release Adoption & Lower-Bound Validation

Versione osservata: v0.3.0

Stato: Completata — `operationally_impractical`

Data: 2026-08-29

Documenti di riferimento:

- `releases/v0.3.0.md`;
- `compatibility.md`;
- `installation.md`;
- `cli.md`;
- `reports/milestone-18-final.md`;
- `reports/milestone-19-thinkpad-adoption.md`.

## Obiettivo

Misurare se l'esatto asset pubblico v0.3.0 è già utile sul ThinkPad usato
quotidianamente, attraverso un progetto Laravel reale e domande Direct Chat
single-file. La milestone raccoglie esperienza post-release: non qualifica il
ThinkPad come reference hardware e non modifica il support claim ufficiale.

## Confine congelato

Sono inclusi soltanto download dalla GitHub Release, verifica dell'archive,
installazione in un prefix isolato, `doctor --mode chat`, una domanda senza
file, cinque domande single-file reali, una prova `--stream`, misure di
latenza/qualità e verifica pre/post del workspace.

Restano fermi verified agent, multi-file, Controlled Mutation, nuovi provider,
altri modelli ufficialmente supportati e qualsiasi modifica al codice o alla
configurazione distribuita. Non sono autorizzati tuning, pull di modelli,
rebuild, write/patch, tool calling o fallback agentico.

Il progetto reale viene indicato soltanto come `project-a`. Path fisico,
remote, prompt, risposte complete e contenuti non entrano nei documenti
committati. Le evidenze grezze locali usano permessi `0600`.

## Identità e ambiente

| Campo | Valore congelato |
|---|---|
| release | `v0.3.0` |
| commit incorporato | `3f4c7d4b4fd2e380644cf250ce9e8fec2311af53` |
| SHA-256 archive | `6c8f0e883ec8f8c05571fc2e7bc1f4ecac608c2bd7e338395ae0a4253fff1aaf` |
| SHA-256 binario | `378a0533083b9a00be6c0212ca52001cebc5f77b476a20038bc8e08d1fc3d42d` |
| piattaforma osservata | ThinkPad T490s, Ubuntu 24.04, Linux `amd64`, CPU-only |
| provider | Ollama `0.32.14`, loopback |
| modello/digest | `qwen3.5:9b` / `6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7` |
| profilo | chat v2 distribuito; timeout 5 minuti; context 4096; thinking false |

## Matrice minima

| ID | Prova | Criterio |
|---|---|---|
| `M19-I0` | download, checksum, archive e installazione fuori checkout | identità completa coerente e binario eseguibile |
| `M19-D0` | `doctor --mode chat` | cinque check PASS |
| `M19-C0` | domanda senza file | completion valida e nessuna invenzione sul progetto |
| `M19-C1` | route reale | endpoint, controller e action corretti |
| `M19-C2` | controller reale | validazione, persistenza ed esiti corretti |
| `M19-C3` | service reale | rami, transazione ed error handling corretti |
| `M19-C4` | service minimo reale | selezione e costruzione del risultato corrette |
| `M19-C5` | model minimo reale con `--stream` | metadata e relazione corretti; terminale valido |
| `M19-W0` | stato e digest workspace pre/post | identici |

Ogni domanda è one-shot, usa zero tool e può leggere al massimo un file logico
esplicito. Una run senza risposta valida è `unevaluable`, non `incorrect`. Il
timeout resta un failure operativo e non viene esteso dopo una run negativa.

## Metriche

Per ogni run si registrano exit code, terminale/reason code, durata, token di
input/output, modalità complete/stream, dimensione del file e giudizio
`correct`, `partial`, `incorrect` o `unevaluable`. Si annotano inoltre tempo
del doctor, checksum dell'artifact, stato Git e digest aggregato del workspace.

Non viene calcolata qualità sulle run prive di risultato. La latenza viene
riportata sia per tutte le run sia per le sole completion, senza nascondere i
deadline.

## Verdetti

| Verdetto | Regola operativa |
|---|---|
| `usable` | 5/5 single-file completed, qualità almeno 4/5 utile, nessun incidente e latenza compatibile con uso quotidiano |
| `functional_but_slow` | percorso affidabile e utile, ma attese sistematicamente elevate senza timeout ripetuti |
| `operationally_impractical` | timeout/failure ripetuti o attese che impediscono il normale ciclo domanda-risposta, anche con risposte corrette quando disponibili |

Qualunque modifica inattesa del workspace costituisce stop immediato e
incidente separato. Nessun verdetto della milestone amplia hardware, modello o
capability supportati.

## Esito

La matrice è stata eseguita integralmente sul ThinkPad con l'asset pubblico.
Il report redatto emette `operationally_impractical`: il percorso è corretto
quando completa, ma due delle cinque domande reali raggiungono il deadline di
300 secondi e le completion riuscite restano troppo lente per un uso
interattivo quotidiano.
