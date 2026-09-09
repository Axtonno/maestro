# Installa e prova Maestro

Il percorso utente è stato spostato nel [Quick Start](quick-start.md):

```sh
maestro setup
maestro chat "Come puoi aiutarmi?"
maestro mutate --preview --file app/Services/Example.php --lines 10:12 \
  "Semplifica questo blocco senza cambiarne il comportamento"
```

Questa pagina resta come compatibilità per i link pubblicati prima della
Milestone 41. Per problemi usare [Troubleshooting](troubleshooting.md); per
qualificazione, replay e prove di release usare la
[Validation Guide](validation.md).
