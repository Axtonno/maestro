'use strict';

function classifyWorkspaceContext(input) {
  const folderCount = Number.isInteger(input.folderCount) ? input.folderCount : 0;
  if (input.hasEditor) {
    if (input.activeScheme !== 'file') {
      return Object.freeze({ kind: 'error', code: 'file_not_local' });
    }
    if (!input.activeFolderFound) {
      return Object.freeze({ kind: 'error', code: 'file_outside_workspace' });
    }
    return Object.freeze({ kind: 'active' });
  }
  if (folderCount === 0) {
    return Object.freeze({ kind: 'generic' });
  }
  if (folderCount === 1) {
    return Object.freeze({ kind: 'single' });
  }
  return Object.freeze({ kind: 'choose' });
}

module.exports = { classifyWorkspaceContext };
