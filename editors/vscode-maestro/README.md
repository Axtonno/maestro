# Maestro for VS Code

Maestro for VS Code runs four explicit [Maestro](https://github.com/Axtonno/maestro)
CLI workflows from the editor: binary identity, doctor, chat about the active
file, and Controlled Mutation of selected lines.

This is an unpublished pre-release candidate. It is installable from a local
VSIX for qualification, but it is not yet available in the Visual Studio
Marketplace. The extension never writes files itself, approves a mutation,
downloads software, or starts work when a workspace opens.

## Requirements

- a trusted local workspace on a Linux extension host, including Remote WSL;
- VS Code 1.85.0 or later;
- Ollama 0.33.1 and the models selected by `maestro setup`;
- Maestro v0.5.0 on the extension-host `PATH`.

The qualified extension environment is Ubuntu 24.04 x86-64 with VS Code
Stable. Windows-native, macOS, web, Dev Containers, and Remote SSH extension
hosts are not qualified by this candidate. The VSIX contains no Maestro
binary, provider, model, user configuration, or telemetry client.

## First run

From the project you want Maestro to use, install the pinned CLI and run setup:

```sh
install -d "$HOME/.local/bin" && curl -fsSL https://github.com/Axtonno/maestro/releases/download/v0.5.0/maestro-v0.5.0-linux-amd64.tar.gz | tar -xz --strip-components=1 -C "$HOME/.local/bin" maestro-v0.5.0-linux-amd64/maestro
maestro setup
```

The first line is the one supported short installation command for this
candidate. It installs only the pinned v0.5.0 binary and never invokes
`sudo`. The checksum-oriented archive procedure remains available in the
[v0.5.0 release notes](https://github.com/Axtonno/maestro/blob/master/docs/releases/v0.5.0.md).
`maestro setup` owns configuration and any consent for model downloads; the
extension does not reproduce either operation.

Install the local candidate into VS Code:

```sh
code --install-extension maestro-local-ai-0.1.1.vsix
```

Then open the Command Palette and run these commands in order:

1. **Maestro: Show Binary Identity**
2. **Maestro: Run Doctor**
3. **Maestro: Ask About Active File**
4. **Maestro: Replace Selected Lines**

The first two commands establish which binary and runtime are in use. Chat
passes one saved local file chosen by you. Controlled Mutation accepts one
complete selection in a saved PHP file below `app/`.

## Settings

- `maestro.binaryPath`: optional executable path. A relative path resolves
  from the selected workspace; when empty, the exact extension-host `PATH` is
  searched without a shell.
- `maestro.configPath`: optional v4 configuration path. When empty, Maestro
  uses the default created by `maestro setup`. A relative explicit path
  resolves inside the selected workspace.
- `maestro.terminalName`: name of the dedicated integrated terminal.

In multi-root workspaces, file commands use the folder containing the active
file. Other commands require an unambiguous focused folder.

## Controlled Mutation and security

Every command runs as a direct process task with separate arguments in a new
integrated terminal. The terminal is authoritative for CLI output and errors.

For mutation, review the complete CLI preview in that terminal. Deny to leave
the workspace unchanged, or allow once to apply exactly the displayed
single-file diff. The extension cannot enter the approval, apply an edit, or
bypass Maestro's stale-source check.

The `Maestro` output channel records only redacted lifecycle fields. It never
records prompts, file contents, diffs, provider output, secrets, configuration
contents, home paths, or the environment. See [SECURITY.md](SECURITY.md) for
the reporting route and trust boundary.

## Troubleshooting

| Error | Next action |
| --- | --- |
| `binary_not_found` | Set `maestro.binaryPath` or expose `maestro` on the extension-host `PATH`. |
| `binary_not_executable` | Select a regular executable Maestro file. |
| `config_not_found`, `config_not_readable` | Correct the explicit config path or its permissions. |
| `config_outside_expected_scope` | Use the CLI default or move the explicit config into the workspace. |
| `workspace_untrusted` | Review and trust the workspace before invoking a local binary. |
| `workspace_virtual`, `platform_unsupported` | Use a local Linux or Remote WSL extension host. |
| `file_dirty` | Save the active file and retry. |
| `selection_empty`, `selection_partial`, `selection_multiple` | Select exactly one range of complete lines. |
| `mutation_target_invalid` | Select a PHP file below `app/`. |
| `cli_exit_nonzero` | Read the terminal, correct the CLI failure, and retry. |

A mutation denial is reported as `denied`, not as a failure. More help and the
information to include in a report are in [SUPPORT.md](SUPPORT.md).

## Updates and removal

No update is downloaded or installed by the extension. Before Marketplace
publication, upgrade by installing a newer signed-off VSIX with
`code --install-extension <file.vsix> --force`.

Remove this candidate with:

```sh
code --uninstall-extension axtonno.maestro-local-ai
```

Uninstalling the extension does not remove Maestro, Ollama, models,
configuration, or project files. Versioning, update, and rollback policy are
documented in [SUPPORT.md](SUPPORT.md).

## Development

Packaging uses the lockfile-pinned `@vscode/vsce`:

```sh
npm ci --ignore-scripts
npm test
npm run package:list
npm run package:vsix
```

Publishing credentials and a `publish` script are intentionally absent. Asset
origin and license are recorded in [ASSET_PROVENANCE.md](ASSET_PROVENANCE.md).
