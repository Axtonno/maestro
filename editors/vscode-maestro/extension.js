'use strict';

const vscode = require('vscode');
const commands = require('./command-builder');
const { formatEvent } = require('./diagnostics');
const { PreviewError, normalizeError } = require('./errors');
const { resolveBinaryPath, resolveConfigPath } = require('./path-resolver');

const GUIDE_ACTION = 'Open Preview Guide';
const SETTINGS_ACTION = 'Open Settings';
let outputChannel;
let extensionContext;

function settingsFor(scope) {
  const configuration = vscode.workspace.getConfiguration('maestro', scope);
  return {
    binaryPath: configuration.get('binaryPath', ''),
    configPath: configuration.get('configPath', ''),
    terminalName: configuration.get('terminalName', 'Maestro')
  };
}

function requireSupportedHost() {
  if (!vscode.workspace.isTrusted) {
    throw new PreviewError('workspace_untrusted');
  }
  if (process.platform !== 'linux') {
    throw new PreviewError('platform_unsupported');
  }
}

function activeWorkspaceEditor(options = {}) {
  const editor = vscode.window.activeTextEditor;
  if (!editor || editor.document.uri.scheme !== 'file') {
    throw new PreviewError('file_not_local');
  }
  const folder = vscode.workspace.getWorkspaceFolder(editor.document.uri);
  if (!folder) {
    throw new PreviewError('file_outside_workspace');
  }
  requireLocalFolder(folder);
  if (editor.document.isDirty) {
    throw new PreviewError('file_dirty');
  }
  const logicalPath = options.mutation
    ? commands.mutationLogicalPath(folder.uri.fsPath, editor.document.uri.fsPath)
    : commands.safeLogicalPath(folder.uri.fsPath, editor.document.uri.fsPath);
  return { editor, folder, logicalPath };
}

function resolveRuntime(folder, options = {}) {
  const settings = settingsFor(folder.uri);
  const binary = resolveBinaryPath({
    configuredPath: settings.binaryPath,
    workspaceRoot: folder.uri.fsPath,
    hostPath: process.env.PATH,
    hostCwd: process.cwd()
  });
  const config = options.configRequired
    ? resolveConfigPath(settings.configPath, folder.uri.fsPath)
    : undefined;
  return {
    binaryPath: binary.path,
    binaryOrigin: binary.origin,
    configPath: config && config.path,
    configLogicalPath: config && config.logicalPath,
    configOrigin: config && config.origin,
    terminalName: validateTerminalName(settings.terminalName)
  };
}

async function askActiveFile(providedQuestion) {
  requireSupportedHost();
  const target = activeWorkspaceEditor();
  const question = typeof providedQuestion === 'string'
    ? providedQuestion
    : await vscode.window.showInputBox({
      title: 'Maestro: Ask About Active File',
      prompt: `Question for ${target.logicalPath}`,
      ignoreFocusOut: true,
      validateInput: value => value.trim() === '' ? 'Enter a question.' : undefined
    });
  if (question === undefined) {
    return;
  }
  if (question.trim() === '') {
    throw new PreviewError('question_empty');
  }
  const runtime = resolveRuntime(target.folder, { configRequired: true });
  showResolution(runtime);
  await launch(target.folder, runtime, commands.buildChatInvocation(runtime, target.logicalPath, question), {
    command: 'chat',
    logicalPath: target.logicalPath
  });
}

async function replaceSelection(providedInstruction) {
  requireSupportedHost();
  const target = activeWorkspaceEditor({ mutation: true });
  if (target.editor.selections.length !== 1) {
    throw new PreviewError('selection_multiple');
  }
  requireWholeLineSelection(target.editor);
  const lines = commands.inclusiveSelectedLines(target.editor.selection);
  const instruction = typeof providedInstruction === 'string'
    ? providedInstruction
    : await vscode.window.showInputBox({
      title: 'Maestro: Replace Selected Lines',
      prompt: `Instruction for ${target.logicalPath}:${lines}`,
      placeHolder: 'Describe only the replacement for the selected lines.',
      ignoreFocusOut: true,
      validateInput: value => value.trim() === '' ? 'Enter a mutation instruction.' : undefined
    });
  if (instruction === undefined) {
    return;
  }
  if (instruction.trim() === '') {
    throw new PreviewError('instruction_empty');
  }
  const runtime = resolveRuntime(target.folder, { configRequired: true });
  showResolution(runtime);
  await launch(target.folder, runtime, commands.buildMutationInvocation(runtime, target.logicalPath, lines, instruction), {
    command: 'mutation',
    logicalPath: target.logicalPath
  });
  void vscode.window.showInformationMessage(
    'Review Maestro\'s complete preview in the terminal, then allow once or deny there.'
  );
}

function requireWholeLineSelection(editor) {
  const selection = editor.selection;
  if (selection.isEmpty) {
    throw new PreviewError('selection_empty');
  }
  const endsAtNextLineStart = selection.end.character === 0 && selection.end.line > selection.start.line;
  const endsAtLineEnd = selection.end.character === editor.document.lineAt(selection.end.line).text.length;
  if (selection.start.character !== 0 || (!endsAtNextLineStart && !endsAtLineEnd)) {
    throw new PreviewError('selection_partial');
  }
}

async function doctor() {
  requireSupportedHost();
  const folder = requireWorkspaceFolder();
  const runtime = resolveRuntime(folder, { configRequired: true });
  showResolution(runtime);
  await launch(folder, runtime, commands.buildDoctorInvocation(runtime), { command: 'doctor' });
}

