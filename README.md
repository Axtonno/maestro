# Maestro

> The intelligence is in the orchestration.

Maestro è una workstation AI locale per lo sviluppo software, costruita su un
runtime modulare. La release v0.5.0 offre un nucleo operativo verificato:
Direct Chat sul workspace e Controlled Mutation con selezione esplicita,
preview completa e approvazione umana prima di ogni scrittura.

La [pagina delle capacità correnti](docs/current-capabilities.md) è il
riferimento sintetico autorevole per supporto, hardware e limiti.

Non è un agente autonomo general-purpose. Oggi esegue due flussi piccoli,
osservabili e delimitati; l'architettura più ampia resta la direzione del
progetto, non una promessa già disponibile.

## Cosa fa oggi

| Capacità | Perimetro v0.5.0 |
| --- | --- |
| Direct Chat | Domanda senza file oppure su un solo file esplicito |
| Controlled Mutation | Sostituzione di un intervallo di righe in un singolo file PHP sotto `app/` |
| Controllo della scrittura | Preview completa, allow-once in TTY, verifica stale e apply atomico |
| Provider | Ollama locale su loopback |
| Modelli | `qwen3.5:9b` per chat; `qwen2.5-coder:14b` per mutation |
| Artifact | Release pubblica Linux `amd64` |
| Evidenza | Linux `amd64`, incluso Linux nativo CPU-only su ThinkPad T490s |

Il write-mode non sceglie autonomamente file o righe:

```text
file + intervallo + istruzione dell'utente
  -> proposta del modello dedicato
  -> validazione host-bound
  -> preview e fingerprint
  -> allow-once o deny in TTY
  -> controllo stale
  -> sostituzione atomica del solo intervallo
```

## Provalo

Prerequisiti: Linux `amd64`, Ollama 0.33.1 attivo su
`127.0.0.1:11434` e i due modelli con i digest indicati nel manifest.

```sh
version=v0.5.0
artifact="maestro-${version}-linux-amd64"
base_url="https://github.com/Axtonno/maestro/releases/download/${version}"
curl -fLO "${base_url}/${artifact}.tar.gz"
curl -fLO "${base_url}/${artifact}.tar.gz.sha256"
sha256sum -c "${artifact}.tar.gz.sha256"
tar -xzf "${artifact}.tar.gz"
cd "${artifact}"

./maestro version --diagnostic
./maestro doctor --mode all --config ./configs/maestro.v0.5.0-candidate.yaml
```

La configurazione distribuita punta alla fixture Laravel inclusa. Una prova
chat non modifica il workspace:

```sh
./maestro chat --config ./configs/maestro.v0.5.0-candidate.yaml --file app/Http/Controllers/OrderController.php "Quali campi valida store e quale risposta HTTP restituisce?"
```

Per la prova guidata del write-mode, inclusi deny e allow-once, seguire
[Installa e prova](docs/install-and-try.md).

## Cosa non fa oggi

- non individua autonomamente file, righe o modifiche;
- non modifica più file o contenuti fuori dallo span scelto;
- non supporta altri linguaggi o directory nel write-mode;
- non offre agent autonomi, retrieval multi-file o tool calling come prodotto;
- non esegue shell, Git, Docker o comandi remoti per conto del modello;
- non è una sandbox e opera con i privilegi dell'utente locale;
- non garantisce correttezza semantica del modello.

La pagina [Controlled Mutation: perimetro supportato](docs/controlled-mutation-support.md)
è il contratto sintetico del write-mode. La
[Compatibility Matrix](docs/compatibility.md) dettaglia piattaforme e modelli
entro il claim definito dalla pagina delle capacità correnti.

## Oggi e dopo

| Disponibile in v0.5.0 | Direzione successiva, non ancora supportata |
| --- | --- |
| Chat senza file o single-file | Contesto e retrieval multi-file |
| Sostituzione host-bound single-range | Pianificazione e modifiche multi-step |
| Ollama e due digest qualificati | Altri provider e modelli |
| PHP sotto `app/` | Altri linguaggi e superfici |
| Approvazione locale allow-once | Workflow e policy più articolati |

Questa distinzione è intenzionale: codice sperimentale o architettura presente
nel repository non amplia il support claim della release.

## Evidenza operativa

La Milestone 38 ha eseguito l'asset pubblico v0.5.0, senza checkout o rebuild,
su Linux `amd64` nativo CPU-only. Risultati: doctor 14/14, target e preview
8/8, sei apply esatti, deny e stale senza scritture, correttezza semantica
11/12, completion 12/12 e zero effetti vietati. L'errore semantico F02 è
documentato senza retry.

Vedere il [report M38](docs/reports/milestone-38-final.md) e le
[release notes v0.5.0](docs/releases/v0.5.0.md).

L'archive v0.5.0 è immutabile e contiene un difetto documentale legacy; il
binario e la configurazione sono corretti, mentre i sorgenti documentali sono
già corretti per le release successive. Dettagli e nome del profilo effettivo
sono nella [pagina delle capacità correnti](docs/current-capabilities.md).

## Documentazione

- [Installa e prova](docs/install-and-try.md)
- [Capacità correnti](docs/current-capabilities.md)
- [Controlled Mutation: perimetro supportato](docs/controlled-mutation-support.md)
- [Installazione completa](docs/installation.md)
- [Quick Start](docs/quick-start.md)
- [CLI](docs/cli.md)
- [Configurazione](docs/configuration.md)
- [Compatibility Matrix](docs/compatibility.md)
- [Security Model](docs/security-model.md)
- [Known Issues](docs/known-issues.md)
- [Roadmap](docs/roadmap.md)

## Sviluppo

Requisiti: Go `1.24.5` e GNU userland per il packaging riproducibile.

```sh
go test ./...
go test -race ./...
go vet ./...
```

I test live sono opt-in e `not_run` non equivale a PASS. Maestro non avvia
provider, non installa modelli e non amplia implicitamente le authority.

## Licenza

Maestro è distribuito sotto [Apache License 2.0](LICENSE). Le attribution sono
in [NOTICE](NOTICE) e [THIRD_PARTY_LICENSES.txt](THIRD_PARTY_LICENSES.txt).
