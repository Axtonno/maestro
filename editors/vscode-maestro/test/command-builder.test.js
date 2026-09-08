'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const {
  buildChatCommand,
  buildDoctorCommand,
  buildMutationCommand,
  buildVersionCommand,
  inclusiveSelectedLines,
  mutationLogicalPath,
  safeLogicalPath,
  shellQuote
} = require('../command-builder');

const settings = {
  binaryPath: '/opt/Maestro/bin/maestro',
  configPath: './configs/maestro.v0.5.0-candidate.yaml'
};

test('quotes every user-controlled shell token', () => {
  assert.equal(shellQuote("a'b"), `'a'"'"'b'`);
  assert.throws(() => shellQuote('a\nb'), /forbidden/);
  assert.equal(
    buildChatCommand(settings, 'app/Order.php', "what's returned?"),
    `'/opt/Maestro/bin/maestro' chat --config './configs/maestro.v0.5.0-candidate.yaml' --file 'app/Order.php' -- 'what'"'"'s returned?'`
  );
  assert.equal(
    buildMutationCommand(settings, 'app/Order.php', '2:3', 'change $x; echo nope'),
    `'/opt/Maestro/bin/maestro' workspace replace --file 'app/Order.php' --lines '2:3' --config './configs/maestro.v0.5.0-candidate.yaml' -- 'change $x; echo nope'`
  );
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
  assert.throws(() => inclusiveSelectedLines({ isEmpty: true }), /select/);
});

test('contains paths and narrows mutation targets', () => {
  assert.equal(safeLogicalPath('/work/project', '/work/project/src/a.go'), 'src/a.go');
  assert.equal(mutationLogicalPath('/work/project', '/work/project/app/A.php'), 'app/A.php');
  assert.throws(() => safeLogicalPath('/work/project', '/work/other/A.php'), /outside/);
  assert.throws(() => mutationLogicalPath('/work/project', '/work/project/src/A.php'), /below app/);
  assert.throws(() => mutationLogicalPath('/work/project', '/work/project/app/A.js'), /PHP/);
});

test('doctor and identity remain explicit terminal commands', () => {
  assert.equal(
    buildDoctorCommand(settings),
    `'/opt/Maestro/bin/maestro' doctor --mode all --config './configs/maestro.v0.5.0-candidate.yaml'`
  );
  assert.equal(buildVersionCommand(settings), `'/opt/Maestro/bin/maestro' version --diagnostic`);
});
