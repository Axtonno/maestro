'use strict';

const path = require('node:path');
const Mocha = require('mocha');

async function run() {
  const mocha = new Mocha({ ui: 'tdd', color: true, timeout: 10000 });
  mocha.addFile(path.resolve(__dirname, 'installed.test.js'));
  await new Promise((resolve, reject) => {
    mocha.run(failures => failures > 0 ? reject(new Error(`${failures} VSIX tests failed`)) : resolve());
  });
}

module.exports = { run };
