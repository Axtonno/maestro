'use strict';

const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const yauzl = require('yauzl');
const yazl = require('yazl');

const root = path.resolve(__dirname, '..');
const output = path.join(root, 'dist', 'maestro-local-ai-0.1.1.vsix');
const fixedTime = new Date('1980-01-01T00:00:00.000Z');

async function main() {
  fs.mkdirSync(path.dirname(output), { recursive: true });
  const vsce = require.resolve('@vscode/vsce/vsce');
  const result = spawnSync(process.execPath, [
    vsce,
    'package',
    '--pre-release',
    '--no-dependencies',
    '--out', output
  ], { cwd: root, encoding: 'utf8', stdio: 'inherit' });
  if (result.error) {
    throw result.error;
  }
  if (result.status !== 0) {
    throw new Error(`vsce package failed with status ${result.status}`);
  }

  const entries = await readArchive(output);
  await writeDeterministicArchive(output, entries);
  process.stdout.write(`Normalized ${entries.length} VSIX entries to a fixed timestamp.\n`);
}

function readArchive(file) {
  return new Promise((resolve, reject) => {
    yauzl.open(file, { lazyEntries: true }, (error, archive) => {
      if (error) {
        reject(error);
        return;
      }
      const entries = [];
      archive.on('error', reject);
      archive.on('end', () => resolve(entries));
      archive.on('entry', entry => {
        archive.openReadStream(entry, (streamError, stream) => {
          if (streamError) {
            reject(streamError);
            return;
          }
          const chunks = [];
          stream.on('error', reject);
          stream.on('data', chunk => chunks.push(chunk));
          stream.on('end', () => {
            entries.push({
              name: entry.fileName,
              content: Buffer.concat(chunks),
              mode: (entry.externalFileAttributes >>> 16) & 0xffff
            });
            archive.readEntry();
          });
        });
      });
      archive.readEntry();
    });
  });
}

function writeDeterministicArchive(file, entries) {
  return new Promise((resolve, reject) => {
    const temporary = `${file}.normalized`;
    const archive = new yazl.ZipFile();
    for (const entry of entries) {
      archive.addBuffer(entry.content, entry.name, {
        mtime: fixedTime,
        mode: entry.mode || 0o100644,
        compress: true
      });
    }
    const destination = fs.createWriteStream(temporary, { mode: 0o600 });
    archive.outputStream.on('error', reject);
    destination.on('error', reject);
    destination.on('close', () => {
      fs.copyFileSync(temporary, file);
      fs.unlinkSync(temporary);
      resolve();
    });
    archive.outputStream.pipe(destination);
    archive.end();
  });
}

main().catch(error => {
  console.error(error);
  process.exitCode = 1;
});
