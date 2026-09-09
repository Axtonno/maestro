# Benchmark e capacità verificate

Questa pagina pubblica risultati sintetici e riproducibili. Report intermedi,
tentativi falliti, selezione dei modelli e roadmap di qualifica restano
documentazione interna.

## Baseline pubblica v0.5.0

| Verifica | Risultato pubblicato |
| --- | --- |
| Piattaforma | Linux `amd64`, incluso un ciclo CPU-only nativo |
| Diagnostica | 14/14 controlli superati |
| Target e preview | 8/8 casi conformi |
| Apply autorizzati | 6 sostituzioni esatte |
| Deny e stale | Nessuna scrittura |
| Correttezza semantica | 11/12 risposte |
| Completion | 12/12 richieste terminate |
| Effetti fuori perimetro | 0 osservati |

Questi numeri descrivono la matrice pubblicata, non una garanzia universale.
Il caso semanticamente errato riguardava una domanda generica senza file; i
controlli di contenimento non hanno prodotto scritture indesiderate.

## Riproduzione minima

Scaricare archive e checksum dalla stessa GitHub Release, quindi verificare
l’identità prima dell’esecuzione:

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
./maestro doctor --mode all \
  --config ./configs/maestro.v0.5.0-candidate.yaml
```

`doctor` non esegue completion e non modifica il workspace. Le prove live
dipendono da hardware, modelli e provider e devono riportare distintamente
`PASS`, `FAIL` e `NOT_RUN`.

## Interpretazione

- i benchmark misurano soltanto combinazioni dichiarate di artifact,
  piattaforma, provider e modello;
- una compilazione riuscita non equivale a supporto;
- latenza e memoria possono cambiare sensibilmente tra CPU e GPU;
- sicurezza del write path e correttezza semantica sono misure separate;
- i limiti correnti restano quelli della [matrice delle capacità](current-capabilities.md).
