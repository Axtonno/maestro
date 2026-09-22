'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const { classifyWorkspaceContext } = require('../workspace-context');

test('selects single-root, active multi-root, ambiguous multi-root, and generic chat', () => {
  assert.deepEqual(classifyWorkspaceContext({ folderCount: 1 }), { kind: 'single' });
  assert.deepEqual(classifyWorkspaceContext({
    folderCount: 2, hasEditor: true, activeScheme: 'file', activeFolderFound: true
  }), { kind: 'active' });
  assert.deepEqual(classifyWorkspaceContext({ folderCount: 2 }), { kind: 'choose' });
  assert.deepEqual(classifyWorkspaceContext({ folderCount: 0 }), { kind: 'generic' });
});

test('fails closed for unsaved and outside-workspace active files', () => {
  assert.deepEqual(classifyWorkspaceContext({
    folderCount: 1, hasEditor: true, activeScheme: 'untitled', activeFolderFound: false
  }), { kind: 'error', code: 'file_not_local' });
  assert.deepEqual(classifyWorkspaceContext({
    folderCount: 1, hasEditor: true, activeScheme: 'file', activeFolderFound: false
  }), { kind: 'error', code: 'file_outside_workspace' });
});
