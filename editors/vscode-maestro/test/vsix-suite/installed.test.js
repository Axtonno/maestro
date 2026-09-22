'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vscode = require('vscode');

suite('Installed Maestro VSIX', () => {
  test('activates the packaged extension and registers all commands', async () => {
    const extension = vscode.extensions.getExtension('axtonno.maestro-local-ai');
    assert.ok(extension, 'installed VSIX was not discovered');
    assert.doesNotMatch(extension.extensionPath, /vscode-maestro$/);
    await extension.activate();
    assert.equal(extension.isActive, true);
    const commands = await vscode.commands.getCommands(true);
    for (const command of [
      'maestro.askActiveFile',
      'maestro.replaceSelection',
      'maestro.doctor',
      'maestro.version',
      'maestro.openSetupGuide',
      'maestro.openChat',
      'maestro.openTroubleshooting'
    ]) {
      assert.ok(commands.includes(command), `${command} was not registered`);
    }
  });

  test('runs the installed VSIX in the qualified extension host', async () => {
    const extension = vscode.extensions.getExtension('axtonno.maestro-local-ai');
    const api = await extension.activate();
    const expected = vscode.env.remoteName === 'wsl' ? 'remote-wsl' : 'local-linux';
    assert.equal(api.executionEnvironment.target, expected);
    if (vscode.env.remoteName === 'wsl') {
      assert.equal(extension.extensionKind, vscode.ExtensionKind.Workspace);
    }
  });

  test('packages and opens the five-step setup walkthrough', async () => {
    const extension = vscode.extensions.getExtension('axtonno.maestro-local-ai');
    const walkthroughs = extension.packageJSON.contributes.walkthroughs;
    assert.equal(walkthroughs.length, 1);
    assert.equal(walkthroughs[0].id, 'maestro.setup');
    assert.equal(walkthroughs[0].steps.length, 5);
    await vscode.commands.executeCommand('maestro.openSetupGuide');
  });

  test('recovers missing binary and configuration with one next action', async () => {
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    const configuration = vscode.workspace.getConfiguration('maestro');
    await configuration.update('binaryPath', '/definitely/missing/maestro', vscode.ConfigurationTarget.Workspace);
    await expectNoTask('maestro.version');
    await configuration.update('binaryPath', process.env.MAESTRO_VSCODE_TEST_BINARY, vscode.ConfigurationTarget.Workspace);
    await configuration.update('configPath', './missing.yaml', vscode.ConfigurationTarget.Workspace);
    const versionStarted = observeTask();
    await vscode.commands.executeCommand('maestro.version');
    assert.deepEqual((await versionStarted).execution.args, ['version', '--diagnostic']);
    await expectNoTask('maestro.doctor');
    await configuration.update('configPath', './maestro.yaml', vscode.ConfigurationTarget.Workspace);
    const doctorStarted = observeTask();
    await vscode.commands.executeCommand('maestro.doctor');
    assert.deepEqual((await doctorStarted).execution.args, [
      'doctor', '--mode', 'all', '--config', path.join(workspace, 'maestro.yaml'), '--workspace-current'
    ]);
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

  test('uses the CLI default configuration after maestro setup', async () => {
    const binary = process.env.MAESTRO_VSCODE_TEST_BINARY;
    const configuration = vscode.workspace.getConfiguration('maestro');
    await configuration.update('binaryPath', binary, vscode.ConfigurationTarget.Workspace);
    await configuration.update('configPath', '', vscode.ConfigurationTarget.Workspace);
    const started = observeTask();
    const ended = observeTaskEnd('doctor');
    await vscode.commands.executeCommand('maestro.doctor');
    const task = await started;
    assert.ok(task.execution instanceof vscode.ProcessExecution);
    assert.equal(task.execution.process, binary);
    assert.deepEqual(task.execution.args, ['doctor', '--mode', 'all', '--workspace-current']);
    const event = await ended;
    assert.equal(event.exitCode, 0);
  });

  test('runs a focused chat from the clean installed profile', async () => {
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    const binary = process.env.MAESTRO_VSCODE_TEST_BINARY;
    const configuration = vscode.workspace.getConfiguration('maestro');
    await configuration.update('binaryPath', binary, vscode.ConfigurationTarget.Workspace);
    await configuration.update('configPath', './maestro.yaml', vscode.ConfigurationTarget.Workspace);
    const document = await vscode.workspace.openTextDocument(path.join(workspace, 'app', 'Example.php'));
    await vscode.window.showTextDocument(document);
    const started = observeTask();
    const ended = observeTaskEnd('chat');
    await vscode.commands.executeCommand('maestro.askActiveFile', 'Which status is returned?');
    const task = await started;
    assert.deepEqual(task.execution.args, [
      'chat', '--config', path.join(workspace, 'maestro.yaml'),
      '--workspace-current', '--file', 'app/Example.php', '--', 'Which status is returned?'
    ]);
    const event = await ended;
    assert.equal(event.exitCode, 0);
  });

  test('serves native chat from the installed VSIX with visible runtime identity', async () => {
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    const configuration = vscode.workspace.getConfiguration('maestro');
    await configuration.update('binaryPath', process.env.MAESTRO_VSCODE_TEST_BINARY, vscode.ConfigurationTarget.Workspace);
    await configuration.update('configPath', './maestro.yaml', vscode.ConfigurationTarget.Workspace);
    const document = await vscode.workspace.openTextDocument(path.join(workspace, 'app', 'Example.php'));
    await vscode.window.showTextDocument(document);
    const extension = vscode.extensions.getExtension('axtonno.maestro-local-ai');
    const api = await extension.activate();
    assert.ok(api && typeof api.handleNativeChat === 'function');
    const rendered = [];
    const stream = {
      markdown(value) { rendered.push(typeof value === 'string' ? value : value.value); },
      progress() {},
      button() {}
    };
    const token = { onCancellationRequested() { return { dispose() {} }; } };
    const result = await api.handleNativeChat({ prompt: 'What is returned?', command: undefined }, stream, token);
    assert.equal(result.metadata.status, 'passed');
    const output = rendered.join('\n');
    assert.match(output, /Profile \| recommended/);
    assert.match(output, /Chat model \| qwen3\.5:9b/);
    assert.match(output, /Mutation model \| qwen2\.5-coder:14b/);
    assert.match(output, /Extension host \| (?:local-linux|remote-wsl)/);
    assert.match(output, /Installed(?: |&nbsp;)native(?: |&nbsp;)chat(?: |&nbsp;)response\./);
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

async function expectNoTask(command) {
  let started = false;
  const listener = vscode.tasks.onDidStartTask(event => {
    if (event.execution.task.definition.type === 'maestro-preview') {
      started = true;
    }
  });
  await vscode.commands.executeCommand(command);
  await new Promise(resolve => setTimeout(resolve, 250));
  listener.dispose();
  assert.equal(started, false, `${command} unexpectedly launched a task`);
}
