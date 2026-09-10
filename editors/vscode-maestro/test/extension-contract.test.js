'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');
const manifest = JSON.parse(fs.readFileSync(path.join(root, 'package.json'), 'utf8'));
const source = fs.readFileSync(path.join(root, 'extension.js'), 'utf8');
const commandBuilder = fs.readFileSync(path.join(root, 'command-builder.js'), 'utf8');
const commandIDs = [
  'maestro.askActiveFile',
  'maestro.replaceSelection',
  'maestro.doctor',
  'maestro.version'
];

test('manifest freezes the four-command workspace-only surface', () => {
  assert.deepEqual(manifest.extensionKind, ['workspace']);
  assert.equal(manifest.capabilities.untrustedWorkspaces.supported, false);
  assert.equal(manifest.capabilities.virtualWorkspaces.supported, false);
  assert.equal(manifest.dependencies, undefined);
  assert.deepEqual(
    manifest.contributes.commands.map(entry => entry.command).sort(),
    [...commandIDs].sort()
  );
  assert.deepEqual(
    manifest.activationEvents.map(entry => entry.replace('onCommand:', '')).sort(),
    [...commandIDs].sort()
  );
});

test('extension delegates mutation and never writes or approves', () => {
  for (const forbidden of [
    'workspace.applyEdit',
    'new vscode.WorkspaceEdit',
    'writeFile',
    "sendText('y'",
    'sendText("y"'
  ]) {
    assert.equal(source.includes(forbidden), false, `forbidden API or approval token: ${forbidden}`);
  }
  assert.match(commandBuilder, /'workspace',[\s\S]*'replace'/);
  assert.match(source, /complete preview in the terminal/);
});
