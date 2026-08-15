# Milestone 9 — Report Fase 2

Data: 2026-08-15

Stato: **COMPLETATA — artifact e preflight fuori dal checkout superati**

## Installazione pulita

La prova parte da una nuova directory `/tmp` contenente soltanto l'archive
finale. Il checkout non viene usato da configurazione, fixture o binario.

| Controllo | Esito |
|---|---|
| SHA-256 | `c785676a177165a2c11ff0fc744931ac8b5d923466155ec32365e7a0c03d271f` |
| Dimensione | 3.604.828 byte |
| Path assoluti o componenti `..` | Assenti |
| Symlink o hard link | Assenti |
| Estrazione | Superata |
| Manifest | `release`, Linux `amd64`, v0.1.0 |
| Versione binario | v0.1.0, commit `f882919798fa6073bc11c6af18a431bf249a7755` |

L'archive contiene soltanto binario, documentazione pubblica, licenze,
configurazione e fixture dichiarate dal manifest.

## CLI e preflight

Root help e help di `doctor`, `models`, `agents`, `run`, `version` e `bench`
sono coerenti con la CLI pubblicata.

Con accesso al server Ollama locale, `doctor` riporta 9 check `pass`, inclusi
`capabilities_11`, capability richieste dal modello e riconoscimento Laravel.
`models` include `llama3.1:8b` ed `embeddinggemma:latest`; `agents` espone
`agent.reference` 1.0.0 con planning, run e workspace awareness.

Il primo `doctor` eseguito nel sandbox non poteva raggiungere il socket locale
e ha riportato due failure di probe. La medesima installazione, ripetuta con
accesso locale esplicito, supera 9/9: l'osservazione è classificata **ambiente
di esecuzione**, non bug Maestro.

## Quick start consecutivi

I due run usano l'istruzione pubblicata e la configurazione inclusa senza
modifiche.

| Run | Exit | Terminale | Turni | Tool call | Token in/out | Durata |
|---:|---:|---|---:|---:|---:|---:|
| 1 | 0 | `completed` | 2 | 1 | 2886 / 99 | 332.255 ms |
| 2 | 0 | `completed` | 2 | 1 | 2887 / 52 | 52.569 ms |

Entrambe le run leggono realmente il controller e identificano
`OrderService::create`. Non compare alcuna approval mutativa.

Il digest aggregato della fixture è
`f5318094c365a9a634d4a983e86691ff5e84f83a96ba5b7f844818759847a250`
prima della prima run, tra le due run e dopo la seconda.

## Privacy

Nei dati conservati non compaiono secret, path dei workspace reali o contenuti
estranei al piccolo controller pubblico incluso nell'archive. Gli ID di run
effimeri non sono necessari alla riproduzione e non vengono promossi a
contratto.

## Gate

- artifact sicuro e identità esatta: superato;
- installazione senza checkout: superata;
- CLI e preflight: superati;
- due quick start consecutivi: superati;
- workspace immutato: superato;
- anti-leak: superato;
- bug v0.1.x osservati: nessuno.

La Fase 2 è completata. La Fase 3 può iniziare.
