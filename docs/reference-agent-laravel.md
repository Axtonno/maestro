# Reference Agent Laravel v0.1.x

## Promessa di prodotto

Il reference agent supportato analizza, interroga e aiuta a comprendere un
progetto Laravel locale in modalità read-only. Non esegue Artisan, Composer,
PHPUnit, shell, Git, Docker o modifiche al workspace.

## Detection

La root deve contenere:

- un file regolare `artisan`;
- un `composer.json` valido entro 1 MiB;
- una dipendenza non vuota `laravel/framework` in `require`.

La fixture inclusa dichiara PHP `^8.3` e Laravel `^12.0`. La detection non
costituisce una matrice di compatibilità per tutte le versioni Laravel: verifica
i marker strutturali e rende disponibile il workspace generico.

## Percorso operativo

`maestro run`:

1. carica il plugin Laravel trusted in-process;
2. ottiene il workspace autorevole;
3. indicizza i file ammessi dalla scan policy;
4. costruisce contesto lessicale bounded;
5. crea una run con provider, modello, policy, tool e limiti espliciti;
6. consente al modello soltanto list/read/search;
7. stampa progresso redatto su stderr e risultato finale su stdout.

La root fisica non viene inserita nel prompt. I tool usano path logici relativi
e rifiutano traversal e symlink.

La scan policy Laravel è distinta dalla policy filesystem generica. Include
`app`, `bootstrap`, `config`, `database`, `lang`, le aree sorgente di
`resources`, `routes`, `tests` e i manifest principali; esclude asset generati
in `public`, dati runtime in `storage`, dipendenze e directory nascoste. Ogni
file sorgente è bounded a 2 MiB e lo snapshot resta bounded a 64 MiB.

## Configurazione ufficiale

```yaml
workspace:
  id: laravel
  root: /absolute/path/to/project
  framework: laravel

agent:
  id: agent.reference
  streaming: true
  tools:
    - workspace.list
    - workspace.read
    - workspace.search

policy:
  id: policy.local-review
  model: allow
  workspace_inspect: allow
  workspace_mutate: deny
```

Usare `configs/maestro.example.yaml` come file completo e validarlo con
`maestro doctor`.

## Task adatti

- spiegare controller, service e route;
- trovare riferimenti testuali o file correlati;
- descrivere il flusso di una richiesta;
- individuare dipendenze visibili nel codice;
- riassumere una porzione del progetto entro i budget configurati.

Il modello può sbagliare o produrre una tool call invalida. Verificare sempre le
conclusioni importanti sul codice sorgente.

## Fuori dal supporto

- scrittura o patch dei file;
- esecuzione di test o comandi del framework;
- comprensione completa di runtime reflection, service container o stato DB;
- memoria persistente e recovery;
- più agenti o workspace nella stessa run;
- garanzie su modelli diversi da `llama3.1:8b`.

I dettagli tecnici del plugin sono in `laravel-plugin.md`; i limiti pubblici
sono in `compatibility.md` e `known-issues.md`.
