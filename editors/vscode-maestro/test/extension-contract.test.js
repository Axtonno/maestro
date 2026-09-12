'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');
const manifest = JSON.parse(fs.readFileSync(path.join(root, 'package.json'), 'utf8'));
const source = fs.readFileSync(path.join(root, 'extension.js'), 'utf8');
const commandBuilder = fs.readFileSync(path.join(root, 'command-builder.js'), 'utf8');
const ignore = fs.readFileSync(path.join(root, '.vscodeignore'), 'utf8');
const readme = fs.readFileSync(path.join(root, 'README.md'), 'utf8');
const commandIDs = [
  'maestro.askActiveFile',
  'maestro.replaceSelection',
  'maestro.doctor',
  'maestro.version'
];
const packagedFiles = [
  'CHANGELOG.md',
  'LICENSE',
  'README.md',
  'command-builder.js',
  'diagnostics.js',
  'errors.js',
  'extension.js',
  'package.json',
  'path-resolver.js'
];

test('manifest identifies the local Preview and freezes the four-command workspace surface', () => {
  assert.equal(manifest.name, 'maestro-vscode-preview');
  assert.equal(manifest.displayName, 'Maestro (Preview)');
  assert.equal(manifest.version, '0.1.0');
  assert.equal(manifest.publisher, 'maestro-local');
  assert.equal(manifest.private, true);
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

test('packaging is allowlisted and has no publication script', () => {
  assert.equal(ignore.split(/\r?\n/)[0], '**/*');
  for (const file of packagedFiles) {
    assert.match(ignore, new RegExp(`^!${file.replace('.', '\\.')}\\s*$`, 'm'));
    assert.ok(fs.statSync(path.join(root, file)).isFile(), `${file} must exist`);
  }
  assert.equal(manifest.scripts.publish, undefined);
  assert.equal(manifest.scripts.deploy, undefined);
  assert.equal(manifest.devDependencies['@vscode/vsce'], '3.9.2');
});

test('extension delegates mutation and never writes or approves', () => {
  for (const forbidden of [
    'workspace.applyEdit',
    'new vscode.WorkspaceEdit',
    'writeFile',
    'fs.write',
    "sendText('y'",
    'sendText("y"',
    'child_process',
    'exec(',
    'spawn(',
    'fetch(',
    'https.request',
    'http.request'
  ]) {
    assert.equal(source.includes(forbidden), false, `forbidden runtime authority: ${forbidden}`);
  }
  assert.match(commandBuilder, /'workspace', 'replace'/);
  assert.match(source, /complete preview in the terminal/);
  assert.match(source, /new vscode\.ProcessExecution/);
});

test('single Maestro channel uses redacted events and Preview docs preserve claim boundary', () => {
  assert.equal((source.match(/createOutputChannel\('Maestro'\)/g) || []).length, 1);
  assert.doesNotMatch(source, /appendLine\([^)]*(question|instruction|configPath|binaryPath)/);
  assert.match(readme, /not available in the\s+Visual Studio Marketplace/);
  assert.match(readme, /never contains prompts/);
  assert.match(readme, /allow once/);
  assert.match(readme, /uninstall-extension maestro-local\.maestro-vscode-preview/);
});
