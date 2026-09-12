'use strict';

const ERROR_CATALOG = Object.freeze({
  binary_not_executable: ['Maestro binary is not executable', 'The configured path is not a regular executable file.', 'Open Settings'],
  binary_not_found: ['Maestro binary not found', 'No executable Maestro binary was found in Settings or the extension host PATH.', 'Open Settings'],
  cli_exit_nonzero: ['Maestro command failed', 'The Maestro CLI exited with a non-zero status. Its terminal contains the authoritative details.', 'Open Preview Guide'],
  command_unavailable: ['Maestro command unavailable', 'VS Code could not start the requested Maestro task.', 'Open Preview Guide'],
  config_not_found: ['Maestro config not found', 'The configured file does not exist or is not a regular file.', 'Open Settings'],
  config_not_readable: ['Maestro config is not readable', 'The extension host cannot read the configured file.', 'Open Settings'],
  config_not_set: ['Maestro config is not set', 'Set Maestro: Config Path for this workspace before running this command.', 'Open Settings'],
  config_outside_expected_scope: ['Maestro config is outside the workspace', 'The configured file must resolve inside the selected workspace folder.', 'Open Settings'],
  file_dirty: ['Save the active file', 'Maestro only receives a saved workspace file.', 'Save File'],
  file_not_local: ['Local file required', 'Open a local file in the selected workspace folder.', 'Open Preview Guide'],
  file_outside_workspace: ['Workspace file required', 'The active file is outside the selected workspace folder.', 'Open Preview Guide'],
  instruction_empty: ['Mutation instruction required', 'Enter a non-empty replacement instruction.', 'Open Preview Guide'],
  instruction_invalid: ['Mutation instruction is invalid', 'The instruction must not contain line breaks or control characters.', 'Open Preview Guide'],
  mutation_target_invalid: ['Mutation target is not allowed', 'Controlled Mutation requires one saved PHP file below app/.', 'Open Preview Guide'],
  platform_unsupported: ['Extension host not supported', 'This Preview supports Linux extension hosts, including Remote WSL.', 'Open Preview Guide'],
  question_empty: ['Question required', 'Enter a non-empty question.', 'Open Preview Guide'],
  question_invalid: ['Question is invalid', 'The question must not contain line breaks or control characters.', 'Open Preview Guide'],
  selection_empty: ['Selection required', 'Select one or more complete lines before running Controlled Mutation.', 'Open Preview Guide'],
  selection_invalid: ['Selection is invalid', 'Select exactly one valid range of complete lines.', 'Open Preview Guide'],
  selection_multiple: ['One selection required', 'Controlled Mutation accepts exactly one selection.', 'Open Preview Guide'],
  selection_partial: ['Complete lines required', 'Select whole lines; Maestro replaces one inclusive line range.', 'Open Preview Guide'],
  terminal_name_invalid: ['Terminal name is invalid', 'Use a non-empty terminal name without control characters.', 'Open Settings'],
  workspace_ambiguous: ['Choose a workspace folder', 'Focus a local file in the workspace folder Maestro should use.', 'Open Preview Guide'],
  workspace_not_open: ['Open a workspace folder', 'Open one local workspace folder before running Maestro.', 'Open Preview Guide'],
  workspace_untrusted: ['Trust required', 'Trust this workspace before invoking the local Maestro binary.', 'Manage Workspace Trust'],
  workspace_virtual: ['Virtual workspace not supported', 'Open the workspace through a local Linux or Remote WSL file system.', 'Open Preview Guide']
});

class PreviewError extends Error {
  constructor(code) {
    const entry = ERROR_CATALOG[code];
    if (!entry) {
      throw new Error(`unknown Maestro Preview error code: ${code}`);
    }
    super(entry[1]);
    this.name = 'PreviewError';
    this.code = code;
    this.title = entry[0];
    this.cause = entry[1];
    this.action = entry[2];
  }
}

function normalizeError(error) {
  return error instanceof PreviewError ? error : new PreviewError('command_unavailable');
}

module.exports = { ERROR_CATALOG, PreviewError, normalizeError };
