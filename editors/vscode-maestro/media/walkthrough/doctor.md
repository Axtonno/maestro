# Verify the local runtime

**Maestro: Doctor** runs `doctor --mode all` in a new integrated terminal. It
uses the explicit workspace configuration when set; otherwise the CLI uses
`MAESTRO_CONFIG`, `XDG_CONFIG_HOME`, or `~/.config/maestro/config.yaml`.

The step completes only after Doctor exits successfully. Read the terminal for
the authoritative diagnostic result.
