'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const { ERROR_CATALOG, PreviewError, normalizeError } = require('../errors');

test('every user error has a stable code, short title, cause, and one action', () => {
  assert.ok(Object.keys(ERROR_CATALOG).length >= 20);
  for (const [code, entry] of Object.entries(ERROR_CATALOG)) {
    assert.match(code, /^[a-z0-9_]+$/);
    assert.equal(entry.length, 3);
    assert.ok(entry.every(value => typeof value === 'string' && value.length > 0));
    const error = new PreviewError(code);
    assert.equal(error.code, code);
    assert.equal(error.title, entry[0]);
    assert.equal(error.cause, entry[1]);
    assert.equal(error.action, entry[2]);
  }
});

test('unexpected implementation failures collapse to a non-sensitive error', () => {
  const normalized = normalizeError(new Error('API_KEY=do-not-leak'));
  assert.equal(normalized.code, 'command_unavailable');
  assert.doesNotMatch(normalized.cause, /API_KEY|do-not-leak/);
});
