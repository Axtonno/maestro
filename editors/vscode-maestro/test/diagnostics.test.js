'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const { formatEvent } = require('../diagnostics');

test('diagnostics contain only deterministic allowlisted fields', () => {
  const line = formatEvent('command_finished', {
    command: 'mutation',
    status: 'passed',
    duration_ms: 42,
    exit_code: 0,
    binary_origin: 'setting',
    logical_path: 'app/Order.php'
  }, new Date('2026-09-11T12:00:00.000Z'));
  assert.equal(
    line,
    '2026-09-11T12:00:00.000Z event=command_finished command=mutation status=passed duration_ms=42 exit_code=0 binary_origin=setting logical_path=app/Order.php'
  );
});

test('diagnostics reject prompts, absolute paths, and traversal', () => {
  assert.throws(() => formatEvent('command_started', { prompt: 'secret' }), /not allowlisted/);
  assert.throws(() => formatEvent('command_started', { logical_path: '/home/user/secret.php' }), /unsafe/);
  assert.throws(() => formatEvent('command_started', { logical_path: '../secret.php' }), /unsafe/);
  assert.throws(() => formatEvent('command started', {}), /unsafe/);
});

test('diagnostics encode harmless whitespace in an authorized logical path', () => {
  const line = formatEvent('command_started', { logical_path: 'app/My Order.php' });
  assert.match(line, /logical_path=app\/My%20Order.php$/);
});

test('diagnostics expose only a stable execution target, never a remote authority or host path', () => {
  const line = formatEvent('command_started', {
    command: 'doctor',
    status: 'started',
    binary_origin: 'path',
    execution_target: 'remote-wsl'
  }, new Date('2026-09-22T12:00:00.000Z'));
  assert.match(line, /execution_target=remote-wsl/);
  assert.doesNotMatch(line, /Ubuntu|\\\\wsl|\/home\//);
  assert.throws(() => formatEvent('command_started', {
    execution_target: '/home/user/private'
  }), /unsafe diagnostic value/);
});