async function version() {
  requireSupportedHost();
  const folder = requireWorkspaceFolder();
  const runtime = resolveRuntime(folder);
  showResolution(runtime);
  await launch(folder, runtime, commands.buildVersionInvocation(runtime), { command: 'version' });
}

function requireWorkspaceFolder() {
  const folders = vscode.workspace.workspaceFolders;
  if (!folders || folders.length === 0) {
    throw new PreviewError('workspace_not_open');
  }
  if (folders.length === 1) {
    requireLocalFolder(folders[0]);
    return folders[0];
  }
  const editor = vscode.window.activeTextEditor;
  const active = editor && editor.document.uri.scheme === 'file'
    ? vscode.workspace.getWorkspaceFolder(editor.document.uri)
    : undefined;
  if (!active) {
    throw new PreviewError('workspace_ambiguous');
  }
  requireLocalFolder(active);
  return active;
}

function requireLocalFolder(folder) {
  if (!folder || folder.uri.scheme !== 'file') {
    throw new PreviewError('workspace_virtual');
  }
}

function validateTerminalName(value) {
  const name = typeof value === 'string' ? value.trim() : '';
  if (name === '' || name.length > 80 || /[\u0000-\u001f\u007f]/.test(name)) {
    throw new PreviewError('terminal_name_invalid');
  }
  return name;
}

async function launch(folder, runtime, invocation, metadata) {
  const definition = {
    type: 'maestro-preview',
    command: metadata.command,
    binaryOrigin: runtime.binaryOrigin,
    logicalPath: metadata.logicalPath,
    startedAt: Date.now()
  };
  const execution = new vscode.ProcessExecution(invocation.executable, invocation.args, {
    cwd: folder.uri.fsPath
  });
  const task = new vscode.Task(
    definition,
    folder,
    runtime.terminalName,
    'Maestro Preview',
    execution,
    []
  );
  task.presentationOptions = {
    reveal: vscode.TaskRevealKind.Always,
    panel: vscode.TaskPanelKind.New,
    focus: true,
    echo: true,
    showReuseMessage: false,
    clear: false,
    close: false
  };
  writeEvent('command_started', {
    command: metadata.command,
    status: 'started',
    binary_origin: runtime.binaryOrigin,
    logical_path: metadata.logicalPath
  });
  try {
    await vscode.tasks.executeTask(task);
  } catch {
    throw new PreviewError('command_unavailable');
  }
}

function onTaskEnded(event) {
  const definition = event.execution.task.definition;
  if (!definition || definition.type !== 'maestro-preview') {
    return;
  }
  const exitCode = Number.isInteger(event.exitCode) ? event.exitCode : -1;
  const duration = Math.max(0, Date.now() - Number(definition.startedAt || Date.now()));
  const denied = definition.command === 'mutation' && exitCode === 3;
  const passed = exitCode === 0;
  writeEvent(denied ? 'command_denied' : 'command_finished', {
    command: definition.command,
    status: denied ? 'denied' : passed ? 'passed' : 'failed',
    duration_ms: duration,
    exit_code: exitCode,
    binary_origin: definition.binaryOrigin,
    logical_path: definition.logicalPath,
    error_code: passed || denied ? undefined : 'cli_exit_nonzero'
  });
  if (!passed && !denied) {
    void presentError(new PreviewError('cli_exit_nonzero'));
  }
}

function showResolution(runtime) {
  const origin = runtime.binaryOrigin === 'setting' ? 'Settings' : 'extension host PATH';
  const config = runtime.configPath
    ? `; config from Settings: ${runtime.configPath}`
    : '';
  vscode.window.setStatusBarMessage(
    `Maestro Preview — binary from ${origin}: ${runtime.binaryPath}${config}`,
    10000
  );
}

function writeEvent(eventCode, fields) {
  outputChannel.appendLine(formatEvent(eventCode, fields));
}

async function presentError(error) {
  writeEvent('command_rejected', { status: 'rejected', error_code: error.code });
  const selected = await vscode.window.showErrorMessage(
    `${error.title} [${error.code}] ${error.cause} Next: ${error.action}.`,
    error.action
  );
  if (selected === SETTINGS_ACTION) {
    await vscode.commands.executeCommand('workbench.action.openSettings', '@ext:maestro-local.maestro-vscode-preview');
  } else if (selected === GUIDE_ACTION) {
    await vscode.commands.executeCommand('markdown.showPreview', vscode.Uri.joinPath(extensionContext.extensionUri, 'README.md'));
  } else if (selected === 'Save File') {
    await vscode.commands.executeCommand('workbench.action.files.save');
  } else if (selected === 'Manage Workspace Trust') {
    await vscode.commands.executeCommand('workbench.trust.manage');
  }
}

function guarded(handler) {
  return async (...args) => {
    try {
      await handler(...args);
    } catch (error) {
      void presentError(normalizeError(error));
    }
  };
}

function activate(context) {
  extensionContext = context;
  outputChannel = vscode.window.createOutputChannel('Maestro');
  context.subscriptions.push(outputChannel, vscode.tasks.onDidEndTaskProcess(onTaskEnded));
  const registrations = [
    ['maestro.askActiveFile', askActiveFile],
    ['maestro.replaceSelection', replaceSelection],
    ['maestro.doctor', doctor],
    ['maestro.version', version]
  ];
  for (const [name, handler] of registrations) {
    context.subscriptions.push(vscode.commands.registerCommand(name, guarded(handler)));
  }
}

function deactivate() {}

module.exports = { activate, deactivate };
