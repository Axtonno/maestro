# Maestro for VS Code

Maestro for VS Code adds `@maestro` to the native VS Code Chat and keeps
Controlled Mutation behind the Maestro CLI terminal approval boundary.

Version 0.4.0 is an unpublished local candidate. It is installable from a
VSIX, but it is not yet available in the Visual Studio Marketplace. The
extension never writes project files, approves a mutation, downloads software,
or starts provider/model work when a workspace opens.

## Requirements

- a trusted workspace on a Linux extension host: local Linux or Remote WSL;
- VS Code 1.100.0 or later with Chat enabled;
- Ollama 0.33.1 and the two models selected by `maestro setup`;
- the current post-v0.5.0 Maestro source candidate, including `profile` and
  `--workspace-current`.

The immutable v0.5.0 archive predates the M50 workspace/profile contract. To
test this local candidate from the repository, build the current CLI in WSL:

```sh
go build -o "$HOME/.local/bin/maestro" ./cmd/maestro
maestro setup
```

The 0.3.0 candidate was qualified on Ubuntu 24.04 x86-64 in Remote WSL. The
0.4.0 execution contract is implemented for Remote WSL and local Linux, but
its clean Extension Development Host and installed-VSIX gates remain pending
for the declared targets. Windows-native, macOS, web, Dev Containers, Remote
SSH, Codespaces, tunnels, and unknown remote extension hosts are not qualified
by this candidate. The VSIX contains no Maestro binary, provider, model, user
configuration, or telemetry client.

## Install and open Chat

Install the local candidate from a VS Code window connected to WSL:

```sh
code --install-extension maestro-local-ai-0.4.0.vsix --force
```

Run **Maestro: Open Chat** from the Command Palette, or open the VS Code Chat
view and type `@maestro`. The Chat surface provides:

- `@maestro <question>` for local Direct Chat;
- `@maestro /status` for effective profile, model, workspace, and setup health;
- `@maestro /preview <instruction>` to open the selected-line Controlled
  Mutation flow in a terminal;
- `@maestro /doctor` to run the full Doctor in a terminal.

Every response starts with the effective profile, Direct Chat model,
Controlled Mutation model, extension host target, selected workspace, mode,
and active file context.
Model output is rendered as untrusted text. The extension never records prompts,
source, responses, or configuration contents in the Maestro output channel.

## Workspace adaptation

The project root is not a setting. Maestro uses the VS Code context:

| VS Code state | Behavior |
| --- | --- |
| One folder | Uses that folder automatically |
| Multi-root with an active file | Uses the folder containing that file |
| Ambiguous multi-root | Asks for one folder |
| File outside the workspace | Blocks contextual chat and mutation |
| Dirty or unsaved file | Blocks file context until it is saved |
| No open folder | Allows generic chat; disables mutation |
| Remote WSL | Uses paths and processes from the WSL extension host |
| Dev Container | Fails with `dev_container_unqualified`; planned until a clean container gate exists |
| Remote SSH or another remote | Fails with `remote_unsupported`; no generic Linux claim |

The extension passes `--workspace-current` and starts the CLI with the selected
folder as its working directory. The CLI validates that override and remains
authoritative for path containment, model identity, stale checks, preview, and
apply.

The manifest fixes `extensionKind` to `workspace`. In Remote WSL, the extension,
CLI process, configuration lookup, workspace paths, and integrated task terminal
therefore live in the distro. Local UI-host paths and `PATH` entries are never
searched. If VS Code is forced to run Maestro on the UI host while a supported
remote is open, commands fail with `extension_host_mismatch`.

## Settings

- `maestro.binaryPath` (machine-overridable, default empty): optional executable
  in the current extension-host environment, for example
  `/home/me/.local/bin/maestro`. Configure it independently for local and Remote
  WSL settings; an explicit value takes precedence over that host's `PATH`.
- `maestro.configPath` (resource scope, default empty): optional workspace
  configuration, for example `./.maestro/config.yaml`. When empty, the CLI uses
  `MAESTRO_CONFIG`, `XDG_CONFIG_HOME`, then
  `~/.config/maestro/config.yaml`.
- `maestro.profile` (resource scope): `recommended`, the qualified two-model
  profile. The effective CLI identity must match this value. The rejected
  single-model evaluation profile is not exposed.
- `maestro.terminalName`: name of the dedicated integrated terminal.

One passive status item shows `Ready`, `Config missing`, or `Binary missing`.
Its tooltip names the effective extension host. It reads only host-local
metadata and never runs the CLI or provider at startup.

## Controlled Mutation and security

Chat is a read-only direct CLI process with bounded output, timeout, and
cancellation. The question is sent on standard input instead of the process
argument list. Classic commands and Controlled Mutation use a direct
`ProcessExecution` task in a new integrated terminal.

For mutation, review the complete CLI preview in the terminal. Deny to keep the
workspace unchanged, or allow once to apply exactly the displayed single-file
diff. The extension cannot enter the approval, call `workspace.applyEdit`, or
bypass Maestro's stale-source check.

## Troubleshooting

| State | Single next action |
| --- | --- |
| Maestro binary missing | Open Maestro Settings |
| Configuration missing | Open Setup Guide |
| CLI profile/workspace contract missing | Open Troubleshooting and build the current CLI |
| Ollama unreachable | Run Doctor |
| Direct Chat model/digest missing | Run Doctor |
| Controlled Mutation model/digest missing | Run Doctor |
| Workspace unsupported or ambiguous | Select/open a supported workspace |
| File dirty | Save File |
| Selection invalid | Select one range of complete lines |
| Request timeout | Run Doctor |

The `Maestro` output channel contains only redacted lifecycle fields. See
[SECURITY.md](SECURITY.md) and [SUPPORT.md](SUPPORT.md) before reporting an
issue.

The extension stores no credentials or secrets. If a future capability needs
one, it must use VS Code SecretStorage and pass a separate security gate; plain
workspace/global state is not an allowed secret store.

## Updates and removal

No update is downloaded or installed by the extension. Before Marketplace
publication, install a newer signed-off VSIX explicitly with `--force`.

```sh
code --uninstall-extension axtonno.maestro-local-ai
```

Uninstalling the extension does not remove Maestro, Ollama, models,
configuration, or project files.

## Development

```sh
npm ci --ignore-scripts
npm test
npm run package:list
npm run package:vsix
```

Publishing credentials and a `publish` script are intentionally absent. Asset
origin and license are recorded in [ASSET_PROVENANCE.md](ASSET_PROVENANCE.md).
