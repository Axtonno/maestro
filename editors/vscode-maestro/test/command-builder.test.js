'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const {
  buildChatInvocation,
  buildDoctorInvocation,
  buildMutationInvocation,
  buildVersionInvocation,
  inclusiveSelectedLines,
  mutationLogicalPath,
  safeLogicalPath
} = require('../command-builder');

const resolved = {
  binaryPath: '/opt/Maestro bin/maestro',
  configPath: '/work/project/configs/maestro.yaml'
};

test('builds direct process invocations without shell reconstruction', () => {
  assert.deepEqual(buildChatInvocation(resolved, 'app/Order.php', "what's returned?"), {
    executable: '/opt/Maestro bin/maestro',
    args: [
      'chat', '--config', '/work/project/configs/maestro.yaml',
      '--file', 'app/Order.php', '--', "what's returned?"
    ]
  });
  assert.deepEqual(buildMutationInvocation(resolved, 'app/Order.php', '2:3', 'change $x; echo nope'), {
    executable: '/opt/Maestro bin/maestro',
    args: [
      'workspace', 'replace', '--file', 'app/Order.php', '--lines', '2:3',
      '--config', '/work/project/configs/maestro.yaml', '--', 'change $x; echo nope'
    ]
  });
});

test('maps an editor selection to inclusive one-based lines', () => {
  assert.equal(inclusiveSelectedLines({
    isEmpty: false,
    start: { line: 4, character: 2 },
    end: { line: 6, character: 3 }
  }), '5:7');
  assert.equal(inclusiveSelectedLines({
    isEmpty: false,
    start: { line: 4, character: 0 },
    end: { line: 6, character: 0 }
  }), '5:6');
  assert.throws(() => inclusiveSelectedLines({ isEmpty: true }), error => error.code === 'selection_empty');
});

test('contains paths and narrows mutation targets', () => {
  assert.equal(safeLogicalPath('/work/project', '/work/project/src/a.go'), 'src/a.go');
  assert.equal(mutationLogicalPath('/work/project', '/work/project/app/A.php'), 'app/A.php');
  assert.throws(() => safeLogicalPath('/work/project', '/work/other/A.php'), error => error.code === 'file_outside_workspace');
  assert.throws(() => mutationLogicalPath('/work/project', '/work/project/src/A.php'), error => error.code === 'mutation_target_invalid');
  assert.throws(() => mutationLogicalPath('/work/project', '/work/project/app/A.js'), error => error.code === 'mutation_target_invalid');
});

test('doctor and identity keep fixed argument vectors', () => {
  assert.deepEqual(buildDoctorInvocation(resolved), {
    executable: '/opt/Maestro bin/maestro',
    args: ['doctor', '--mode', 'all', '--config', '/work/project/configs/maestro.yaml']
  });
  assert.deepEqual(buildVersionInvocation(resolved), {
    executable: '/opt/Maestro bin/maestro',
    args: ['version', '--diagnostic']
  });
});

test('rejects control bytes and empty direct command input', () => {
  assert.throws(
    () => buildChatInvocation(resolved, 'app/Order.php', '  '),
    error => error.code === 'question_empty'
  );
  assert.throws(
    () => buildMutationInvocation(resolved, 'app/Order.php', '2:3', 'bad\nargument'),
    error => error.code === 'instruction_invalid'
  );
});
