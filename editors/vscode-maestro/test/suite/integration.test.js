'use strict';

const assert = require('node:assert/strict');
const path = require('node:path');
const vscode = require('vscode');

const commandIDs = [
  'maestro.askActiveFile',
  'maestro.replaceSelection',
  'maestro.doctor',
  'maestro.version',
  'maestro.openSetupGuide',
  'maestro.openChat',
  'maestro.openTroubleshooting'
];

suite('Maestro for VS Code', () => {

  test('activates and registers the complete command surface', async () => {
    const extension = vscode.extensions.getExtension('axtonno.maestro-local-ai');
    assert.ok(extension, 'development extension was not discovered');
    await extension.activate();
    assert.equal(extension.isActive, true);
    const registered = await vscode.commands.getCommands(true);
    for (const command of commandIDs) {
      assert.ok(registered.includes(command), `${command} was not registered`);
    }
  });

  test('launches diagnostics as a direct process task', async () => {
    await configureHarmlessBinary();
    const task = await executeAndObserve('maestro.version');
    assertTask(task, ['version', '--diagnostic']);
  });

  test('opens the setup walkthrough and recovers missing binary and config', async () => {
    await vscode.commands.executeCommand('maestro.openSetupGuide');
    const configuration = vscode.workspace.getConfiguration('maestro');
    await configuration.update('binaryPath', '/definitely/missing/maestro', vscode.ConfigurationTarget.Workspace);
    await expectNoTask('maestro.version');
    await configuration.update('binaryPath', process.env.MAESTRO_VSCODE_TEST_BINARY, vscode.ConfigurationTarget.Workspace);
    await configuration.update('configPath', './missing.yaml', vscode.ConfigurationTarget.Workspace);
    assertTask(await executeAndObserve('maestro.version'), ['version', '--diagnostic']);
    await expectNoTask('maestro.doctor');
    await configuration.update('configPath', './maestro.yaml', vscode.ConfigurationTarget.Workspace);
    assertTask(await executeAndObserve('maestro.doctor'), [
      'doctor', '--mode', 'all', '--config', path.join(process.env.MAESTRO_VSCODE_TEST_WORKSPACE, 'maestro.yaml'), '--workspace-current'
    ]);
  });

  test('delegates the default setup configuration to the CLI', async () => {
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    const configuration = vscode.workspace.getConfiguration('maestro');
    await configuration.update('binaryPath', process.env.MAESTRO_VSCODE_TEST_BINARY, vscode.ConfigurationTarget.Workspace);
    await configuration.update('configPath', '', vscode.ConfigurationTarget.Workspace);
    const task = await executeAndObserve('maestro.doctor');
    assertTask(task, ['doctor', '--mode', 'all', '--workspace-current']);
  });

  test('binds chat and mutation to the saved active file and selection', async () => {
    await configureHarmlessBinary();
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    assert.ok(workspace);
    const document = await vscode.workspace.openTextDocument(path.join(workspace, 'app', 'Example.php'));
    const editor = await vscode.window.showTextDocument(document);

    const chat = await executeAndObserve('maestro.askActiveFile', 'Which status is returned?');
    assertTask(chat, [
      'chat', '--config', path.join(workspace, 'maestro.yaml'),
      '--workspace-current', '--file', 'app/Example.php', '--', 'Which status is returned?'
    ]);

    const line = document.lineAt(1);
    editor.selection = new vscode.Selection(1, 0, 1, line.text.length);
    const mutation = await executeAndObserve('maestro.replaceSelection', 'Change only 201 to 202.');
    assertTask(mutation, [
      'workspace', 'replace', '--file', 'app/Example.php', '--lines', '2:2',
      '--config', path.join(workspace, 'maestro.yaml'), '--workspace-current', '--', 'Change only 201 to 202.'
    ]);
    assert.equal(document.getText(), '<?php\nreturn 201;\n');
  });

  test('serves native chat with visible profile, models, workspace, and local response', async () => {
    await configureHarmlessBinary();
    const extension = vscode.extensions.getExtension('axtonno.maestro-local-ai');
    const api = await extension.activate();
    assert.ok(api && typeof api.handleNativeChat === 'function');
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    const document = await vscode.workspace.openTextDocument(path.join(workspace, 'app', 'Example.php'));
    await vscode.window.showTextDocument(document);
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
    assert.match(output, /Workspace \| workspace/);
    assert.match(output, /Native(?: |&nbsp;)chat(?: |&nbsp;)response\./);
  });

  test('rejects dirty buffers and non-single whole-line selections before launch', async () => {
    await configureHarmlessBinary();
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    const document = await vscode.workspace.openTextDocument(path.join(workspace, 'app', 'Example.php'));
    const editor = await vscode.window.showTextDocument(document);
    const started = [];
    const listener = vscode.tasks.onDidStartTask(event => {
      if (event.execution.task.definition.type === 'maestro-preview') {
        started.push(event.execution.task);
      }
    });

    await editor.edit(builder => builder.insert(new vscode.Position(1, 0), '// dirty\n'));
    await vscode.commands.executeCommand('maestro.askActiveFile', 'What is returned?');
    assert.equal(started.length, 0);
    await vscode.commands.executeCommand('workbench.action.files.revert');

    editor.selections = [
      new vscode.Selection(1, 0, 1, document.lineAt(1).text.length),
      new vscode.Selection(0, 0, 0, document.lineAt(0).text.length)
    ];
    await vscode.commands.executeCommand('maestro.replaceSelection', 'Change only 201 to 202.');
    assert.equal(started.length, 0);

    editor.selection = new vscode.Selection(1, 0, 1, 3);
    await vscode.commands.executeCommand('maestro.replaceSelection', 'Change only 201 to 202.');
    assert.equal(started.length, 0);
    listener.dispose();
  });
});

async function configureHarmlessBinary() {
  const configuration = vscode.workspace.getConfiguration('maestro');
  await configuration.update('binaryPath', process.env.MAESTRO_VSCODE_TEST_BINARY || '/bin/true', vscode.ConfigurationTarget.Workspace);
  await configuration.update('configPath', './maestro.yaml', vscode.ConfigurationTarget.Workspace);
}

function assertTask(task, expectedArgs) {
  assert.ok(task);
  assert.equal(task.definition.type, 'maestro-preview');
  assert.ok(task.execution instanceof vscode.ProcessExecution);
  assert.equal(task.execution.process, process.env.MAESTRO_VSCODE_TEST_BINARY || '/bin/true');
  assert.deepEqual(task.execution.args, expectedArgs);
  assert.equal(task.execution.options.cwd, process.env.MAESTRO_VSCODE_TEST_WORKSPACE);
}

async function executeAndObserve(command, ...args) {
  let listener;
  const started = new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      listener.dispose();
      reject(new Error('timed out waiting for Maestro task'));
    }, 5000);
    listener = vscode.tasks.onDidStartTask(event => {
      if (event.execution.task.definition.type === 'maestro-preview') {
        clearTimeout(timeout);
        listener.dispose();
        resolve(event.execution.task);
      }
    });
  });
  await vscode.commands.executeCommand(command, ...args);
  return started;
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
