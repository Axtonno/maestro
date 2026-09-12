'use strict';

const path = require('node:path');
const { PreviewError } = require('./errors');

const FORBIDDEN_ARGUMENT_BYTES = /[\u0000\r\n]/;

function safeLogicalPath(workspaceRoot, filePath) {
  const relative = path.relative(workspaceRoot, filePath);
  if (relative === '' || path.isAbsolute(relative)) {
    throw new PreviewError('file_outside_workspace');
  }
  const logical = relative.split(path.sep).join('/');
  if (logical === '..' || logical.startsWith('../') || FORBIDDEN_ARGUMENT_BYTES.test(logical)) {
    throw new PreviewError('file_outside_workspace');
  }
  return logical;
}

function mutationLogicalPath(workspaceRoot, filePath) {
  const logical = safeLogicalPath(workspaceRoot, filePath);
  if (!logical.startsWith('app/') || path.posix.extname(logical).toLowerCase() !== '.php') {
    throw new PreviewError('mutation_target_invalid');
  }
  return logical;
}

function inclusiveSelectedLines(selection) {
  if (!selection || selection.isEmpty) {
    throw new PreviewError('selection_empty');
  }
  const start = selection.start.line + 1;
  const end = selection.end.character === 0 && selection.end.line > selection.start.line
    ? selection.end.line
    : selection.end.line + 1;
  if (start < 1 || end < start) {
    throw new PreviewError('selection_invalid');
  }
  return `${start}:${end}`;
}

function checkedArgument(value, field) {
  if (typeof value !== 'string' || value.length === 0 || FORBIDDEN_ARGUMENT_BYTES.test(value)) {
    throw new PreviewError(field);
  }
  return value;
}

function checkedUserText(value, emptyCode, invalidCode) {
  if (typeof value !== 'string' || value.trim() === '') {
    throw new PreviewError(emptyCode);
  }
  const cleaned = value.trim();
  if (FORBIDDEN_ARGUMENT_BYTES.test(cleaned)) {
    throw new PreviewError(invalidCode);
  }
  return cleaned;
}

function buildChatInvocation(resolved, logicalPath, question) {
  return invocation(resolved.binaryPath, [
    'chat',
    '--config', resolved.configPath,
    '--file', checkedArgument(logicalPath, 'file_outside_workspace'),
    '--', checkedUserText(question, 'question_empty', 'question_invalid')
  ]);
}

function buildMutationInvocation(resolved, logicalPath, lines, instruction) {
  return invocation(resolved.binaryPath, [
    'workspace', 'replace',
    '--file', checkedArgument(logicalPath, 'file_outside_workspace'),
    '--lines', checkedArgument(lines, 'selection_invalid'),
    '--config', resolved.configPath,
    '--', checkedUserText(instruction, 'instruction_empty', 'instruction_invalid')
  ]);
}

function buildDoctorInvocation(resolved) {
  return invocation(resolved.binaryPath, ['doctor', '--mode', 'all', '--config', resolved.configPath]);
}

function buildVersionInvocation(resolved) {
  return invocation(resolved.binaryPath, ['version', '--diagnostic']);
}

function invocation(executable, args) {
  return Object.freeze({
    executable: checkedArgument(executable, 'binary_not_executable'),
    args: Object.freeze(args.map(value => checkedArgument(value, 'command_unavailable')))
  });
}

module.exports = {
  buildChatInvocation,
  buildDoctorInvocation,
  buildMutationInvocation,
  buildVersionInvocation,
  inclusiveSelectedLines,
  mutationLogicalPath,
  safeLogicalPath
};
