'use strict';

const path = require('node:path');
const Mocha = require('mocha');

function run() {
  const mocha = new Mocha({ ui: 'tdd', color: true, timeout: 20000 });
  mocha.addFile(path.resolve(__dirname, 'integration.test.js'));
  return new Promise((resolve, reject) => {
    mocha.run(failures => failures === 0
      ? resolve()
      : reject(new Error(`${failures} VS Code integration test(s) failed`)));
  });
}

module.exports = { run };
