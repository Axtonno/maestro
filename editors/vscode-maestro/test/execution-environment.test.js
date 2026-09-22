'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const {
  EXECUTION_TARGETS,
  classifyExecutionEnvironment,
  requireExecutionEnvironment
} = require('../execution-environment');
const { PreviewError } = require('../errors');

test('supports only the qualified local Linux and Remote WSL execution targets', () => {
  assert.deepEqual(classifyExecutionEnvironment({ platform: 'linux' }), {
    supported: true,
    target: EXECUTION_TARGETS.LOCAL_LINUX,
    remote: false
  });
  assert.deepEqual(classifyExecutionEnvironment({
    platform: 'linux', remoteName: 'wsl', workspaceExtension: true
  }), {
    supported: true,
    target: EXECUTION_TARGETS.REMOTE_WSL,
    remote: true
  });
});

test('fails closed for unqualified containers, SSH, unknown remotes, and non-Linux local hosts', () => {
  assert.deepEqual(classifyExecutionEnvironment({
    platform: 'linux', remoteName: 'dev-container', workspaceExtension: true
  }), { supported: false, code: 'dev_container_unqualified' });
  for (const remoteName of ['ssh-remote', 'codespaces', 'tunnel', 'unexpected']) {
    assert.deepEqual(classifyExecutionEnvironment({
      platform: 'linux', remoteName, workspaceExtension: true
    }), { supported: false, code: 'remote_unsupported' });
  }
  assert.deepEqual(classifyExecutionEnvironment({ platform: 'win32' }), {
    supported: false,
    code: 'platform_unsupported'
  });
  assert.deepEqual(classifyExecutionEnvironment({ platform: 'darwin' }), {
    supported: false,
    code: 'platform_unsupported'
  });
});

test('rejects a supported remote when the extension is running on the UI host', () => {
  assert.deepEqual(classifyExecutionEnvironment({
    platform: 'linux', remoteName: 'wsl', workspaceExtension: false
  }), { supported: false, code: 'extension_host_mismatch' });
  assert.throws(() => requireExecutionEnvironment({
    platform: 'linux', remoteName: 'wsl', workspaceExtension: false
  }, PreviewError), error => error.code === 'extension_host_mismatch');
});
