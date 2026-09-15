# Install or locate Maestro

The extension uses the Maestro CLI already present on the Linux extension host.
It never downloads a binary or model.

For the pinned v0.5.0 candidate:

```sh
install -d "$HOME/.local/bin" && curl -fsSL https://github.com/Axtonno/maestro/releases/download/v0.5.0/maestro-v0.5.0-linux-amd64.tar.gz | tar -xz --strip-components=1 -C "$HOME/.local/bin" maestro-v0.5.0-linux-amd64/maestro
maestro setup
```

Run **Show Binary Identity** from this step. The step completes only when the
passive executable check succeeds.
