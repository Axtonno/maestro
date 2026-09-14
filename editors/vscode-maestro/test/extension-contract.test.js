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
const support = fs.readFileSync(path.join(root, 'SUPPORT.md'), 'utf8');
const commandIDs = [
  'maestro.askActiveFile',
  'maestro.replaceSelection',
  'maestro.doctor',
  'maestro.version'
];
const packagedFiles = [
  'CHANGELOG.md',
  'ASSET_PROVENANCE.md',
  'LICENSE',
  'README.md',
  'SECURITY.md',
  'SUPPORT.md',
  'command-builder.js',
  'diagnostics.js',
  'errors.js',
  'extension.js',
  'media/icon.png',
  'package.json',
  'path-resolver.js'
];

test('manifest freezes the Marketplace candidate identity and four-command workspace surface', () => {
  assert.equal(manifest.name, 'maestro-local-ai');
  assert.equal(manifest.displayName, 'Maestro for VS Code');
  assert.equal(manifest.version, '0.1.1');
  assert.equal(manifest.publisher, 'axtonno');
  assert.equal(manifest.private, undefined);
  assert.equal(manifest.preview, true);
  assert.equal(manifest.license, 'SEE LICENSE IN LICENSE');
  assert.equal(manifest.icon, 'media/icon.png');
  assert.equal(manifest.pricing, 'Free');
  assert.equal(manifest.repository.url, 'https://github.com/Axtonno/maestro.git');
  assert.equal(manifest.bugs.url, 'https://github.com/Axtonno/maestro/issues');
  assert.match(manifest.homepage, /^https:\/\/github\.com\/Axtonno\/maestro\//);
  assert.ok(manifest.categories.includes('Machine Learning'));
  assert.ok(manifest.keywords.length > 0 && manifest.keywords.length <= 30);
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
  assert.equal(manifest.devDependencies.yauzl, '3.4.0');
  assert.equal(manifest.devDependencies.yazl, '2.5.1');
  assert.equal(manifest.scripts['package:vsix'], 'node test/package-vsix.js');
  const packager = fs.readFileSync(path.join(root, 'test', 'package-vsix.js'), 'utf8');
  assert.match(packager, /'--pre-release'/);
  assert.match(packager, /1980-01-01T00:00:00\.000Z/);
  const icon = fs.readFileSync(path.join(root, manifest.icon));
  assert.deepEqual([...icon.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
  assert.equal(icon.readUInt32BE(16), 256);
  assert.equal(icon.readUInt32BE(20), 256);
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

test('public docs preserve the claim boundary and define support and updates', () => {
  assert.equal((source.match(/createOutputChannel\('Maestro'\)/g) || []).length, 1);
  assert.doesNotMatch(source, /appendLine\([^)]*(question|instruction|configPath|binaryPath)/);
  assert.match(readme, /not yet available in the Visual Studio\s+Marketplace/);
  assert.match(readme, /never\s+records prompts/);
  assert.match(readme, /allow once/);
  assert.match(readme, /releases\/download\/v0\.5\.0\/maestro-v0\.5\.0-linux-amd64\.tar\.gz/);
  assert.match(readme, /tar -xz --strip-components=1/);
  assert.match(readme, /uninstall-extension axtonno\.maestro-local-ai/);
  assert.match(support, /`0\.1\.x` is the first pre-release line/);
  assert.match(support, /higher patch containing the revert/);
});
