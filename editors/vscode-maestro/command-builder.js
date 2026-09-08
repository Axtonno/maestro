'use strict';

const path = require('path');

const FORBIDDEN_TOKEN_BYTES = /[\u0000\r\n]/;

function shellQuote(value, field = 'argument') {
  if (typeof value !== 'string' || value.length === 0) {
    throw new Error(`${field} must not be empty`);
  }
  if (FORBIDDEN_TOKEN_BYTES.test(value)) {
    throw new Error(`${field} contains a forbidden control character`);
  }
  return "'" + value.replace(/'/g, "'\"'\"'") + "'";
}

function safeLogicalPath(workspaceRoot, filePath) {
  const relative = path.relative(workspaceRoot, filePath);
  if (relative === '' || path.isAbsolute(relative)) {
    throw new Error('the active editor is not a workspace file');
  }
  const logical = relative.split(path.sep).join('/');
  if (logical === '..' || logical.startsWith('../') || FORBIDDEN_TOKEN_BYTES.test(logical)) {
    throw new Error('the active editor is outside the workspace');
  }
  return logical;
}

function mutationLogicalPath(workspaceRoot, filePath) {
  const logical = safeLogicalPath(workspaceRoot, filePath);
  if (!logical.startsWith('app/') || path.posix.extname(logical).toLowerCase() !== '.php') {
    throw new Error('Controlled Mutation requires a PHP file below app/');
  }
  return logical;
}

function inclusiveSelectedLines(selection) {
  if (!selection || selection.isEmpty) {
    throw new Error('select at least one character');
  }
  const start = selection.start.line + 1;
  const end = selection.end.character === 0 && selection.end.line > selection.start.line
    ? selection.end.line
    : selection.end.line + 1;
  if (start < 1 || end < start) {
    throw new Error('the editor selection is invalid');
  }
  return `${start}:${end}`;
}

function configuredTokens(settings) {
  const binary = settings.binaryPath.trim();
  const config = settings.configPath.trim();
  const tokens = [shellQuote(binary, 'maestro.binaryPath')];
  if (config !== '') {
    tokens.push('--config', shellQuote(config, 'maestro.configPath'));
  }
  return { binary: tokens[0], config: tokens.slice(1) };
}

function buildChatCommand(settings, logicalPath, question) {
  const configured = configuredTokens(settings);
  return [
    configured.binary,
    'chat',
    ...configured.config,
    '--file',
    shellQuote(logicalPath, 'file'),
    '--',
    shellQuote(question.trim(), 'question')
  ].join(' ');
}

function buildMutationCommand(settings, logicalPath, lines, instruction) {
  const configured = configuredTokens(settings);
  return [
    configured.binary,
    'workspace',
    'replace',
    '--file',
    shellQuote(logicalPath, 'file'),
    '--lines',
    shellQuote(lines, 'lines'),
    ...configured.config,
    '--',
    shellQuote(instruction.trim(), 'instruction')
  ].join(' ');
}

function buildDoctorCommand(settings) {
  const configured = configuredTokens(settings);
  return [configured.binary, 'doctor', '--mode', 'all', ...configured.config].join(' ');
}

function buildVersionCommand(settings) {
  const configured = configuredTokens(settings);
  return [configured.binary, 'version', '--diagnostic'].join(' ');
}

module.exports = {
  buildChatCommand,
  buildDoctorCommand,
  buildMutationCommand,
  buildVersionCommand,
  inclusiveSelectedLines,
  mutationLogicalPath,
  safeLogicalPath,
  shellQuote
};
