'use strict';

const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { runTests } = require('@vscode/test-electron');

async function main() {
  const vscodeExecutablePath = process.env.MAESTRO_VSCODE_EXECUTABLE;
  if (!vscodeExecutablePath) {
    throw new Error('MAESTRO_VSCODE_EXECUTABLE must point to a VS Code executable');
  }

  const extensionDevelopmentPath = path.resolve(__dirname, '..');
  const extensionTestsPath = path.resolve(__dirname, 'suite', 'index.js');
  const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'maestro-vscode-integration-'));
  const workspace = path.join(temporary, 'workspace');
  const userData = path.join(temporary, 'user-data');
  const extensions = path.join(temporary, 'extensions');
  fs.mkdirSync(path.join(workspace, 'app'), { recursive: true });
  fs.writeFileSync(path.join(workspace, 'app', 'Example.php'), "<?php\nreturn 201;\n", 'utf8');

  // Codex may itself run inside an extension host. Those variables would make
  // the downloaded Electron binary start in Node/CLI mode instead of opening
  // the isolated test workbench.
  delete process.env.ELECTRON_RUN_AS_NODE;
  delete process.env.ELECTRON_NO_ATTACH_CONSOLE;
  for (const name of Object.keys(process.env)) {
    if (name.startsWith('VSCODE_')) {
      delete process.env[name];
    }
  }

  try {
    await runTests({
      vscodeExecutablePath,
      extensionDevelopmentPath,
      extensionTestsPath,
      launchArgs: [
        workspace,
        '--disable-extensions',
        `--user-data-dir=${userData}`,
        `--extensions-dir=${extensions}`
      ],
      extensionTestsEnv: {
        MAESTRO_VSCODE_TEST_WORKSPACE: workspace
      }
    });
  } finally {
    fs.rmSync(temporary, { recursive: true, force: true });
  }
}

main().catch(error => {
  console.error(error);
  process.exitCode = 1;
});
