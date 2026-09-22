'use strict';

const fs = require('node:fs');
const path = require('node:path');
const vscode = require('vscode');

const EXTENSION_ID = 'axtonno.maestro-local-ai';
const REPORT_PATH = '/tmp/maestro-m52-remote-probe.json';

async function activate() {
  const report = {
    schema_version: 1,
    remote_name: vscode.env.remoteName || '',
    workspace_trusted: vscode.workspace.isTrusted,
    checks: {}
  };
  try {
    check(report, 'remote_name', vscode.env.remoteName === 'wsl');
    const extension = vscode.extensions.getExtension(EXTENSION_ID);
    check(report, 'candidate_installed', Boolean(extension));
    if (!extension) {
      throw new Error('candidate extension is not installed in the remote host');
    }
    report.extension_path = extension.extensionPath;
    report.extension_kind = extension.extensionKind;
    check(report, 'workspace_extension', extension.extensionKind === vscode.ExtensionKind.Workspace);
    check(report, 'remote_extension_path', extension.extensionPath.startsWith('/home/') || extension.extensionPath.startsWith('/tmp/'));
    check(report, 'workspace_trusted', vscode.workspace.isTrusted);

    await extension.activate();
    const registered = await vscode.commands.getCommands(true);
    check(report, 'commands_registered', ['maestro.version', 'maestro.doctor', 'maestro.askActiveFile', 'maestro.replaceSelection']
      .every(command => registered.includes(command)));

    const folder = vscode.workspace.workspaceFolders && vscode.workspace.workspaceFolders[0];
    check(report, 'remote_workspace_open', Boolean(folder && folder.uri.scheme === 'file'));
    if (!folder) {
      throw new Error('remote fixture workspace is not open');
    }
    report.workspace_path = folder.uri.fsPath;
    const binary = path.join(folder.uri.fsPath, 'maestro-cli-probe');
    const config = vscode.workspace.getConfiguration('maestro', folder.uri);
    await config.update('binaryPath', binary, vscode.ConfigurationTarget.Workspace);
    await config.update('configPath', './maestro.yaml', vscode.ConfigurationTarget.Workspace);

    const task = await executeAndObserve('maestro.version');
    report.task = {
      process: task.execution.process,
      args: task.execution.args,
      cwd: task.execution.options.cwd,
      execution_target: task.definition.executionTarget
    };
    check(report, 'remote_binary_selected', task.execution.process === binary);
    check(report, 'remote_cwd_selected', task.execution.options.cwd === folder.uri.fsPath);
    check(report, 'remote_target_recorded', task.definition.executionTarget === 'remote-wsl');
    report.verdict = 'passed';
  } catch (error) {
    report.verdict = 'failed';
    report.error = error instanceof Error ? error.message : String(error);
  }
  fs.writeFileSync(REPORT_PATH, `${JSON.stringify(report, null, 2)}\n`, { encoding: 'utf8', mode: 0o600 });
}

function check(report, name, passed) {
  report.checks[name] = passed ? 'passed' : 'failed';
  if (!passed) {
    throw new Error(`remote check failed: ${name}`);
  }
}

async function executeAndObserve(command) {
  let listener;
  const started = new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      listener.dispose();
      reject(new Error(`timed out waiting for ${command}`));
    }, 10000);
    listener = vscode.tasks.onDidStartTask(event => {
      if (event.execution.task.definition.type === 'maestro-preview') {
        clearTimeout(timeout);
        listener.dispose();
        resolve(event.execution.task);
      }
    });
  });
  await vscode.commands.executeCommand(command);
  return started;
}

function deactivate() {}

module.exports = { activate, deactivate };
