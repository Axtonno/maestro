'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const test = require('node:test');
const { resolveBinaryPath, resolveConfigPath } = require('../path-resolver');

function fixture(t) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'maestro-resolver-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  fs.mkdirSync(path.join(root, 'bin-a'));
  fs.mkdirSync(path.join(root, 'bin-b'));
  fs.mkdirSync(path.join(root, 'config'));
  return root;
}

function executable(file) {
  fs.writeFileSync(file, '#!/bin/sh\nexit 0\n');
  fs.chmodSync(file, 0o755);
}

test('explicit binary setting wins and relative paths use the workspace', t => {
  const root = fixture(t);
  executable(path.join(root, 'configured-maestro'));
  executable(path.join(root, 'bin-a', 'maestro'));
  assert.deepEqual(resolveBinaryPath({
    configuredPath: './configured-maestro',
    workspaceRoot: root,
    hostPath: path.join(root, 'bin-a')
  }), {
    path: path.join(root, 'configured-maestro'),
    origin: 'setting'
  });
});

test('PATH search preserves directory order and requires an executable regular file', t => {
  const root = fixture(t);
  fs.writeFileSync(path.join(root, 'bin-a', 'maestro'), 'not executable');
  executable(path.join(root, 'bin-b', 'maestro'));
  assert.deepEqual(resolveBinaryPath({
    configuredPath: '',
    workspaceRoot: root,
    hostPath: [path.join(root, 'bin-a'), path.join(root, 'bin-b')].join(path.delimiter)
  }), {
    path: path.join(root, 'bin-b', 'maestro'),
    origin: 'path'
  });
});

test('missing and invalid binaries have stable errors', t => {
  const root = fixture(t);
  fs.writeFileSync(path.join(root, 'configured-maestro'), 'not executable');
  assert.throws(() => resolveBinaryPath({
    configuredPath: './configured-maestro',
    workspaceRoot: root,
    hostPath: ''
  }), error => error.code === 'binary_not_executable');
  assert.throws(() => resolveBinaryPath({
    configuredPath: './missing-maestro',
    workspaceRoot: root,
    hostPath: ''
  }), error => error.code === 'binary_not_found');
  assert.throws(() => resolveBinaryPath({
    configuredPath: '',
    workspaceRoot: root,
    hostPath: path.join(root, 'bin-a')
  }), error => error.code === 'binary_not_found');
});

test('config resolution requires a readable regular file contained by the workspace', t => {
  const root = fixture(t);
  const config = path.join(root, 'config', 'maestro.yaml');
  fs.writeFileSync(config, 'version: 4\n');
  assert.deepEqual(resolveConfigPath('./config/maestro.yaml', root), {
    path: config,
    logicalPath: 'config/maestro.yaml',
    origin: 'setting'
  });
  assert.throws(() => resolveConfigPath('', root), error => error.code === 'config_not_set');
  assert.throws(() => resolveConfigPath('./missing.yaml', root), error => error.code === 'config_not_found');

  const outside = path.join(path.dirname(root), `${path.basename(root)}-outside.yaml`);
  fs.writeFileSync(outside, 'version: 4\n');
  t.after(() => fs.rmSync(outside, { force: true }));
  assert.throws(() => resolveConfigPath(outside, root), error => error.code === 'config_outside_expected_scope');
});

test('config resolution handles UTF-8 names and CRLF without changing containment', t => {
  const root = fixture(t);
  const directory = path.join(root, 'config', 'caffè');
  fs.mkdirSync(directory);
  const config = path.join(directory, 'maestro.yaml');
  fs.writeFileSync(config, 'version: 4\r\n', 'utf8');
  const resolved = resolveConfigPath('./config/caffè/maestro.yaml', root);
  assert.equal(resolved.path, config);
  assert.equal(resolved.logicalPath, 'config/caffè/maestro.yaml');
  assert.equal(fs.readFileSync(resolved.path, 'utf8'), 'version: 4\r\n');
});

test('config symlinks cannot escape the workspace', { skip: process.platform === 'win32' }, t => {
  const root = fixture(t);
  const outside = path.join(path.dirname(root), `${path.basename(root)}-secret.yaml`);
  const link = path.join(root, 'config', 'linked.yaml');
  fs.writeFileSync(outside, 'secret: true\n');
  fs.symlinkSync(outside, link);
  t.after(() => fs.rmSync(outside, { force: true }));
  assert.throws(() => resolveConfigPath('./config/linked.yaml', root), error => {
    return error.code === 'config_outside_expected_scope';
  });
});
