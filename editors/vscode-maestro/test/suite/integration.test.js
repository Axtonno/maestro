'use strict';

const assert = require('node:assert/strict');
const path = require('node:path');
const vscode = require('vscode');

const commandIDs = [
  'maestro.askActiveFile',
  'maestro.replaceSelection',
  'maestro.doctor',
  'maestro.version'
];

suite('Maestro VS Code Preview', () => {

  test('activates and registers the complete command surface', async () => {
    const extension = vscode.extensions.getExtension('maestro-local.maestro-vscode-preview');
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

  test('binds chat and mutation to the saved active file and selection', async () => {
    await configureHarmlessBinary();
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    assert.ok(workspace);
    const document = await vscode.workspace.openTextDocument(path.join(workspace, 'app', 'Example.php'));
    const editor = await vscode.window.showTextDocument(document);

    const chat = await executeAndObserve('maestro.askActiveFile', 'Which status is returned?');
    assertTask(chat, [
      'chat', '--config', path.join(workspace, 'maestro.yaml'),
      '--file', 'app/Example.php', '--', 'Which status is returned?'
    ]);

    const line = document.lineAt(1);
    editor.selection = new vscode.Selection(1, 0, 1, line.text.length);
    const mutation = await executeAndObserve('maestro.replaceSelection', 'Change only 201 to 202.');
    assertTask(mutation, [
      'workspace', 'replace', '--file', 'app/Example.php', '--lines', '2:2',
      '--config', path.join(workspace, 'maestro.yaml'), '--', 'Change only 201 to 202.'
    ]);
    assert.equal(document.getText(), '<?php\nreturn 201;\n');
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
  await configuration.update('binaryPath', '/bin/true', vscode.ConfigurationTarget.Workspace);
  await configuration.update('configPath', './maestro.yaml', vscode.ConfigurationTarget.Workspace);
}

function assertTask(task, expectedArgs) {
  assert.ok(task);
  assert.equal(task.definition.type, 'maestro-preview');
  assert.ok(task.execution instanceof vscode.ProcessExecution);
  assert.equal(task.execution.process, '/bin/true');
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
