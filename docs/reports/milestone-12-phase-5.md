# Milestone 12 — Phase 5 Report

Data: 2026-08-21

Stato: **COMPLETATA**

## Oggetto della qualificazione

La qualificazione live ha usato il percorso supportato Linux `amd64`, Ollama,
`llama3.1:8b`, reference agent Laravel e soli tool read-only. Provider e
modelli erano già disponibili: la prova non ha avviato servizi, scaricato
modelli o modificato la configurazione del provider.

Candidate e release candidate sono archive distinti, installati fuori dal
checkout e identificati da versione, commit, stato manifest e checksum.

## Serie dei packaging candidate

| Candidate | Commit | SHA-256 | Esito live | Classificazione |
|---|---|---|---|---|
| `v0.2.0-pc.1` | `7d3f45ee0268fc758b9e3722e57c91e486065615` | `e5f98bedcb94ab40236d3f315cf9af0be976825abbd2d9a6ea756ad26200fc13` | una run senza read osservata | prodotto |
| `v0.2.0-pc.2` | `5ed48eaade1b102f3720ea30c27bef7851f614cc` | `5d7287a60a987422d9ab57e141fa1751bbee878227a0aee4d7a9a5044546bd74` | arguments read invalidi | prodotto/modello |
| `v0.2.0-pc.3` | `a4709e27a1ef666656af246c5c38020464bf19e6` | `1bbb1470bcded3d9bbe45de781d87c2a24a34c2e4cdf4bc3709e63035ecb9649` | deadline dopo recovery | prodotto/modello |
| `v0.2.0-pc.4` | `e240afef2be7ed9bb3b253a96defba1f86b5083b` | `273846ec54f99949470619967736f90f4ba9541db2b2eff5e0a982aa2f2c5e2d` | deadline senza tool call nativa | modello |
| `v0.2.0-pc.5` | `e8aaad800f1a72eb395f895ba5c8b54195ce0388` | `d504050720a538a549fe1451fbde87848707a1cbc3a1afd2a782e67498cfe7b8` | due quick start consecutivi | **PASS** |

Ogni failure ha interrotto la serie del relativo candidate. Gli archive
respinti restano immutabili e non sono stati rinominati o promossi. Le
correzioni successive richiedono sempre un nuovo `pc.N`.

L'hardening risultante:

- impedisce una risposta finale alle istruzioni esplicite di lettura senza una
  `workspace.read` riuscita;
- rende recuperabili, entro gli stessi hard limit, gli arguments invalidi di
  un tool read-only;
- limita a `workspace.read` il tool set mostrato prima della read richiesta;
- per la grammatica stretta `Read <logical-path> ...`, esegue la read
  deterministica attraverso Tool Runtime prima della prima inferenza.

L'ultimo punto non aggira il runtime: applica policy, containment, limiti,
contatori ed eventi della normale invocazione, quindi consegna al modello un
risultato correlato. Nessun contenuto testuale che simuli una tool call viene
eseguito.

## Packaging candidate qualificato

`v0.2.0-pc.5` supera doppio build byte-identico, checksum, audit archive,
installazione esterna al checkout, identità del binario e `doctor` 9/9.

Le due run consecutive completano rispettivamente in 212052 ms e 14874 ms.
Entrambe osservano esattamente un turno modello e una read, producono una
risposta semanticamente corretta sul servizio chiamato dal controller e
lasciano la fixture byte-identica alla baseline.

## Release candidate congelato

| Campo | Valore |
|---|---|
| Versione | `v0.2.0-rc.1` |
| Stato manifest | `release-candidate` |
| Commit | `e8aaad800f1a72eb395f895ba5c8b54195ce0388` |
| Piattaforma | Linux `amd64` |
| SHA-256 | `056f557abe0b95a3a1d758b8827e04907500a988b719e2af9a6ddbfb24886fab` |

Il release candidate è stato costruito due volte in modo byte-identico dallo
stesso commit qualificato, ma come artifact distinto dal packaging candidate.
L'installazione pulita conferma manifest, `maestro version` e `doctor` 9/9.

Le due run consecutive sull'esatto archive RC completano in 6350 ms e
20244 ms. Entrambe registrano un turno modello, una read, risposta
semanticamente corretta e fixture invariata.

## Gate live sul release candidate

| Scenario | Exit | Terminale | Turni/tool | Wall time |
|---|---:|---|---:|---:|
| SIGINT dopo 3 secondi | 130 | `canceled` | 1/1 | 2996 ms |
| deadline run 1 secondo | 130 | `deadline_exceeded` | 1/1 | 1001 ms |
| hard limit `model_turns: 1` | 1 | `limit_exceeded` | 1/1 | 167900 ms |

I terminali sono bounded e autorevoli. Nel caso hard limit il messaggio CLI
finale resta il sintetico `execution_failed`, già dichiarato nei known issues.
Nessuna prova ha mostrato approval o capability mutative.

## Immutabilità e anti-leak

La fixture dell'installazione RC coincide byte-per-byte con la baseline del
repository dopo quick start e controlli negativi. L'audit automatico
dell'archive è negativo per path del checkout, credenziali-shaped, token
documentali irrisolti, symlink e superficie mutativa pubblicata.

Il report conserva soltanto identità, checksum, contatori, terminal reason,
durate e valutazione semantica. Prompt completi, response complete, contenuti
del workspace, arguments tool, endpoint fisici e secret non sono pubblicati.

## Gate

**PASS.** `v0.2.0-rc.1` è il release candidate immutabile qualificato. Supera
due quick start consecutivi, cancellazione, deadline, hard limit, immutabilità
e anti-leak. Non restano blocker per l'ingresso nella Fase 6; l'RC non verrà
rinominato o riutilizzato come artifact finale.
