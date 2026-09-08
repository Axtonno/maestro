'use strict';

const vscode = require('vscode');
const commands = require('./command-builder');

function settingsFor(scope) {
  const configuration = vscode.workspace.getConfiguration('maestro', scope);
  return {
    binaryPath: configuration.get('binaryPath', 'maestro'),
    configPath: configuration.get('configPath', ''),
    terminalName: configuration.get('terminalName', 'Maestro')
  };
}

function requireLinux() {
  if (!vscode.workspace.isTrusted) {
    throw new Error('trust the workspace before invoking a local Maestro binary');
  }
  if (process.platform !== 'linux') {
    throw new Error('the M40 prototype supports only Linux extension hosts, including Remote WSL');
  }
}

function activeWorkspaceEditor(options = {}) {
  const editor = vscode.window.activeTextEditor;
  if (!editor || editor.document.uri.scheme !== 'file') {
    throw new Error('open a local workspace file first');
  }
  const folder = vscode.workspace.getWorkspaceFolder(editor.document.uri);
  if (!folder) {
    throw new Error('the active file is not inside an open workspace folder');
  }
  if (editor.document.isDirty) {
    throw new Error('save the active file before invoking Maestro');
  }
  const logicalPath = options.mutation
    ? commands.mutationLogicalPath(folder.uri.fsPath, editor.document.uri.fsPath)
    : commands.safeLogicalPath(folder.uri.fsPath, editor.document.uri.fsPath);
  return { editor, folder, logicalPath };
}

function launch(folder, settings, command) {
  const name = settings.terminalName.trim() || 'Maestro';
  const terminal = vscode.window.createTerminal({ name, cwd: folder.uri, shellPath: '/bin/sh' });
  terminal.show(false);
  terminal.sendText(command, true);
}

async function askActiveFile() {
  requireLinux();
  const target = activeWorkspaceEditor();
  const question = await vscode.window.showInputBox({
    title: 'Maestro: Ask About Active File',
    prompt: `Question for ${target.logicalPath}`,
    ignoreFocusOut: true,
    validateInput: value => value.trim() === '' ? 'Enter a question.' : undefined
  });
  if (question === undefined) {
    return;
  }
  const settings = settingsFor(target.folder.uri);
  launch(target.folder, settings, commands.buildChatCommand(settings, target.logicalPath, question));
}

async function replaceSelection() {
  requireLinux();
  const target = activeWorkspaceEditor({ mutation: true });
  if (target.editor.selections.length !== 1) {
    throw new Error('Controlled Mutation accepts exactly one editor selection');
  }
  requireWholeLineSelection(target.editor);
  const lines = commands.inclusiveSelectedLines(target.editor.selection);
  const instruction = await vscode.window.showInputBox({
    title: 'Maestro: Replace Selected Lines',
    prompt: `Instruction for ${target.logicalPath}:${lines}`,
    placeHolder: 'Describe only the replacement for the selected lines.',
    ignoreFocusOut: true,
    validateInput: value => value.trim() === '' ? 'Enter a mutation instruction.' : undefined
  });
  if (instruction === undefined) {
    return;
  }
  const settings = settingsFor(target.folder.uri);
  const command = commands.buildMutationCommand(settings, target.logicalPath, lines, instruction);
  launch(target.folder, settings, command);
  vscode.window.showInformationMessage('Review Maestro\'s complete preview in the terminal, then allow once or deny there.');
}

function requireWholeLineSelection(editor) {
  const selection = editor.selection;
  const endsAtNextLineStart = selection.end.character === 0 && selection.end.line > selection.start.line;
  const endsAtLineEnd = selection.end.character === editor.document.lineAt(selection.end.line).text.length;
  if (selection.start.character !== 0 || (!endsAtNextLineStart && !endsAtLineEnd)) {
    throw new Error('select complete lines; Maestro replaces the inclusive line range');
  }
}

async function doctor() {
  requireLinux();
  const folder = requireWorkspaceFolder();
  const settings = settingsFor(folder.uri);
  launch(folder, settings, commands.buildDoctorCommand(settings));
}

async function version() {
  requireLinux();
  const folder = requireWorkspaceFolder();
  const settings = settingsFor(folder.uri);
  launch(folder, settings, commands.buildVersionCommand(settings));
}

function requireWorkspaceFolder() {
  const folders = vscode.workspace.workspaceFolders;
  if (!folders || folders.length === 0) {
    throw new Error('open a workspace folder first');
  }
  if (folders.length === 1) {
    return folders[0];
  }
  const editor = vscode.window.activeTextEditor;
  const active = editor && vscode.workspace.getWorkspaceFolder(editor.document.uri);
  if (!active) {
    throw new Error('focus a file in the workspace folder Maestro should use');
  }
  return active;
}

function guarded(handler) {
  return async () => {
    try {
      await handler();
    } catch (error) {
      const message = error instanceof Error ? error.message : 'unexpected prototype failure';
      vscode.window.showErrorMessage(`Maestro: ${message}`);
    }
  };
}

function activate(context) {
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
