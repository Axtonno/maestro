'use strict';

const assert = require('node:assert/strict');
const { EventEmitter } = require('node:events');
const { PassThrough } = require('node:stream');
const test = require('node:test');
const { runCli } = require('../cli-runner');

test('runs a direct process, sends the prompt on stdin, and captures bounded output', async () => {
  const child = fakeChild();
  const called = [];
  const resultPromise = runCli({ executable: '/bin/maestro', args: ['chat'] }, {
    cwd: '/work',
    input: 'secret prompt',
    spawn(executable, args, options) {
      called.push({ executable, args, options });
      return child;
    }
  });
  child.stdout.write('mode\tchat\nresult\nok\n');
  child.stderr.end();
  child.stdout.end();
  child.emit('close', 0, null);
  const result = await resultPromise;
  assert.equal(result.exitCode, 0);
  assert.equal(result.stdout, 'mode\tchat\nresult\nok\n');
  assert.equal(child.input, 'secret prompt\n');
  assert.equal(called[0].options.shell, false);
  assert.equal(called[0].options.cwd, '/work');
  assert.deepEqual(called[0].args, ['chat']);
});

test('terminates canceled and over-limit requests', async () => {
  const cancellation = new EventEmitter();
  cancellation.onCancellationRequested = listener => {
    cancellation.on('cancel', listener);
    return { dispose: () => cancellation.off('cancel', listener) };
  };
  const canceled = fakeChild();
  const canceledPromise = runCli({ executable: '/bin/maestro', args: [] }, {
    token: cancellation,
    spawn: () => canceled
  });
  cancellation.emit('cancel');
  canceled.emit('close', null, 'SIGTERM');
  await assert.rejects(canceledPromise, error => error.code === 'command_canceled');
  assert.deepEqual(canceled.signals, ['SIGTERM']);

  const limited = fakeChild();
  const limitedPromise = runCli({ executable: '/bin/maestro', args: [] }, {
    maxOutputBytes: 3,
    spawn: () => limited
  });
  limited.stdout.write('four');
  limited.emit('close', null, 'SIGTERM');
  await assert.rejects(limitedPromise, error => error.code === 'cli_output_limit');
});

test('terminates a request that exceeds its runtime bound', async () => {
  const child = fakeChild();
  child.exitCode = null;
  child.kill = signal => {
    child.signals.push(signal);
    child.killed = true;
    queueMicrotask(() => {
      child.exitCode = 143;
      child.emit('close', null, signal);
    });
    return true;
  };
  const result = runCli({ executable: '/bin/maestro', args: [] }, {
    timeoutMs: 1,
    spawn: () => child
  });
  await assert.rejects(result, error => error.code === 'command_timeout');
  assert.deepEqual(child.signals, ['SIGTERM']);
});

function fakeChild() {
  const child = new EventEmitter();
  child.stdout = new PassThrough();
  child.stderr = new PassThrough();
  child.stdin = new PassThrough();
  child.input = '';
  child.stdin.on('data', chunk => { child.input += chunk.toString('utf8'); });
  child.signals = [];
  child.killed = false;
  child.kill = signal => {
    child.signals.push(signal);
    child.killed = true;
    return true;
  };
  return child;
}
