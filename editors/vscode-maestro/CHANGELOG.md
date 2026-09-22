# Changelog

All notable changes to Maestro for VS Code are recorded here. Versions follow
the policy in [SUPPORT.md](SUPPORT.md).

## 0.4.0 — remote execution contract candidate

- Freeze the extension as a workspace extension and expose the effective
  execution target as `local-linux` or `remote-wsl` in Chat and redacted
  lifecycle diagnostics.
- Resolve binary, configuration, workspace paths, processes, and the approval
  terminal exclusively in the current extension host; a binary available only
  on the UI host is never treated as remotely available.
- Fail closed with actionable errors when a supported remote is forced onto the
  UI host or when the workspace uses Dev Containers, Remote SSH, Codespaces,
  tunnels, or an unknown remote authority.
- Keep Dev Containers and Remote SSH explicitly planned until each has passed a
  clean Extension Development Host and installed-VSIX gate.
- Preserve Workspace Trust, virtual-workspace rejection, redacted output,
  CLI-only mutation authority, and the absence of stored secrets.

## 0.3.0 — Native Chat and workspace adaptation candidate

- Add the single native `@maestro` Chat assistant with `/status`, `/preview`,
  and `/doctor` commands, without a hard GitHub Copilot extension dependency.
- Show the effective qualified profile, Direct Chat model, Controlled Mutation
  model, workspace, mode, and file/selection context in Chat.
- Derive the project root from single-root, active-file multi-root, or explicit
  multi-root selection and pass it through the CLI `--workspace-current`
  contract; generic no-folder chat remains read-only.
- Add bounded, cancellable, no-shell CLI capture for profile identity, Doctor,
  and chat while keeping mutation preview/approval in a real terminal TTY.
- Expose only the `recommended` two-model profile and add actionable setup,
  provider, model, timeout, workspace, file, and selection failures.
- Keep the candidate local and unpublished with no automatic CLI/model
  downloads, telemetry, direct edits, or auto-apply.

## 0.2.0 — VS Code onboarding candidate

- Add a native five-step walkthrough with verifiable completion for CLI
  discovery, Doctor, path recovery, focused chat, and reviewed mutation.
- Add one passive status bar item with `Ready`, `Config missing`, and `Binary
  missing` states, each linked to its next recovery action.
- Rename Command Palette labels to the stable M49 vocabulary while preserving
  the four M47 command IDs, and add `maestro.openSetupGuide`.
- Document setting defaults, scopes, examples, and precedence without exposing
  the rejected `maestro.profile` setting.
- Preserve the CLI-only runtime, terminal approval boundary, zero-write
  extension authority, and absence of telemetry or automatic downloads.

## 0.1.1 — Marketplace readiness candidate

- Freeze the prospective public identity as `axtonno.maestro-local-ai` and
  the display name as **Maestro for VS Code**.
- Add Marketplace metadata, a 256×256 PNG icon, public support and security
  guidance, and asset provenance.
- Let an empty `maestro.configPath` use the CLI configuration created by
  `maestro setup`, preserving explicit workspace-contained paths when set.
- Package as an unpublished pre-release VSIX and add an upgrade path to the
  installed-VSIX qualification harness.
- Keep the four commands, Linux/Remote WSL claim, terminal approval boundary,
  and zero-write extension authority unchanged.

## 0.1.0 — Local Preview

- Package the four-command Maestro surface as a locally installable VSIX.
- Resolve the binary from the explicit setting or the extension-host `PATH`
  without invoking a shell.
- Resolve explicit workspace-scoped configuration paths deterministically.
- Add stable, actionable errors and a redacted `Maestro` output channel.
- Keep chat, preview, deny, and allow-once in an integrated terminal controlled
  by the Maestro CLI.

No version has been uploaded to the Visual Studio Marketplace.
