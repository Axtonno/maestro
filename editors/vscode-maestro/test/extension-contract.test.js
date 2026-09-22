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
  'maestro.version',
  'maestro.openSetupGuide',
  'maestro.openChat',
  'maestro.openTroubleshooting'
];
const packagedFiles = [
  'CHANGELOG.md',
  'ASSET_PROVENANCE.md',
  'LICENSE',
  'README.md',
  'SECURITY.md',
  'SUPPORT.md',
  'command-builder.js',
  'chat-protocol.js',
  'cli-runner.js',
  'diagnostics.js',
  'errors.js',
  'execution-environment.js',
  'extension.js',
  'media/icon.png',
  'media/walkthrough/binary-path.md',
  'media/walkthrough/chat.md',
  'media/walkthrough/doctor.md',
  'media/walkthrough/install.md',
  'media/walkthrough/mutation.md',
  'onboarding.js',
  'package.json',
  'path-resolver.js',
  'workspace-context.js'
];

test('manifest freezes the unpublished remote-execution candidate identity', () => {
  assert.equal(manifest.name, 'maestro-local-ai');
  assert.equal(manifest.displayName, 'Maestro for VS Code');
  assert.equal(manifest.version, '0.4.0');
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
  assert.ok(manifest.categories.includes('Chat'));
  assert.ok(manifest.keywords.length > 0 && manifest.keywords.length <= 30);
  assert.deepEqual(manifest.extensionKind, ['workspace']);
  assert.equal(manifest.capabilities.untrustedWorkspaces.supported, false);
  assert.equal(manifest.capabilities.virtualWorkspaces.supported, false);
  assert.equal(manifest.dependencies, undefined);
  assert.equal(manifest.engines.vscode, '^1.100.0');
  assert.equal(manifest.contributes.chatParticipants.length, 1);
  assert.equal(manifest.contributes.chatParticipants[0].id, 'maestro.chat');
  assert.equal(manifest.contributes.chatParticipants[0].name, 'maestro');
  assert.deepEqual(manifest.contributes.chatParticipants[0].commands.map(command => command.name), ['status', 'preview', 'doctor']);
  assert.deepEqual(
    manifest.contributes.commands.map(entry => entry.command).sort(),
    [...commandIDs].sort()
  );
  assert.ok(manifest.activationEvents.includes('onStartupFinished'));
  assert.deepEqual(manifest.activationEvents
    .filter(entry => entry.startsWith('onCommand:'))
    .map(entry => entry.replace('onCommand:', '')).sort(), [...commandIDs].sort());
  assert.deepEqual(Object.fromEntries(manifest.contributes.commands.map(command => [command.command, command.title])), {
    'maestro.askActiveFile': 'Maestro: Chat About Active File',
    'maestro.replaceSelection': 'Maestro: Mutate Selection',
    'maestro.doctor': 'Maestro: Doctor',
    'maestro.version': 'Maestro: Show Binary Identity',
    'maestro.openSetupGuide': 'Maestro: Open Setup Guide',
    'maestro.openChat': 'Maestro: Open Chat',
    'maestro.openTroubleshooting': 'Maestro: Open Troubleshooting'
  });
});

