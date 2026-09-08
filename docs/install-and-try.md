# Installa e prova Maestro v0.5.0

Questa è la via più breve per verificare Direct Chat e il write-mode
controllato sulla fixture inclusa nell'archive. Non richiede il checkout del
repository.

Il support claim è riassunto nella [pagina delle capacità correnti](current-capabilities.md):
v0.5.0 è qualificata su Linux `amd64`, incluso il reference nativo CPU-only,
e Controlled Mutation resta confinata al perimetro testato con approvazione
obbligatoria.

## Prerequisiti

- Linux `amd64`;
- Ollama 0.33.1 già attivo su `http://127.0.0.1:11434`;
- `qwen3.5:9b` e `qwen2.5-coder:14b` già installati;
- digest dei modelli uguali a quelli in `ARTIFACT-MANIFEST.txt`;
- un terminale interattivo per Controlled Mutation.

Maestro non installa o avvia Ollama e non scarica modelli.

Se i modelli non sono presenti, installarli con Ollama:

```sh
ollama pull qwen3.5:9b
ollama pull qwen2.5-coder:14b
ollama list
```

I tag sono mutabili: confrontare i digest mostrati con
`ARTIFACT-MANIFEST.txt` dopo l'estrazione e fermarsi se non coincidono.

## 1. Scarica e verifica

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
```

Versione, stato `release`, commit e SHA-256 devono coincidere con il
manifest. L'hash atteso dell'archive è
`0afcfe4d648edcde3caf4327c4f995606fb4c3974c05606e13f90dd8cff321d9`.
Fermarsi se una verifica diverge.

## 2. Crea una baseline Git della fixture

L'archive non contiene metadati Git. Per rendere osservabile ogni effetto:

```sh
git -C fixtures/laravel-v1 init
git -C fixtures/laravel-v1 add .
git -C fixtures/laravel-v1 -c user.name="Maestro Trial" -c user.email="trial@localhost" commit -m "Trial baseline"
git -C fixtures/laravel-v1 status --short
```

L'ultimo comando non deve produrre output.

## 3. Esegui doctor

```sh
./maestro doctor --mode all --config ./configs/maestro.v0.5.0-candidate.yaml
```

La configurazione punta alla fixture Laravel inclusa. Tutti i 14 check devono
essere `pass`. Doctor non esegue completion e non modifica file.

## 4. Prova Direct Chat

```sh
./maestro chat --config ./configs/maestro.v0.5.0-candidate.yaml --file app/Http/Controllers/OrderController.php "Quali campi valida store e quale risposta HTTP restituisce?"
```

La risposta deve identificare `customer_id`, `items` e la risposta JSON con
status 201. Verificare sempre la risposta sul sorgente: il modello resta
generativo.

## 5. Prova una mutation senza scrivere

```sh
./maestro workspace replace --config ./configs/maestro.v0.5.0-candidate.yaml --file app/Http/Controllers/OrderController.php --lines 22:22 "Cambia soltanto lo status HTTP da 201 a 202, preservando il resto della riga."
```

Controllare file, righe, digest, fingerprint e diff, poi digitare `d`. Il
comando deve terminare con `approval_rejected` e il file deve restare
invariato.

Per provare la scrittura, ripetere il comando e digitare `o` soltanto se la
preview è esatta. Maestro ricontrolla la sorgente e applica esclusivamente
l'intervallo mostrato.

Verificare subito l'effetto:

```sh
git -C fixtures/laravel-v1 diff -- app/Http/Controllers/OrderController.php
git -C fixtures/laravel-v1 status --short
```

Il diff deve mostrare soltanto `201` → `202` sulla riga selezionata e lo stato
deve elencare soltanto `app/Http/Controllers/OrderController.php`.

## Progetto reale

Copiare la configurazione, proteggere i permessi e cambiare soltanto
`workspace.root`:

```sh
install -d "$HOME/.config/maestro"
install -m 0600 ./configs/maestro.v0.5.0-candidate.yaml "$HOME/.config/maestro/v0.5.0.yaml"
```

Eseguire di nuovo `doctor --mode all` prima di operare. Usare un repository
Git pulito e verificare il diff dopo ogni apply.

Per confini e garanzie leggere
[Controlled Mutation: perimetro supportato](controlled-mutation-support.md).
Per diagnosi più dettagliate vedere [Troubleshooting](troubleshooting.md).

Non sono promessi multi-file, agent autonomi, Windows nativo, provider o
modelli alternativi.
