# Maestro @MAESTRO_VERSION@ Quick Start

Questo è il percorso per iniziare a usare Maestro in un progetto reale. Non è
una procedura di qualificazione della release e non richiede Git.

> Nota di disponibilità: i comandi `setup` e `mutate` sono presenti nel branch
> principale dopo v0.5.0. L'archive pubblico v0.5.0 resta immutabile; la nuova
> esperienza diventerà pubblica con una release successiva qualificata.

## Prerequisiti

- Linux `amd64`;
- [Ollama](https://ollama.com/download) installato e avviato;
- spazio sufficiente per i due modelli consigliati;
- un progetto locale, con Controlled Mutation oggi limitata a PHP sotto
  `app/`.

## 1. Installa Maestro

Scarica e installa il binario seguendo la guida di
[installazione](installation.md), quindi entra nel progetto:

```sh
cd /percorso/del/progetto
```

## 2. Configura

```sh
maestro setup
```

`setup`:

- crea `~/.config/maestro/config.yaml` con permessi `0600`;
- usa la directory corrente come workspace;
- verifica che Ollama risponda;
- verifica nome e digest dei modelli chat e mutation;
- chiede conferma prima di scaricare un modello mancante;
- non sostituisce una configurazione esistente o un digest diverso.

In un terminale non interattivo, un modello mancante produce un'indicazione
esplicita. `maestro setup --pull` autorizza il download senza prompt.

## 3. Chat

```sh
maestro chat "Come puoi aiutarmi?"
```

Senza `--file`, Maestro non cerca automaticamente nei file del progetto. Per
fare una domanda su un file specifico:

```sh
maestro chat --file app/Services/Example.php "Riassumi responsabilità e dipendenze"
```

## Prima Controlled Mutation

Scegli un file PHP sotto `app/` e un intervallo di righe completo:

```sh
maestro mutate --preview \
  --file app/Services/Example.php \
  --lines 10:12 \
  "Semplifica questo blocco senza cambiarne il comportamento"
```

La preview non richiede TTY, non chiede approvazione e non scrive. Se il diff
è corretto, ripeti senza `--preview`:

```sh
maestro mutate \
  --file app/Services/Example.php \
  --lines 10:12 \
  "Semplifica questo blocco senza cambiarne il comportamento"
```

L'apply richiede una TTY reale e una conferma allow-once sulla preview esatta.

## Se qualcosa non va

Esegui `maestro doctor --mode all` e consulta
[Troubleshooting](troubleshooting.md). I risultati e la procedura minima
riproducibile sono nella pagina [Benchmark](benchmarks.md).
