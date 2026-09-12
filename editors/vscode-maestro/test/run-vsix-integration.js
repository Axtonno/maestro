'use strict';

const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { runTests } = require('@vscode/test-electron');

const extensionID = 'maestro-local.maestro-vscode-preview';

async function main() {
  const vscodeExecutablePath = process.env.MAESTRO_VSCODE_EXECUTABLE;
  if (!vscodeExecutablePath) {
    throw new Error('MAESTRO_VSCODE_EXECUTABLE must point to a VS Code executable');
  }
  const vscodeCLIPath = process.env.MAESTRO_VSCODE_CLI || path.join(path.dirname(vscodeExecutablePath), 'bin', 'code');
  if (!fs.statSync(vscodeCLIPath).isFile()) {
    throw new Error(`VS Code CLI does not exist: ${vscodeCLIPath}`);
  }
  const vsixPath = path.resolve(
    process.env.MAESTRO_VSIX_PATH || path.join(__dirname, '..', 'dist', 'maestro-vscode-preview-0.1.0.vsix')
  );
  if (!fs.statSync(vsixPath).isFile()) {
    throw new Error(`VSIX does not exist: ${vsixPath}`);
  }

  const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'maestro-vsix-integration-'));
  const workspace = path.join(temporary, 'workspace');
  const userData = path.join(temporary, 'user-data');
  const extensions = path.join(temporary, 'extensions');
  fs.mkdirSync(path.join(workspace, 'app'), { recursive: true });
  fs.writeFileSync(path.join(workspace, 'app', 'Example.php'), "<?php\nreturn 201;\n", 'utf8');
  fs.writeFileSync(path.join(workspace, 'maestro.yaml'), 'version: 4\n', 'utf8');
  const ttyProbe = path.join(workspace, 'maestro-tty-probe');
  fs.writeFileSync(ttyProbe, [
    '#!/bin/sh',
    'if [ ! -t 0 ]; then exit 42; fi',
    'if [ "$1" = "version" ]; then exit 0; fi',
    'if [ "$1" != "workspace" ] || [ "$2" != "replace" ] || [ "$3" != "--file" ] || [ "$4" != "app/Example.php" ] || [ "$5" != "--lines" ] || [ "$6" != "2:2" ]; then exit 43; fi',
    'printf "preview ready; deny or allow once: "',
    'IFS= read -r decision',
    'if [ "$decision" = "d" ]; then exit 3; fi',
    'if [ "$decision" = "o" ]; then printf "<?php\\nreturn 202;\\n" > app/Example.php; exit 0; fi',
    'exit 44',
    ''
  ].join('\n'), 'utf8');
  fs.chmodSync(ttyProbe, 0o755);

  const cleanEnvironment = { ...process.env };
  delete cleanEnvironment.ELECTRON_RUN_AS_NODE;
  delete cleanEnvironment.ELECTRON_NO_ATTACH_CONSOLE;
  for (const name of Object.keys(cleanEnvironment)) {
    if (name.startsWith('VSCODE_')) {
      delete cleanEnvironment[name];
    }
  }
  const cli = (...args) => execFileSync(vscodeCLIPath, [
    `--user-data-dir=${userData}`,
    `--extensions-dir=${extensions}`,
    ...args
  ], { env: cleanEnvironment, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'], timeout: 30000 });

  try {
    cli('--install-extension', vsixPath, '--force');
    assertListed(cli('--list-extensions', '--show-versions'), true);
    cli('--uninstall-extension', extensionID);
    assertListed(cli('--list-extensions', '--show-versions'), false);
    cli('--install-extension', vsixPath, '--force');
    assertListed(cli('--list-extensions', '--show-versions'), true);

    delete process.env.ELECTRON_RUN_AS_NODE;
    delete process.env.ELECTRON_NO_ATTACH_CONSOLE;
    for (const name of Object.keys(process.env)) {
      if (name.startsWith('VSCODE_')) {
        delete process.env[name];
      }
    }

    await runTests({
      vscodeExecutablePath,
      extensionDevelopmentPath: path.join(__dirname, 'vsix-harness'),
      extensionTestsPath: path.join(__dirname, 'vsix-suite', 'index.js'),
      launchArgs: [
        workspace,
        '--disable-workspace-trust',
        `--user-data-dir=${userData}`,
        `--extensions-dir=${extensions}`
      ],
      extensionTestsEnv: {
        MAESTRO_VSCODE_TEST_WORKSPACE: workspace,
        MAESTRO_VSCODE_TEST_BINARY: process.env.MAESTRO_VSCODE_TEST_BINARY || '/bin/true',
        MAESTRO_VSCODE_TTY_PROBE: ttyProbe
      }
    });
  } finally {
    fs.rmSync(temporary, { recursive: true, force: true });
  }
}

function assertListed(output, expected) {
  const listed = output.split(/\r?\n/).some(line => line.trim() === `${extensionID}@0.1.0`);
  if (listed !== expected) {
    throw new Error(`extension listing mismatch: expected listed=${expected}`);
  }
}

main().catch(error => {
  console.error(error);
  process.exitCode = 1;
});
