'use strict';

const childProcess = require('node:child_process');
const { PreviewError } = require('./errors');

const DEFAULT_MAX_OUTPUT_BYTES = 2 * 1024 * 1024;
const DEFAULT_TIMEOUT_MS = 5 * 60 * 1000;

function runCli(invocation, options = {}) {
  const spawn = options.spawn || childProcess.spawn;
  const timeoutMs = positiveInteger(options.timeoutMs, DEFAULT_TIMEOUT_MS);
  const maxOutputBytes = positiveInteger(options.maxOutputBytes, DEFAULT_MAX_OUTPUT_BYTES);
  const cwd = options.cwd;
  const input = typeof options.input === 'string' ? options.input : '';
  const token = options.token;

  return new Promise((resolve, reject) => {
    let child;
    try {
      child = spawn(invocation.executable, invocation.args, {
        cwd,
        env: process.env,
        shell: false,
        windowsHide: true,
        stdio: ['pipe', 'pipe', 'pipe']
      });
    } catch {
      reject(new PreviewError('command_unavailable'));
      return;
    }

    let stdout = Buffer.alloc(0);
    let stderr = Buffer.alloc(0);
    let finished = false;
    let failure;
    const append = (current, chunk) => {
      const next = Buffer.concat([current, Buffer.from(chunk)]);
      if (stdout.length + stderr.length + Buffer.byteLength(chunk) > maxOutputBytes) {
        failure = new PreviewError('cli_output_limit');
        stopChild(child);
        return current;
      }
      return next;
    };
    child.stdout.on('data', chunk => { stdout = append(stdout, chunk); });
    child.stderr.on('data', chunk => { stderr = append(stderr, chunk); });

    const timeout = setTimeout(() => {
      failure = new PreviewError('command_timeout');
      stopChild(child);
    }, timeoutMs);
    const cancellation = token && typeof token.onCancellationRequested === 'function'
      ? token.onCancellationRequested(() => {
        failure = new PreviewError('command_canceled');
        stopChild(child);
      })
      : undefined;

    const finish = callback => {
      if (finished) {
        return;
      }
      finished = true;
      clearTimeout(timeout);
      if (cancellation && typeof cancellation.dispose === 'function') {
        cancellation.dispose();
      }
      callback();
    };
    child.once('error', () => finish(() => reject(new PreviewError('command_unavailable'))));
    child.once('close', (code, signal) => finish(() => {
      if (failure) {
        reject(failure);
        return;
      }
      resolve(Object.freeze({
        exitCode: Number.isInteger(code) ? code : -1,
        signal: signal || '',
        stdout: stdout.toString('utf8'),
        stderr: stderr.toString('utf8')
      }));
    }));

    child.stdin.once('error', () => {});
    child.stdin.end(input === '' ? '' : `${input}\n`);
  });
}

function stopChild(child) {
  if (!child || child.killed) {
    return;
  }
  child.kill('SIGTERM');
  const force = setTimeout(() => {
    if (child.exitCode === null || child.exitCode === undefined) {
      child.kill('SIGKILL');
    }
  }, 2000);
  if (typeof force.unref === 'function') {
    force.unref();
  }
}

function positiveInteger(value, fallback) {
  return Number.isInteger(value) && value > 0 ? value : fallback;
}

module.exports = { DEFAULT_MAX_OUTPUT_BYTES, DEFAULT_TIMEOUT_MS, runCli };
