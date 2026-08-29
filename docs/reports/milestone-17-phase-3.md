# Milestone 17 — Fase 3: contesto esplicito single-file

Data: 2026-08-29

Stato: **COMPLETATA — PASS**

Baseline di ingresso: commit `7251326`

## Obiettivo verificato

Soltanto il singolo file logico scelto dall'utente può entrare nella request
provider. La validazione è locale al servizio Direct Chat, il path è confinato
tramite `os.Root`, ogni componente symlink viene respinto e il file è riletto
prima della disclosure per rilevare sostituzioni o mutazioni concorrenti.

## Contratto del path

Sono accettati soltanto path UTF-8 relativi e già normalizzati. Sono respinti:

- stringa vuota, `.`, `..`, prefisso assoluto e traversal;
- componenti `.`/`..`, separatori duplicati e backslash;
- NUL, caratteri di controllo, formattatori Unicode invisibili e separatori di
  linea/paragrafo;
- workspace root symlink, symlink interni o evasivi in qualsiasi componente;
- directory, FIFO, device, file assente e file non regolare;
- file oltre limite, UTF-8 invalido o contenente NUL;
- cambio di identità, mode, size, mtime o contenuto durante la doppia lettura;
- sostituzione del file, della directory padre o della root workspace.

Il path logico viene inserito nel messaggio di sistema con quoting JSON. Il
path fisico non viene serializzato.

## Contratto del contenuto

- il limite `max_file_bytes` è inclusivo;
- un file vuoto è una prova valida ma priva di fatti e viene preservato;
- un BOM UTF-8 è ammesso e preservato byte per byte;
- il contenuto è un messaggio user separato e interamente non attendibile;
- non vengono aggiunti marker `BEGIN/END` collidibili con il contenuto;
- il messaggio di sistema dichiara che anche istruzioni o delimitatori apparenti
  nel file sono soltanto evidenza;
- senza `--file` non vengono eseguiti scan, discovery o letture workspace.

## Matrice deterministica

| Gruppo | Scenari | Esito |
|---|---|---|
| normalizzazione | vuoto, `.`, `..`, assoluto, traversal, slash duplicato, backslash | PASS fail-closed |
| caratteri | newline, tab, NUL, bidi/formattatore invisibile | PASS fail-closed |
| tipo fisico | directory, FIFO, missing, symlink interno/evasivo, root symlink | PASS fail-closed |
| contenuto | vuoto, BOM UTF-8, limite esatto | PASS e preservazione byte-esatta |
| contenuto invalido | UTF-8 invalido, NUL, limite + 1 | PASS fail-closed |
| race | contenuto, replacement, chmod, symlink replacement | PASS fail-closed |
| gerarchia race | directory padre e root sostituite | PASS fail-closed |
| provider I/O su rifiuto | capability, complete e stream | 0 |
| prompt boundary | path quoted, payload raw in messaggio separato | PASS |

## Immutabilità

Il digest della fixture autorevole `routes/api.php` prima e dopo la matrice è
rimasto:

`7e224d7e57bf0be6d2618e668d8515b07d332f94b7960de4640e4388b31bbc39`.

I test del servizio confrontano inoltre snapshot ricorsivi di path, tipo, mode
e contenuto per terminali positivi e negativi. Nessun file temporaneo o write
appartiene al percorso Direct Chat.

## Invalidazione live controllata

Il passaggio da sentinelle testuali a confini di messaggio e il quoting del path
modificano il prompt template. I PASS live della Milestone 15 restano evidenza
di fattibilità e seed del profilo, ma non qualificano il nuovo candidate. La
Fase 6 deve ripetere C0, C1 e streaming sul prompt congelato dopo le Fasi 4–5.

## Verifiche

| Comando | Esito |
|---|---|
| `go test ./...` | PASS |
| `go vet ./...` | PASS |
| `go test -race -count=3 ./internal/directchat ./cmd/maestro` | PASS |
| `go test -count=10 ./internal/directchat` | PASS |
| `git diff --check` | PASS |

## Gate di uscita

- sola disclosure single-file esplicita: **PASS**;
- evasione e race prima del provider: **PASS**;
- nessun file significa nessuna lettura workspace: **PASS**;
- fixture e workspace byte-identici: **PASS**;
- matrice loader positiva/negativa: **PASS**.

Verdetto della fase: **PASS**. La Fase 4 può iniziare.
