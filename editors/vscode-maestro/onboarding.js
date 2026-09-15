'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { resolveBinaryPath, resolveConfigPath } = require('./path-resolver');

const ONBOARDING_STATES = Object.freeze({
  READY: 'ready',
  CONFIG_MISSING: 'config_missing',
  BINARY_MISSING: 'binary_missing'
});

function inspectOnboarding(options) {
  const settings = options.settings || {};
  let binary;
  try {
    binary = resolveBinaryPath({
      configuredPath: settings.binaryPath || '',
      workspaceRoot: options.workspaceRoot,
      hostPath: options.hostPath,
      hostCwd: options.hostCwd
    });
  } catch {
    return Object.freeze({
      state: ONBOARDING_STATES.BINARY_MISSING,
      binaryReady: false,
      configReady: false,
      nextAction: 'settings'
    });
  }

  if (!configurationReady(settings.configPath || '', options)) {
    return Object.freeze({
      state: ONBOARDING_STATES.CONFIG_MISSING,
      binaryReady: true,
      configReady: false,
      nextAction: 'guide',
      binaryOrigin: binary.origin
    });
  }

  return Object.freeze({
    state: ONBOARDING_STATES.READY,
    binaryReady: true,
    configReady: true,
    nextAction: 'doctor',
    binaryOrigin: binary.origin
  });
}

function configurationReady(configuredPath, options) {
  if (configuredPath.trim() !== '') {
    try {
      resolveConfigPath(configuredPath, options.workspaceRoot);
      return true;
    } catch {
      return false;
    }
  }
  const candidate = defaultConfigPath(options.environment || process.env, options.workspaceRoot);
  return candidate !== undefined && isReadableRegular(candidate);
}

function defaultConfigPath(environment, hostCwd = process.cwd()) {
  const explicit = cleanEnvironmentPath(environment.MAESTRO_CONFIG);
  if (explicit !== '') {
    return path.isAbsolute(explicit) ? path.normalize(explicit) : path.resolve(hostCwd, explicit);
  }
  const xdg = cleanEnvironmentPath(environment.XDG_CONFIG_HOME);
  if (xdg !== '') {
    return absoluteFrom(hostCwd, path.join(xdg, 'maestro', 'config.yaml'));
  }
  const home = cleanEnvironmentPath(environment.HOME);
  return home === '' ? undefined : absoluteFrom(hostCwd, path.join(home, '.config', 'maestro', 'config.yaml'));
}

function absoluteFrom(base, candidate) {
  return path.isAbsolute(candidate) ? path.normalize(candidate) : path.resolve(base, candidate);
}

function cleanEnvironmentPath(value) {
  if (typeof value !== 'string') {
    return '';
  }
  const cleaned = value.trim();
  return /[\u0000\r\n]/.test(cleaned) ? '' : cleaned;
}

function isReadableRegular(candidate) {
  try {
    return fs.statSync(candidate).isFile() && (fs.accessSync(candidate, fs.constants.R_OK), true);
  } catch {
    return false;
  }
}

module.exports = { ONBOARDING_STATES, defaultConfigPath, inspectOnboarding };
