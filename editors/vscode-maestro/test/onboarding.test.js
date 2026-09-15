'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const test = require('node:test');
const { ONBOARDING_STATES, defaultConfigPath, inspectOnboarding } = require('../onboarding');

test('maps passive checks to one onboarding state and next action', t => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'maestro-onboarding-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const binary = path.join(root, 'maestro');
  const config = path.join(root, 'maestro.yaml');
  const hostCwd = path.join(root, 'cwd');
  fs.mkdirSync(hostCwd);
  fs.writeFileSync(binary, '#!/bin/sh\nexit 0\n', { mode: 0o755 });
  fs.writeFileSync(config, 'version: 4\n', 'utf8');

  const base = {
    workspaceRoot: root,
    hostPath: '',
    hostCwd,
    environment: {}
  };
  assert.deepEqual(inspectOnboarding({ ...base, settings: {} }), {
    state: ONBOARDING_STATES.BINARY_MISSING,
    binaryReady: false,
    configReady: false,
    nextAction: 'settings'
  });
  assert.deepEqual(inspectOnboarding({
    ...base,
    settings: { binaryPath: binary, configPath: './missing.yaml' }
  }), {
    state: ONBOARDING_STATES.CONFIG_MISSING,
    binaryReady: true,
    configReady: false,
    nextAction: 'guide',
    binaryOrigin: 'setting'
  });
  assert.deepEqual(inspectOnboarding({
    ...base,
    settings: { binaryPath: binary, configPath: './maestro.yaml' }
  }), {
    state: ONBOARDING_STATES.READY,
    binaryReady: true,
    configReady: true,
    nextAction: 'doctor',
    binaryOrigin: 'setting'
  });
});

test('mirrors CLI precedence for the passive default config check', () => {
  assert.equal(
    defaultConfigPath({ MAESTRO_CONFIG: './custom.yaml', XDG_CONFIG_HOME: '/xdg', HOME: '/home/user' }, '/work'),
    path.resolve('/work', 'custom.yaml')
  );
  assert.equal(
    defaultConfigPath({ XDG_CONFIG_HOME: '/xdg', HOME: '/home/user' }),
    path.join('/xdg', 'maestro', 'config.yaml')
  );
  assert.equal(
    defaultConfigPath({ XDG_CONFIG_HOME: '.config-root' }, '/work'),
    path.resolve('/work', '.config-root', 'maestro', 'config.yaml')
  );
  assert.equal(
    defaultConfigPath({ HOME: '/home/user' }),
    path.join('/home/user', '.config', 'maestro', 'config.yaml')
  );
  assert.equal(defaultConfigPath({}), undefined);
});
