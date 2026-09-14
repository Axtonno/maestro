# Security

Maestro for VS Code delegates all runtime and mutation authority to the
separately installed Maestro CLI. The extension requires Workspace Trust,
does not run on activation, uses no shell or network API, writes no workspace
files, and cannot approve a mutation. Preview, allow-once, stale checking, and
atomic replacement remain CLI responsibilities in the integrated terminal.

Report a vulnerability with GitHub's private **Report a vulnerability** route
for [Axtonno/maestro](https://github.com/Axtonno/maestro/security). If that
route is unavailable, open a public issue without sensitive detail and ask the
maintainer for a private contact.

Never publish API keys, real workspace content, prompts, diffs, provider
output, complete configuration, exploit payloads, or local paths. Include the
extension version, Maestro binary identity, VS Code version, extension-host
platform, impact, and a minimal redacted reproduction.
