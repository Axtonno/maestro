# Milestone 20 — Fase C: correzioni operative indipendenti

Data: 2026-08-30

Stato: **COMPLETATA — `operational_corrections_ready`**

## Confine

Le correzioni sono state implementate soltanto dopo la chiusura delle serie
prestazionali A/B. Non cambiano quei binari, payload, prompt, modelli o misure
e non vengono attribuite retroattivamente a v0.3.0.

Il risultato è un candidate di sviluppo per la futura qualifica CPU, non un
nuovo asset pubblico.

## C1 — Errori di configurazione specifici

`internal/productconfig` espone ora una diagnostica tipizzata e redatta:

| Categoria | Caso |
|---|---|
| `read_failed` | file assente o non leggibile |
| `yaml_invalid` | YAML malformato, duplicato o strutturalmente invalido |
| `unknown_field` | chiave estranea allo schema strict |
| `missing_field` | campo obbligatorio assente |
| `invalid_value` | campo presente con valore non valido |

Il loader percorre lo schema YAML e conserva il path logico completo anche per
chiavi annidate. Le validation error possiedono il field owner; la presenza
nel documento distingue missing da invalid value.

Chat e doctor mantengono exit 2 e i reason code pubblici. L'output aggiuntivo
è allowlisted:

```text
chat failed: invalid_request
configuration	kind=unknown_field field=interaction.chat.example
```

Non vengono stampati valore, path fisico del config, secret, errore decoder o
testo remoto. Chiavi non conformi al formato sicuro non vengono riflesse
nell'output.

## C2 — Identità del binario

L'output storico di `maestro version` resta invariato. Il nuovo comando
esplicito:

```text
maestro version --diagnostic
```

restituisce versione, status incorporato, commit, dirty flag, path eseguibile
risolto attraverso i symlink e SHA-256 del file realmente aperto. Un errore di
risoluzione o hashing termina fail-closed con
`version failed: identity_unavailable`, senza propagare il path o l'errore
originario.

`internal/buildinfo` incorpora ora anche `Status`. Lo script di packaging lo
imposta via linker flag e il verifier confronta status, path e hash con il
binario estratto prima dei gate operativi.

## C3 — Heartbeat redatto

Il servizio Direct Chat notifica l'inizio della generation soltanto dopo
file loading e preflight. La CLI arma in quel punto un heartbeat su stderr:

```text
progress	state=generating elapsed_ms=15000
```

Proprietà congelate:

- cadenza 15 secondi;
- massimo 40 righe, sufficiente a coprire il ceiling di 10 minuti;
- stato e tempo trascorso sono gli unici campi;
- zero domanda, modello, file/path, risposta/chunk, secret o errore remoto;
- stdout resta atomico;
- `Stop` chiude il ticker e attende la goroutine prima di stampare risultato o
  failure;
- start e stop sono idempotenti e non lasciano heartbeat dopo il terminale.

## Test e sicurezza

I test coprono:

- cinque categorie di configurazione e path annidato;
- redazione di value/path sentinella in chat e doctor;
- output normale e diagnostico della versione;
- failure dell'identità senza leak;
- status development, release, unknown e dirty;
- heartbeat presente durante una completion lenta, bounded, redatto e fermo
  dopo il terminale;
- race detector sul package CLI;
- sintassi degli script di packaging.

La suite completa, race globale e vet appartengono al gate finale della
milestone e sono registrati nel report finale.

## Verdetto

Verdetto C: **`operational_corrections_ready`**. Le tre correzioni sono pronte
per entrare in un nuovo candidate e in un artifact di qualificazione. Non
promuovono `qwen2.5-coder:7b` e non modificano v0.3.0.

