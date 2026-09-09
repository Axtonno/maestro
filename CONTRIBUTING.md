# Contribuire a Maestro

Grazie per l’interesse verso Maestro. Le contribution pubbliche devono
migliorare il prodotto osservabile senza ampliare implicitamente il perimetro
di supporto.

## Prima di iniziare

- apri una issue per cambiamenti di comportamento o nuove superfici pubbliche;
- mantieni compatibili i contratti esistenti oppure documenta la rottura;
- non aggiungere prompt, report grezzi, dati di utenti, roadmap tattiche o
  artefatti di selezione dei modelli;
- usa dati sintetici e redatti nei test e negli esempi.

## Verifica locale

```sh
go test ./...
go test -race ./...
go vet ./...
```

I test che richiedono Ollama o hardware specifico devono essere opt-in,
riproducibili e distinguere chiaramente `PASS`, `FAIL` e `NOT_RUN`.

## Pull request

Una pull request dovrebbe contenere:

- problema e perimetro della soluzione;
- test automatici proporzionati al cambiamento;
- aggiornamento delle guide pubbliche se cambia il comportamento utente;
- dichiarazione esplicita di piattaforme, provider o modelli non verificati;
- nessun dato sensibile o percorso locale nei log e nelle fixture.

Per vulnerabilità non aprire una issue pubblica: seguire [SECURITY.md](SECURITY.md).
