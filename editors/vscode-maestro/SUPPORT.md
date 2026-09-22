# Support

## Supported candidate

The current unpublished candidate is `axtonno.maestro-local-ai` 0.4.0. Its
execution contract targets local Linux and VS Code Remote WSL extension hosts,
with VS Code 1.100.0 or later and the current post-v0.5.0 Maestro source
candidate documented in the README. The prior 0.3.0 candidate passed the live
Remote WSL gate; clean Extension Development Host and installed-VSIX gates for
the 0.4.0 targets are still pending.

Windows-native, macOS, web, Dev Containers, Remote SSH, Codespaces, tunnels,
and unknown remote extension hosts are not supported by this candidate. Dev
Containers and Remote SSH are planned and fail explicitly rather than inheriting
a generic Linux claim. Support is best effort while the extension remains in
the 0.x pre-release series.

## Ask for help or report a defect

Use the [Maestro issue tracker](https://github.com/Axtonno/maestro/issues) for
non-sensitive defects. Include:

- extension version and VS Code version;
- extension-host OS and the `local-linux` or `remote-wsl` target shown by
  `@maestro /status`;
- `Maestro: Show Binary Identity` output after removing local paths;
- the stable extension error code and the redacted `Maestro` output events;
- minimal reproduction steps with synthetic files.

Do not include prompts, proprietary source, diffs, API keys, configuration
contents, model output, home paths, or other secrets. Report vulnerabilities
through the private route in [SECURITY.md](SECURITY.md).

## Version, update, and rollback policy

- Versions use `major.minor.patch`; SemVer suffixes are not used because the
  Marketplace pre-release channel does not support them.
- `0.1.x` established Marketplace readiness, `0.2.x` added guided onboarding,
  `0.3.x` added native Chat plus workspace adaptation, and `0.4.x` freezes the
  remote execution contract and fail-closed claim matrix. Every Marketplace
  upload, if later authorized, receives a new patch version.
- VS Code automatic updates will be accepted only after the private/public
  release milestone authorizes publication. This extension never implements
  a separate updater or downloads the Maestro CLI.
- Breaking setting or command changes require at least a minor version bump
  and migration notes. Stable command IDs are preferred.
- Before publication, rollback means uninstalling the candidate and installing
  the previously qualified VSIX. After publication, a bad release is replaced
  by a higher patch containing the revert; Marketplace artifacts are not
  overwritten.

To avoid an unwanted candidate update during a controlled trial, disable
automatic updates for this extension in VS Code before installing the VSIX.
