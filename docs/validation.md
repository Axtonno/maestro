# Maestro Validation Guide

Questa guida è per maintainer, QA, supporto e release qualification. Non fa
parte dell'installazione utente.

La baseline Git della fixture serve a rendere una prova riproducibile e
misurabile: mostra l'esatto perimetro di un apply e dimostra che deny e stale
non producono effetti. Non è richiesta per installare o usare Maestro.

## Livelli operativi

| Livello | Scopo | Obbligatorio per l'utente |
| --- | --- | --- |
| Installazione e `setup` | primo utilizzo | sì |
| `doctor` | diagnosi, CI e bug report | no |
| Controlled Mutation guidata | prova del write-mode | no |
| baseline Git | verifica degli effetti | no |
| Laravel fixture | benchmark e QA | no |

## Asset e identità

Per qualificare una release, scaricare archive e checksum dalla stessa GitHub
Release, verificare SHA-256 e confrontare versione, commit, stato e manifest:

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

## Baseline Git della fixture

L'archive di qualifica non contiene metadati Git:

```sh
git -C fixtures/laravel-v1 init
git -C fixtures/laravel-v1 add .
git -C fixtures/laravel-v1 \
  -c user.name="Maestro Validation" \
  -c user.email="validation@localhost" \
  commit -m "Validation baseline"
git -C fixtures/laravel-v1 status --short
```

Lo stato iniziale deve essere pulito.

## Diagnostica completa

```sh
./maestro doctor --mode all \
  --config ./configs/maestro.v0.5.0-candidate.yaml
```

Per v0.5.0 sono attesi 14 check `pass`. Doctor non esegue completion e non
modifica il workspace.

## Controlled Mutation: deny e allow-once

```sh
./maestro workspace replace \
  --config ./configs/maestro.v0.5.0-candidate.yaml \
  --file app/Http/Controllers/OrderController.php \
  --lines 22:22 \
  "Cambia soltanto lo status HTTP da 201 a 202, preservando il resto della riga."
```

Al primo passaggio digitare `d`: il terminale atteso è
`approval_rejected`, exit code 3 e file invariato. Ripetere su una baseline
pulita e digitare `o` soltanto se la preview è esatta.

```sh
git -C fixtures/laravel-v1 diff -- \
  app/Http/Controllers/OrderController.php
git -C fixtures/laravel-v1 status --short
```

Il diff deve contenere esclusivamente `201` → `202` e lo stato deve elencare
un solo file. Una prova stale deve cambiare la sorgente tra preview e conferma,
terminare `stale_source` e conservare il contenuto concorrente senza applicare
la proposta obsoleta.

## Evidenza e gate

Una qualifica preserva almeno:

- identità e SHA-256 dell'artifact;
- versione e digest dei modelli;
- output doctor;
- preview e fingerprint;
- decisione deny o allow-once;
- digest prima e dopo;
- diff e stato Git;
- terminale, exit code e contatori di effetti vietati.

Le matrici, i manifest, i replay e i report restano documenti di milestone.
Vedere [Mutation Qualification](mutation-qualification.md),
[Packaging Candidate](packaging-candidate.md) e i
[report M39](reports/milestone-39-final.md).
