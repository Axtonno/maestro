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
  fs.writeFileSync(path.join(workspace, 'maestro.yaml'), 'version: 4\n', 'utf8');
  const binary = path.join(workspace, 'maestro-cli-probe');
  fs.writeFileSync(binary, [
    '#!/bin/sh',
    'case "$1" in',
    '  profile) printf \'%s\\n\' \'{"schema_version":1,"profile":"recommended","provider":"ollama","chat_model":"qwen3.5:9b","mutation_model":"qwen2.5-coder:14b"}\' ;;',
    '  chat) case " $* " in *" -- "*) : ;; *) cat >/dev/null ;; esac; printf \'mode\\tchat\\nterminal\\tcompleted\\nmodel\\tqwen3.5:9b\\nfinish_reason\\tstop\\nresult\\nNative chat response.\\n\' ;;',
    '  doctor) printf \'pass\\tmutation_provider\\tollama_available\\npass\\tmutation_direct_chat_model\\tqualified_model_digest\\npass\\tmutation_controlled_mutation_model\\tqualified_model_digest\\n\' ;;',
    '  *) exit 0 ;;',
    'esac',
    ''
  ].join('\n'), 'utf8');
  fs.chmodSync(binary, 0o755);

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
        MAESTRO_VSCODE_TEST_WORKSPACE: workspace,
        MAESTRO_VSCODE_TEST_BINARY: binary,
        MAESTRO_VSCODE_TEST_MODE: '1'
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
