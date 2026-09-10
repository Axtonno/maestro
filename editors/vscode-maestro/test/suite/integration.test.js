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

suite('Maestro VS Code prototype', () => {
  teardown(() => {
    for (const terminal of vscode.window.terminals) {
      terminal.dispose();
    }
  });

  test('activates and registers the complete command surface', async () => {
    const extension = vscode.extensions.getExtension('maestro-local.maestro-vscode-prototype');
    assert.ok(extension, 'development extension was not discovered');
    await extension.activate();
    assert.equal(extension.isActive, true);
    const registered = await vscode.commands.getCommands(true);
    for (const command of commandIDs) {
      assert.ok(registered.includes(command), `${command} was not registered`);
    }
  });

  test('launches diagnostic commands in a contained POSIX terminal', async () => {
    await configureHarmlessBinary();
    const before = vscode.window.terminals.length;
    await vscode.commands.executeCommand('maestro.version');
    await waitFor(() => vscode.window.terminals.length === before + 1);
    assertTerminal(vscode.window.terminals.at(-1));
  });

  test('binds chat and mutation to the saved active file and selection', async () => {
    await configureHarmlessBinary();
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    assert.ok(workspace);
    const document = await vscode.workspace.openTextDocument(path.join(workspace, 'app', 'Example.php'));
    const editor = await vscode.window.showTextDocument(document);

    let before = vscode.window.terminals.length;
    await vscode.commands.executeCommand('maestro.askActiveFile', 'Which status is returned?');
    await waitFor(() => vscode.window.terminals.length === before + 1);
    assertTerminal(vscode.window.terminals.at(-1));

    const line = document.lineAt(1);
    editor.selection = new vscode.Selection(1, 0, 1, line.text.length);
    before = vscode.window.terminals.length;
    await vscode.commands.executeCommand('maestro.replaceSelection', 'Change only 201 to 202.');
    await waitFor(() => vscode.window.terminals.length === before + 1);
    assertTerminal(vscode.window.terminals.at(-1));
    assert.equal(document.getText(), '<?php\nreturn 201;\n');
  });

  test('rejects dirty buffers and non-single whole-line selections before launch', async () => {
    await configureHarmlessBinary();
    const workspace = process.env.MAESTRO_VSCODE_TEST_WORKSPACE;
    const document = await vscode.workspace.openTextDocument(path.join(workspace, 'app', 'Example.php'));
    const editor = await vscode.window.showTextDocument(document);
    const terminalCount = vscode.window.terminals.length;

    await editor.edit(builder => builder.insert(new vscode.Position(1, 0), '// dirty\n'));
    await vscode.commands.executeCommand('maestro.askActiveFile', 'What is returned?');
    assert.equal(vscode.window.terminals.length, terminalCount);
    await vscode.commands.executeCommand('workbench.action.files.revert');

    editor.selections = [
      new vscode.Selection(1, 0, 1, document.lineAt(1).text.length),
      new vscode.Selection(0, 0, 0, document.lineAt(0).text.length)
    ];
    await vscode.commands.executeCommand('maestro.replaceSelection', 'Change only 201 to 202.');
    assert.equal(vscode.window.terminals.length, terminalCount);

    editor.selection = new vscode.Selection(1, 0, 1, 3);
    await vscode.commands.executeCommand('maestro.replaceSelection', 'Change only 201 to 202.');
    assert.equal(vscode.window.terminals.length, terminalCount);
  });
});

async function configureHarmlessBinary() {
  const configuration = vscode.workspace.getConfiguration('maestro');
  await configuration.update('binaryPath', '/bin/true', vscode.ConfigurationTarget.Workspace);
  await configuration.update('configPath', '', vscode.ConfigurationTarget.Workspace);
}

function assertTerminal(terminal) {
  assert.ok(terminal);
  assert.equal(terminal.creationOptions.shellPath, '/bin/sh');
  assert.equal(terminal.creationOptions.name, 'Maestro');
  assert.equal(terminal.creationOptions.cwd.fsPath, process.env.MAESTRO_VSCODE_TEST_WORKSPACE);
}

async function waitFor(predicate) {
  const deadline = Date.now() + 5000;
  while (!predicate()) {
    if (Date.now() >= deadline) {
      throw new Error('timed out waiting for VS Code terminal creation');
    }
    await new Promise(resolve => setTimeout(resolve, 25));
  }
}
