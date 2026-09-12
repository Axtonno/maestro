'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { PreviewError } = require('./errors');

const FORBIDDEN_PATH_BYTES = /[\u0000\r\n]/;

function resolveBinaryPath(options) {
  const configured = cleanPathSetting(options.configuredPath || '', 'binary_not_found');
  const workspaceRoot = options.workspaceRoot;
  if (configured !== '') {
    const candidate = path.isAbsolute(configured)
      ? path.normalize(configured)
      : path.resolve(workspaceRoot, configured);
    assertExecutable(candidate);
    return Object.freeze({ path: candidate, origin: 'setting' });
  }

  const hostPath = typeof options.hostPath === 'string' ? options.hostPath : '';
  const hostCwd = options.hostCwd || process.cwd();
  for (const entry of hostPath.split(path.delimiter)) {
    const directory = entry === '' ? hostCwd : path.resolve(hostCwd, entry);
    const candidate = path.join(directory, 'maestro');
    if (isExecutable(candidate)) {
      return Object.freeze({ path: candidate, origin: 'path' });
    }
  }
  throw new PreviewError('binary_not_found');
}

function resolveConfigPath(configuredPath, workspaceRoot) {
  const configured = cleanPathSetting(configuredPath || '', 'config_not_set');
  if (configured === '') {
    throw new PreviewError('config_not_set');
  }
  const candidate = path.isAbsolute(configured)
    ? path.normalize(configured)
    : path.resolve(workspaceRoot, configured);
  assertRegular(candidate, 'config_not_found');

  let resolvedWorkspace;
  let resolvedCandidate;
  try {
    resolvedWorkspace = fs.realpathSync(workspaceRoot);
    resolvedCandidate = fs.realpathSync(candidate);
  } catch {
    throw new PreviewError('config_not_found');
  }
  const relative = path.relative(resolvedWorkspace, resolvedCandidate);
  if (relative === '' || path.isAbsolute(relative) || relative === '..' || relative.startsWith(`..${path.sep}`)) {
    throw new PreviewError('config_outside_expected_scope');
  }
  try {
    fs.accessSync(resolvedCandidate, fs.constants.R_OK);
  } catch {
    throw new PreviewError('config_not_readable');
  }
  return Object.freeze({
    path: resolvedCandidate,
    logicalPath: relative.split(path.sep).join('/'),
    origin: 'setting'
  });
}

function cleanPathSetting(value, emptyCode) {
  if (typeof value !== 'string') {
    throw new PreviewError(emptyCode);
  }
  const cleaned = value.trim();
  if (FORBIDDEN_PATH_BYTES.test(cleaned)) {
    throw new PreviewError(emptyCode);
  }
  return cleaned;
}

function assertExecutable(candidate) {
  let info;
  try {
    info = fs.statSync(candidate);
  } catch {
    throw new PreviewError('binary_not_found');
  }
  if (!info.isFile()) {
    throw new PreviewError('binary_not_executable');
  }
  try {
    fs.accessSync(candidate, fs.constants.X_OK);
  } catch {
    throw new PreviewError('binary_not_executable');
  }
}

function assertRegular(candidate, code) {
  try {
    if (!fs.statSync(candidate).isFile()) {
      throw new PreviewError(code);
    }
  } catch (error) {
    if (error instanceof PreviewError) {
      throw error;
    }
    throw new PreviewError(code);
  }
}

function isExecutable(candidate) {
  try {
    return fs.statSync(candidate).isFile() && (fs.accessSync(candidate, fs.constants.X_OK), true);
  } catch {
    return false;
  }
}

module.exports = { resolveBinaryPath, resolveConfigPath };
