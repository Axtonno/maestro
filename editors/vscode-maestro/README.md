# Maestro for VS Code — Preview

Maestro Preview exposes four explicit Maestro CLI workflows in VS Code while
leaving authority with the CLI. It does not write files, inspect model output,
approve a mutation, download software, or run when a workspace opens.

This is a local Preview for controlled trials. It is not available in the
Visual Studio Marketplace and is not a statement of general editor support.

## Qualified environment and prerequisites

The Preview is intended for a Linux extension host, including VS Code connected
through Remote WSL. The installable trial was qualified on Ubuntu 24.04 x86-64
with VS Code 1.137.0, Maestro v0.5.0, and Ollama 0.33.1. Windows-native, macOS,
and web extension hosts are not qualified.

Install and configure these separately before using the extension:

- a Maestro v0.5.0 binary executable by the VS Code extension host;
- a Maestro v4 configuration file inside the workspace;
- Ollama and the models named by that configuration;
- a trusted, local workspace with the files referenced by the configuration.

The VSIX contains none of those prerequisites and does not modify them. If this
is a new Maestro environment, run `maestro setup` in a terminal and review its
consent and model-download steps before configuring the extension.

## Install the local VSIX

Use an isolated VS Code profile for a trial when possible:

```sh
code --user-data-dir /path/to/clean-profile \
  --extensions-dir /path/to/clean-extensions \
  --install-extension /path/to/maestro-vscode-preview-0.1.0.vsix
```

The VS Code command palette action **Extensions: Install from VSIX...** is an
equivalent manual route. Installing a local VSIX does not publish it.

## Configure path resolution

Open Settings and search for `Maestro Preview`.

- **Maestro: Binary Path** (`maestro.binaryPath`) takes precedence when set.
  An absolute path is used directly; a relative path is resolved from the
  selected workspace folder. The result must be a regular executable file.
- When **Binary Path** is empty, the extension searches for a file named
  `maestro` in the exact `PATH` inherited by the extension host, in order. It
  does not invoke a shell or scan the filesystem.
- **Maestro: Config Path** (`maestro.configPath`) is required by doctor, chat,
  and mutation. Relative paths resolve from the selected workspace folder. The
  resolved regular, readable file must stay inside that folder.

In a multi-root workspace, file commands use the folder containing the active
file. Diagnostic commands use the only folder, or the folder containing the
focused file; otherwise the Preview asks you to remove the ambiguity.

**Maestro: Show Binary Identity** displays the resolved binary path and whether
it came from Settings or the extension-host `PATH`, then runs the identity
command in the terminal. The extension performs no background identity probe.
Configuration content is never read or logged.

## Recommended first run

Run the commands in this order:

1. **Maestro: Show Binary Identity** — runs `version --diagnostic`.
2. **Maestro: Run Doctor** — runs `doctor --mode all` with the resolved config.
3. **Maestro: Ask About Active File** — sends one saved, local active file and
   your question to `maestro chat`.
4. **Maestro: Replace Selected Lines** — sends one complete selection in one
   saved PHP file below `app/` to `maestro workspace replace`.

Each command runs as a process task in a new integrated terminal. Arguments are
passed directly, without a shell. The terminal is the authoritative source for
CLI output and errors.

For Controlled Mutation, read the complete CLI preview in that terminal. Deny
there to leave the workspace unchanged, or allow once there to apply exactly
the displayed single-file diff. The extension never enters a response, applies
an edit, or bypasses the CLI preview.

## Diagnostics and troubleshooting

The **Maestro** output channel contains only redacted lifecycle events: event
and command codes, status, duration, exit status, binary-path origin, and an
already-authorized workspace-relative target path. It never contains prompts,
source, diffs, provider output, secrets, the environment, home paths,
configuration content, or reconstructed commands.

Errors have a stable code and one suggested action:

| Code | What to do |
| --- | --- |
| `binary_not_found` | Set **Maestro: Binary Path**, or add `maestro` to the extension-host `PATH`, then retry identity. |
| `binary_not_executable` | Point **Binary Path** at a regular executable Maestro file. |
| `config_not_set` | Set **Maestro: Config Path** for the selected workspace. |
| `config_not_found` | Correct the configured path and verify it is a regular file. |
| `config_not_readable` | Grant the extension host read access to the config. |
| `config_outside_expected_scope` | Move or select the config inside the chosen workspace. |
| `workspace_untrusted` | Review and trust the workspace before invoking a local binary. |
| `workspace_virtual` | Reopen it through a local Linux or Remote WSL filesystem. |
| `workspace_not_open`, `workspace_ambiguous` | Open one folder or focus a file in the intended folder. |
| `file_not_local`, `file_outside_workspace` | Focus a local file contained by the intended workspace folder. |
| `file_dirty` | Save the active file and retry. |
| `selection_empty`, `selection_partial`, `selection_multiple`, `selection_invalid` | Select exactly one range made of complete lines. |
| `mutation_target_invalid` | Use one PHP file below `app/`. |
| `question_empty`, `instruction_empty` | Enter non-empty text in the corresponding prompt. |
| `question_invalid`, `instruction_invalid` | Remove line breaks or control characters from the input. |
| `command_unavailable` | Verify the extension host and retry the command from this guide. |
| `cli_exit_nonzero` | Read the terminal for the authoritative CLI failure, correct it, and retry. |
| `terminal_name_invalid` | Set a non-empty terminal name without control characters. |
| `platform_unsupported` | Use a Linux extension host or Remote WSL. |

A Controlled Mutation denial uses the CLI's intentional exit status and is
recorded as `denied`, not as `cli_exit_nonzero`.

## Uninstall

With the same profile and extensions directory used for installation:

```sh
code --user-data-dir /path/to/clean-profile \
  --extensions-dir /path/to/clean-extensions \
  --uninstall-extension maestro-local.maestro-vscode-preview
```

You can also uninstall **Maestro (Preview)** from the Extensions view. The
extension creates no workspace configuration and performs no migration or
cleanup of Maestro, Ollama, models, or user files.

## Development gates

Packaging uses the official `@vscode/vsce` 3.9.2 tool pinned by the npm
lockfile. Unit tests need Node.js 22.12.0 or later:

```sh
npm ci
npm test
npm run package:list
npm run package:vsix
```

Extension-host and installed-VSIX tests additionally require a VS Code Linux
executable:

```sh
MAESTRO_VSCODE_EXECUTABLE=/path/to/VSCode-linux-x64/code npm run test:integration
MAESTRO_VSCODE_EXECUTABLE=/path/to/VSCode-linux-x64/code npm run test:vsix
```

Marketplace publishing, publisher creation, signing, and an update channel are
intentionally absent from this Preview.
