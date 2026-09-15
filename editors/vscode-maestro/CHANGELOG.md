# Changelog

All notable changes to Maestro for VS Code are recorded here. Versions follow
the policy in [SUPPORT.md](SUPPORT.md).

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