test('walkthrough and settings encode onboarding plus the qualified profile', () => {
  assert.equal(manifest.contributes.walkthroughs.length, 1);
  const walkthrough = manifest.contributes.walkthroughs[0];
  assert.equal(walkthrough.id, 'maestro.setup');
  assert.equal(walkthrough.when, "workspacePlatform == 'linux'");
  assert.deepEqual(walkthrough.steps.map(step => step.id), [
    'maestro.setup.locate',
    'maestro.setup.doctor',
    'maestro.setup.binaryPath',
    'maestro.setup.chat',
    'maestro.setup.mutation'
  ]);
  for (const step of walkthrough.steps) {
    assert.ok(step.media.markdown);
    assert.ok(step.completionEvents.length > 0);
    assert.ok(fs.statSync(path.join(root, step.media.markdown)).isFile());
  }
  const properties = manifest.contributes.configuration.properties;
  assert.equal(properties['maestro.binaryPath'].default, '');
  assert.equal(properties['maestro.binaryPath'].scope, 'machine-overridable');
  assert.match(properties['maestro.binaryPath'].markdownDescription, /Example:.*takes precedence/s);
  assert.match(properties['maestro.binaryPath'].markdownDescription, /UI host is never used remotely/);
  assert.equal(properties['maestro.configPath'].default, '');
  assert.equal(properties['maestro.configPath'].scope, 'resource');
  assert.match(properties['maestro.configPath'].markdownDescription, /Example:.*takes precedence/s);
  assert.equal(properties['maestro.profile'].default, 'recommended');
  assert.equal(properties['maestro.profile'].scope, 'resource');
  assert.deepEqual(properties['maestro.profile'].enum, ['recommended']);
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

test('extension delegates mutation, captures only read-only CLI calls, and never writes or approves', () => {
  for (const forbidden of [
    'workspace.applyEdit',
    'new vscode.WorkspaceEdit',
    'writeFile',
    'fs.write',
    "sendText('y'",
    'sendText("y"',
    'exec(',
    'fetch(',
    'https.request',
    'http.request'
  ]) {
    assert.equal(source.includes(forbidden), false, `forbidden runtime authority: ${forbidden}`);
  }
  assert.match(commandBuilder, /'workspace', 'replace'/);
  assert.match(source, /complete preview in the terminal/);
  assert.match(source, /new vscode\.ProcessExecution/);
  const runner = fs.readFileSync(path.join(root, 'cli-runner.js'), 'utf8');
  assert.match(runner, /childProcess\.spawn/);
  assert.match(runner, /shell: false/);
  assert.doesNotMatch(runner, /exec\(|execFile\(|shell: true/);
  assert.match(source, /buildCapturedChatInvocation/);
  assert.match(source, /buildProfileInvocation/);
  assert.doesNotMatch(source, /telemetry|createWebview|registerWebview|createTreeView/i);
  assert.match(source, /vscode\.env\.remoteName/);
  assert.match(source, /vscode\.ExtensionKind\.Workspace/);
  assert.doesNotMatch(source, /context\.secrets\.(store|delete)|workspaceState\.update|globalState\.update/);
});

test('one passive status item exposes the required states and next actions', () => {
  assert.equal((source.match(/createStatusBarItem\(/g) || []).length, 1);
  for (const label of ['Maestro: Ready', 'Maestro: Config missing', 'Maestro: Binary missing']) {
    assert.ok(source.includes(label), `missing status label: ${label}`);
  }
  assert.match(source, /statusItem\.command = 'maestro\.doctor'/);
  assert.match(source, /statusItem\.command = 'maestro\.openSetupGuide'/);
  assert.match(source, /command: 'workbench\.action\.openSettings'/);
  assert.doesNotMatch(source, /statusBarItem\.(backgroundColor|color)/);
});

test('public docs preserve the claim boundary and define support and updates', () => {
  assert.equal((source.match(/createOutputChannel\('Maestro'\)/g) || []).length, 1);
  assert.doesNotMatch(source, /appendLine\([^)]*(question|instruction|configPath|binaryPath)/);
  assert.match(readme, /not yet available in the Visual Studio\s+Marketplace/);
  assert.match(readme, /never\s+records prompts/);
  assert.match(readme, /allow once/);
  assert.match(readme, /post-v0\.5\.0 Maestro source candidate/);
  assert.match(readme, /go build -o/);
  assert.match(readme, /@maestro \/status/);
  assert.match(readme, /uninstall-extension axtonno\.maestro-local-ai/);
  assert.match(support, /`0\.4\.x` freezes the\s+remote execution contract/);
  assert.match(readme, /Dev Container \| Fails with `dev_container_unqualified`/);
  assert.match(readme, /Remote SSH or another remote \| Fails with `remote_unsupported`/);
  assert.match(support, /higher patch containing the revert/);
});
