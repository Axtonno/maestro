'use strict';

const EXECUTION_TARGETS = Object.freeze({
  LOCAL_LINUX: 'local-linux',
  REMOTE_WSL: 'remote-wsl'
});

const SUPPORTED_REMOTES = Object.freeze({
  wsl: EXECUTION_TARGETS.REMOTE_WSL
});

function classifyExecutionEnvironment(input = {}) {
  const platform = typeof input.platform === 'string' ? input.platform : '';
  const remoteName = cleanRemoteName(input.remoteName);
  const workspaceExtension = input.workspaceExtension === true;

  if (remoteName === undefined) {
    if (platform !== 'linux') {
      return Object.freeze({ supported: false, code: 'platform_unsupported' });
    }
    return Object.freeze({
      supported: true,
      target: EXECUTION_TARGETS.LOCAL_LINUX,
      remote: false
    });
  }

  const target = SUPPORTED_REMOTES[remoteName];
  if (remoteName === 'dev-container') {
    return Object.freeze({ supported: false, code: 'dev_container_unqualified' });
  }
  if (!target) {
    return Object.freeze({ supported: false, code: 'remote_unsupported' });
  }
  if (!workspaceExtension) {
    return Object.freeze({ supported: false, code: 'extension_host_mismatch' });
  }
  if (platform !== 'linux') {
    return Object.freeze({ supported: false, code: 'platform_unsupported' });
  }
  return Object.freeze({ supported: true, target, remote: true });
}

function requireExecutionEnvironment(input, ErrorType) {
  const result = classifyExecutionEnvironment(input);
  if (!result.supported) {
    throw new ErrorType(result.code);
  }
  return result;
}

function cleanRemoteName(value) {
  if (value === undefined || value === null || value === '') {
    return undefined;
  }
  return typeof value === 'string' ? value : '__invalid__';
}

module.exports = { EXECUTION_TARGETS, classifyExecutionEnvironment, requireExecutionEnvironment };
