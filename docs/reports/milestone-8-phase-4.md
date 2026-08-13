# Milestone 8 — Fase 4 Packaging and Installation Report

Data: 2026-08-13

Stato: Completata

---

# Esito

Il gate della Fase 4 è superato. Esiste un packaging candidate Linux `amd64`
installabile fuori dal repository, riconducibile a un commit esatto e
riproducibile byte-for-byte dallo stesso input.

Il candidate non è una release candidate e non è candidato alla pubblicazione.
Provider, modelli e scenario agentico live devono essere verificati nella Fase
5 prima di qualsiasi promozione.

# Identità del candidate

| Campo | Valore |
|---|---|
| Versione | `v0.1.0-pc.1` |
| Artifact | `maestro-v0.1.0-pc.1-linux-amd64.tar.gz` |
| Piattaforma | `linux/amd64` |
| Commit | `4578c132682e6b715317a6b4d1de958459cfc086` |
| Go | `go1.24.5` |
| Dimensione archive | `3586154` byte |
| SHA-256 | `18d67a2a6bbeb3db2e46c8a99229fc36346b7de2567f395e3323c91bb75d8e97` |
| Licenza | `Apache-2.0` |
| Fixture | `maestro-laravel-mini@1.0.0` |

L'archive e il checksum locale sono prodotti sotto `dist/`, directory esclusa
dal controllo versione. La fonte di verità riproducibile è il commit indicato
insieme agli script versionati.

# Build e riproducibilità

`scripts/package-candidate.sh`:

- rifiuta worktree sporchi e file non tracciati;
- richiede versione semantica esplicita;
- deriva commit e `SOURCE_DATE_EPOCH` da Git;
- costruisce con `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`, `-trimpath`,
  `-buildvcs=false`, build ID vuoto e moduli read-only;
- incorpora versione e commit tramite linker flags;
- usa cache Go temporanea e toolchain locale;
- normalizza ordine, timestamp, owner, group e mode dell'archive;
- genera checksum SHA-256 senza sovrascrivere artifact esistenti.

`scripts/verify-package-candidate.sh` ha eseguito due build in directory
temporanee indipendenti. `cmp` conferma archive e checksum identici; entrambe le
build producono lo SHA-256 registrato sopra.

# Contenuto dell'archive

Il contenuto è stato enumerato e ispezionato. Include:

- binario `maestro` eseguibile;
- `ARTIFACT-MANIFEST.txt` con versione, commit, piattaforma, toolchain,
  licenza, fixture e stato `packaging-candidate`;
- `LICENSE`, `NOTICE` e `THIRD_PARTY_LICENSES.txt`;
- README e guide preliminari di installazione, configurazione, CLI e UX;
- `configs/maestro.example.yaml`, strict `version: 1` e senza secret;
- fixture Laravel versionata con `artisan`, `composer.json`, sorgenti e
  `dataset.json`.

La fixture non contiene symlink, `.git`, `.env`, `vendor` o `node_modules` e
non installa dipendenze. L'archive non scarica modelli, plugin o package.

# Licenza

Il maintainer ha scelto Apache License 2.0, registrata in ADR-0027. Il testo
ufficiale è nel file root `LICENSE` e nell'archive. `NOTICE` e
`THIRD_PARTY_LICENSES.txt` conservano attribution e termini dual MIT/Apache di
`gopkg.in/yaml.v3 v3.0.1`, unica dipendenza inclusa nel binario corrente.

# Verifica fuori dal checkout

Il gate ha estratto l'archive in una directory temporanea e ha verificato:

- `maestro version` restituisce `v0.1.0-pc.1` e il commit esatto;
- root help restituisce la superficie CLI attesa;
- `doctor` carica la configurazione inclusa e rileva la fixture Laravel;
- con provider deliberatamente non disponibile, `doctor` termina con exit
  code 1, mantiene verdi config/workspace/composition/agent/tool/policy/Laravel
  e classifica il provider come non disponibile;
- il binario installato in un prefisso temporaneo esegue version e help da una
  working directory vuota, senza accesso al repository;
- configurazione e fixture usano soltanto path relativi interni all'archive.

La guida `docs/installation.md` documenta verifica checksum, installazione
utente, configurazione, upgrade e rimozione senza installer privilegiato o
modifiche implicite del sistema.

# Redazione e sicurezza dell'artifact

Le verifiche automatiche rifiutano:

- path assoluti o traversal nell'archive;
- path del workspace di build nei file estratti o nel binario;
- private key e token con forme canoniche AWS, GitHub o OpenAI;
- symlink e directory di dipendenze nella fixture;
- artifact prodotti da stato Git sporco o file non tracciati.

Nessuna credenziale o path locale è stato rilevato. Il candidate conserva il
security boundary trusted in-process: packaging e checksum non costituiscono
sandbox o firma crittografica dell'autore.

# Gate eseguiti

```text
./scripts/verify-package-candidate.sh --version v0.1.0-pc.1
sha256sum -c maestro-v0.1.0-pc.1-linux-amd64.tar.gz.sha256
go test -count=3 ./...
go test -race ./...
go vet ./...
go test -run '^$' -bench BenchmarkAgentLoopDeterministic -benchmem ./internal/agent
git diff --check
```

Tutti i gate sono verdi. Il benchmark finale osservato è circa `40477 ns/op`,
`15517 B/op` e `219 allocs/op` sulla macchina di verifica; resta un indicatore
locale, non un budget di release.

# Limiti residui

- l'unica piattaforma prodotta e verificata è Linux `amd64`;
- l'archive non è firmato e non è pubblicato;
- Ollama, `llama3.1:8b` e lo scenario reference agent live non sono stati
  eseguiti in questa fase;
- il debito documentale/live llama.cpp resta assegnato alla Fase 5;
- security model pubblico, compatibility matrix e note finali appartengono
  ancora alla Fase 6;
- non esistono tag o artifact finali `v0.1.0`.

# Verdetto

**GO alla Fase 5 — Validazione live e release candidate.**

Il packaging candidate può essere usato come input dei gate live, ma non deve
essere presentato come release-ready finché la Fase 5 non è conclusa.
