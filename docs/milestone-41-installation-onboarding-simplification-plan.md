# Milestone 41 — Installation & Onboarding Simplification

Stato: **in corso — `implementation_complete_validation_pending`**

## Obiettivo

Ridurre il primo avvio di Maestro a quattro azioni comprensibili:

```text
1. installa
2. maestro setup
3. maestro chat
4. maestro mutate --preview
```

Installazione, diagnosi e qualificazione sono percorsi diversi. Il percorso
utente non richiede repository Git, dataset di benchmark, manifest di replay o
conoscenza dei gate interni.

## Confini

- `setup` può creare una configurazione e acquisire modelli soltanto dopo
  consenso dell'utente;
- non installa né avvia servizi e non usa `sudo`;
- non sovrascrive configurazioni o modelli con digest inatteso;
- non riusa silenziosamente per un altro progetto una configurazione esistente;
- `mutate --preview` deve avere zero authority di scrittura;
- `doctor` resta read-only ed è documentato come troubleshooting;
- `workspace replace` resta compatibile;
- l'asset pubblico v0.5.0 e le evidenze M38/M39 restano immutabili.

## Deliverable

- README orientato al risultato;
- Quick Start su un progetto reale;
- `docs/validation.md` separata;
- Troubleshooting con `doctor` riposizionato;
- `maestro setup` idempotente;
- `maestro mutate --preview` fail-closed e non interattivo;
- contratti CLI e documentali;
- aggiornamento del packaging per includere la Validation Guide.

`maestro demo` non è nel gate minimo: potrà essere aggiunto solo se genera un
esempio isolato, rifiuta collisioni e non diventa un prerequisito del Quick
Start. Anche un installer `curl | sh` richiede un endpoint, una policy di
version pinning e una verifica checksum prima di essere pubblicizzato.

## Gate

La milestone può chiudere soltanto quando:

- i test unitari provano creazione `0600`, idempotenza e no-overwrite;
- Ollama irraggiungibile, modello mancante e digest errato falliscono con esito
  stabile;
- i download richiedono prompt TTY o `--pull` esplicito;
- la preview produce diff e `effect=unchanged` senza TTY e senza scrittura;
- README, installazione e Quick Start non introducono materiale di qualifica;
- la Validation Guide contiene baseline, deny, stale, digest e verifica Git;
- link, test, race detector, vet e packaging riproducibile sono verdi;
- un clean install del nuovo artifact completa setup → chat → preview.

Lo stato corrente è `implementation_complete_validation_pending`: codice,
documentazione e test isolati sono presenti, ma packaging e clean install live
non sono ancora evidenza conclusiva.

La matrice autorevole è
[milestone-41-installation-onboarding-matrix.yaml](milestone-41-installation-onboarding-matrix.yaml).
