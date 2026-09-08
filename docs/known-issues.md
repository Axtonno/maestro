# Maestro v0.5.0 Known Issues and Limitations

La [pagina delle capacità correnti](current-capabilities.md) è il riferimento
sintetico per ciò che v0.5.0 supporta oggi.

## Archive pubblico v0.5.0

- L'archive immutabile contiene un paragrafo documentale legacy sullo schema
  v3 e non evidenzia abbastanza il nome del profilo distribuito.
- Il profilo operativo è `configs/maestro.v0.5.0-candidate.yaml`; binario,
  manifest e configurazione sono corretti.
- I sorgenti documentali sono corretti per le release successive. L'asset
  v0.5.0 non viene sostituito o ripubblicato.

## Qualità del modello

- I modelli sono generativi: temperatura zero riduce il drift ma non garantisce
  correttezza semantica.
- M38 ha registrato 11/12 richieste corrette. F02 ha trattato erroneamente una
  domanda PHP generale senza file come informazione non determinabile dal
  workspace.
- Preview e contenimento corretti non dimostrano che la modifica sia quella
  desiderata: l'utente deve leggere il diff prima dell'allow-once.
- `num_ctx_effective` e `thinking_effective` possono risultare `unknown`
  quando Ollama non fornisce un'attestazione per-run.

## Prestazioni e memoria

- Latenza e memoria dipendono da hardware, carico e residenza del modello.
- I due profili non devono essere residenti insieme: Maestro scarica il modello
  uscente prima di caricare quello richiesto.
- Sul T490s CPU-only M38 ha osservato chat tra circa 64 e 172 secondi. Non è uno
  SLA né un requisito minimo.
- M38 ha osservato 323.584 byte di swap aggiuntivo senza OOM o restart. La
  qualifica non promette zero swap su ogni macchina.

## Perimetro del workspace

- Direct Chat usa zero o un solo file esplicito; non effettua retrieval.
- Controlled Mutation usa un solo file PHP sotto `app/` e un unico intervallo
  inclusivo. Non supporta multi-file, directory, glob o target automatici.
- Symlink, traversal, path assoluti, file non regolari, UTF-8 invalido e input
  oltre limite vengono rifiutati.
- Dopo un apply riuscito non esiste rollback automatico. Usare Git e verificare
  i test del progetto.

## Autorità e isolamento

- Maestro non è una sandbox e usa i privilegi dell'utente locale.
- Domanda e file selezionato vengono inviati al provider configurato.
- Il sorgente è trattato come input non attendibile, ma non esiste una difesa
  perfetta contro prompt injection o output ingannevole.
- L'allow-once autorizza esclusivamente la preview corrente; non certifica il
  significato della modifica.
- Non esistono shell, Git, Composer, Docker o comandi remoti eseguiti dal
  percorso supportato.

## Prodotto ed ecosistema

- È pubblicato soltanto l'artifact Linux `amd64`.
- Ollama locale e i due digest v0.5.0 sono l'unica combinazione qualificata.
- CLI e schema sono versionati ma possono cambiare durante la serie 0.x.
- Non esistono installer di sistema, auto-update o service unit.
- Maestro non installa Ollama, modelli, PHP, Composer o dipendenze Laravel.
- Agent, retrieval multi-file, tool calling, altri provider e plugin di terze
  parti restano direzioni architetturali, non capacità di prodotto.

Vedere [Capacità correnti](current-capabilities.md) per il claim autorevole,
[Compatibility Matrix](compatibility.md) per il dettaglio,
[Controlled Mutation](controlled-mutation-support.md) per il write-mode e
[Troubleshooting](troubleshooting.md) per le azioni operative.
