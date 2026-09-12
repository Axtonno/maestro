'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vscode = require('vscode');

suite('Installed Maestro VSIX', () => {
  test('activates the packaged extension and registers all commands', async () => {
    const extension = vscode.extensions.getExtension('maestro-local.maestro-vscode-preview');
    assert.ok(extension, 'installed VSIX was not discovered');
    assert.doesNotMatch(extension.extensionPath, /vscode-maestro$/);
    await extension.activate();
    assert.equal(extension.isActive, true);
    const commands = await vscode.commands.getCommands(true);
    for (const command of [
      'maestro.askActiveFile',
      'maestro.replaceSelection',
      'maestro.doctor',
      'maestro.version'
    ]) {
      assert.ok(commands.includes(command), `${command} was not registered`);
    }
  });

  test('starts the installed identity command from explicit binary resolution', async () => {
    const binary = process.env.MAESTRO_VSCODE_TEST_BINARY;
    assert.ok(binary);
    const configuration = vscode.workspace.getConfiguration('maestro');
    await configuration.update('binaryPath', binary, vscode.ConfigurationTarget.Workspace);
    const started = observeTask();
    const ended = observeTaskEnd('version');
    await vscode.commands.executeCommand('maestro.version');
    const task = await started;
    assert.equal(task.definition.type, 'maestro-preview');
    assert.equal(task.definition.binaryOrigin, 'setting');
    assert.ok(task.execution instanceof vscode.ProcessExecution);
    assert.equal(task.execution.process, binary);
    assert.deepEqual(task.execution.args, ['version', '--diagnostic']);
    const event = await ended;
    assert.equal(event.exitCode, 0);
  });

  test('keeps mutation approval in a real TTY and deny has zero effects', async () => {
    const fixture = await mutationFixture();
    const terminalOpened = observeTerminal();
    const ended = observeTaskEnd('mutation');
    await vscode.commands.executeCommand('maestro.replaceSelection', 'Change only 201 to 202.');
    const terminal = await terminalOpened;
    terminal.sendText('d', true);
    const event = await ended;
    assert.equal(event.exitCode, 3);
    assert.equal(fs.readFileSync(fixture.file, 'utf8'), "<?php\nreturn 201;\n");
  });

  test('allow-once applies only the previewed single-line change', async () => {
    const fixture = await mutationFixture();
    const terminalOpened = observeTerminal();
    const ended = observeTaskEnd('mutation');
    await vscode.commands.executeCommand('maestro.replaceSelection', 'Change only 201 to 202.');
    const terminal = await terminalOpened;
    terminal.sendText('o', true);
    const event = await ended;
    assert.equal(event.exitCode, 0);
    assert.equal(fs.readFileSync(fixture.file, 'utf8'), "<?php\nreturn 202;\n");
  });
});

async function mutationFixture() {
  const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
  const binary = process.env.MAESTRO_VSCODE_TTY_PROBE;
  assert.ok(workspace);
  assert.ok(binary);
  const file = path.join(workspace, 'app', 'Example.php');
  fs.writeFileSync(file, "<?php\nreturn 201;\n", 'utf8');
  const configuration = vscode.workspace.getConfiguration('maestro');
  await configuration.update('binaryPath', binary, vscode.ConfigurationTarget.Workspace);
  await configuration.update('configPath', './maestro.yaml', vscode.ConfigurationTarget.Workspace);
  const document = await vscode.workspace.openTextDocument(file);
  const editor = await vscode.window.showTextDocument(document);
  const line = document.lineAt(1);
  editor.selection = new vscode.Selection(1, 0, 1, line.text.length);
  return { file };
}

function observeTask() {
  let listener;
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      listener.dispose();
      reject(new Error('timed out waiting for installed Maestro task'));
    }, 5000);
    listener = vscode.tasks.onDidStartTask(event => {
      if (event.execution.task.definition.type === 'maestro-preview') {
        clearTimeout(timeout);
        listener.dispose();
        resolve(event.execution.task);
      }
    });
  });
}

function observeTaskEnd(command) {
  let listener;
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      listener.dispose();
      reject(new Error(`timed out waiting for ${command} task completion`));
    }, 5000);
    listener = vscode.tasks.onDidEndTaskProcess(event => {
      const definition = event.execution.task.definition;
      if (definition.type === 'maestro-preview' && definition.command === command) {
        clearTimeout(timeout);
        listener.dispose();
        resolve(event);
      }
    });
  });
}

function observeTerminal() {
  let listener;
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      listener.dispose();
      reject(new Error('timed out waiting for Maestro task terminal'));
    }, 5000);
    listener = vscode.window.onDidOpenTerminal(terminal => {
      clearTimeout(timeout);
      listener.dispose();
      resolve(terminal);
    });
  });
}
