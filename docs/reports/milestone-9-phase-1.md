# Milestone 9 — Report Fase 1

Data: 2026-08-15

Stato: **COMPLETATA — contratto di osservazione congelato**

## Risultato

La campagna post-release dispone di artifact, ambiente, profilo live, regole di
classificazione e criteri di stop definiti prima delle nuove run. La baseline
statica precedente resta un input e non viene promossa a prova live.

## Artifact congelato

| Campo | Valore |
|---|---|
| Artifact | `maestro-v0.1.0-linux-amd64.tar.gz` |
| Versione | `v0.1.0` |
| Commit incorporato | `f882919798fa6073bc11c6af18a431bf249a7755` |
| SHA-256 | `c785676a177165a2c11ff0fc744931ac8b5d923466155ec32365e7a0c03d271f` |
| Dimensione | 3.604.828 byte |
| Manifest | `release` |
| Piattaforma | Linux `amd64` |

L'archive alternativo con SHA-256
`5ad3e297e28033868488c42a3ff58e47a44d393f6c830cc33085a461cc564124`
è la build rifiutata già descritta dalla Milestone 8 e non entra nella
campagna.

## Profilo live congelato

| Campo | Valore |
|---|---|
| Sistema | Linux `amd64` |
| CPU | Intel Core i5-8365U, 4 core / 8 thread |
| RAM / swap | 15 GiB / 4 GiB |
| Provider | Ollama 0.32.5 su `127.0.0.1:11434` |
| Modello chat | `llama3.1:8b` |
| Retrieval | lessicale |
| Profilo | `agent.reference`, read-only |
| Tool | `workspace.list`, `workspace.read`, `workspace.search` |
| Policy | `workspace_mutate: deny` |
| Limite run | 10 minuti |
| Run ripetute | 2 per profilo positivo |

Ollama viene avviato esplicitamente dall'operatore. Maestro non avvia server e
non esegue pull di modelli.

## Workspace

La fixture inclusa nell'artifact viene usata soltanto nella Fase 2. La Fase 3
usa due repository Laravel locali distinti, registrati nei report come
`real-laravel-a` e `real-laravel-b`. I path fisici e il contenuto restano fuori
dai report; entrambi contengono `artisan`, `composer.json` e directory `app` e
sono accessibili in sola lettura dal profilo ufficiale.

## Matrice congelata

| Area | Scenario | Criterio |
|---|---|---|
| Artifact | checksum, inventory, estrazione pulita | identità esatta e nessun membro pericoloso |
| CLI | version/help/doctor/models/agents | exit e output coerenti con la v0.1.0 |
| Fixture | due quick start consecutivi | `completed`, read reale, risposta corretta |
| Workspace reali | due run per workspace | almeno una read/search reale e terminale coerente |
| Immutabilità | stato prima/dopo ogni run | nessuna modifica fisica |
| Protocollo modello | pseudo-call e call invalide | failure bounded e classificata |
| Controlli | SIGINT, deadline, hard limit | terminale ed exit code documentati |
| Privacy | scansione degli output conservati | nessun secret, root fisica o contenuto completo |

Una run è `PASS` soltanto quando soddisfa il criterio predefinito. Un
prerequisito assente è `SKIP`, un comportamento osservato contrario al gate è
`FAIL` e un errore del runner è distinto dal failure del prodotto.

## Registro delle osservazioni

Ogni osservazione conserva soltanto:

- ID e timestamp;
- artifact e profilo;
- scenario ed exit code;
- terminale, reason code, turni, tool call e durata;
- stato del workspace prima e dopo;
- classe: bug v0.1.x, limite modello, ambiente, UX o evolutiva;
- severità, riproducibilità, impatto e destinazione.

Prompt, risposte complete, secret, root fisiche e contenuto dei workspace non
vengono conservati.

## Gate

- artifact e profilo identificati senza fallback: superato;
- matrice e criteri fissati prima delle run: superato;
- workspace reali separati dalla fixture: superato;
- immutabilità e anti-leak inclusi nel protocollo: superato;
- nessuna modifica applicativa: superato.

La Fase 1 è completata. La Fase 2 può iniziare.
