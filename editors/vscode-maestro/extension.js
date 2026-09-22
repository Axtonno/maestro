'use strict';

const os = require('node:os');
const vscode = require('vscode');
const { runCli } = require('./cli-runner');
const { classifyCliFailure, classifyDoctorOutput, parseChatEnvelope, parseProfileIdentity } = require('./chat-protocol');
const commands = require('./command-builder');
const { formatEvent } = require('./diagnostics');
const { PreviewError, normalizeError } = require('./errors');
const { requireExecutionEnvironment } = require('./execution-environment');
const { ONBOARDING_STATES, inspectOnboarding } = require('./onboarding');
const { resolveBinaryPath, resolveConfigPath } = require('./path-resolver');
const { classifyWorkspaceContext } = require('./workspace-context');

const EXTENSION_ID = 'axtonno.maestro-local-ai';
const WALKTHROUGH_ID = `${EXTENSION_ID}#maestro.setup`;
const CHAT_PARTICIPANT_ID = 'maestro.chat';
const GUIDE_ACTION = 'Open Setup Guide';
const SETTINGS_ACTION = 'Open Settings';
let outputChannel;
let statusItem;

function settingsFor(scope) {
  const configuration = vscode.workspace.getConfiguration('maestro', scope);
  return {
    binaryPath: configuration.get('binaryPath', ''),
    configPath: configuration.get('configPath', ''),
    profile: configuration.get('profile', 'recommended'),
    terminalName: configuration.get('terminalName', 'Maestro')
  };
}

function currentExecutionEnvironment() {
  const extension = vscode.extensions.getExtension(EXTENSION_ID);
  return requireExecutionEnvironment({
    platform: process.platform,
    remoteName: vscode.env.remoteName,
    workspaceExtension: Boolean(extension && extension.extensionKind === vscode.ExtensionKind.Workspace)
  }, PreviewError);
}

function requireSupportedHost() {
  if (!vscode.workspace.isTrusted) {
    throw new PreviewError('workspace_untrusted');
  }
  return currentExecutionEnvironment();
}

function activeWorkspaceEditor(options = {}) {
  const editor = vscode.window.activeTextEditor;
  if (!editor || editor.document.uri.scheme !== 'file') {
    throw new PreviewError('file_not_local');
  }
  const folder = vscode.workspace.getWorkspaceFolder(editor.document.uri);
  if (!folder) {
    throw new PreviewError('file_outside_workspace');
  }
  requireLocalFolder(folder);
  if (editor.document.isDirty) {
    throw new PreviewError('file_dirty');
  }
  const logicalPath = options.mutation
    ? commands.mutationLogicalPath(folder.uri.fsPath, editor.document.uri.fsPath)
    : commands.safeLogicalPath(folder.uri.fsPath, editor.document.uri.fsPath);
  return { editor, folder, logicalPath };
}

function resolveRuntime(folder, options = {}) {
  const execution = currentExecutionEnvironment();
  const scope = folder && folder.uri;
  const settings = settingsFor(scope);
  const workspaceRoot = folder ? folder.uri.fsPath : undefined;
  if (!workspaceRoot && settings.configPath.trim() !== '') {
    throw new PreviewError('workspace_not_open');
  }
  const binary = resolveBinaryPath({
    configuredPath: settings.binaryPath,
    workspaceRoot: workspaceRoot || process.cwd(),
    hostPath: process.env.PATH,
    hostCwd: process.cwd()
  });
  const config = options.configRequired && settings.configPath.trim() !== ''
    ? resolveConfigPath(settings.configPath, workspaceRoot)
    : undefined;
  return {
    binaryPath: binary.path,
    binaryOrigin: binary.origin,
    configPath: config && config.path,
    configLogicalPath: config && config.logicalPath,
    configOrigin: config && config.origin,
    profile: settings.profile,
    executionTarget: execution.target,
    terminalName: validateTerminalName(settings.terminalName)
  };
}

async function askActiveFile(providedQuestion) {
  requireSupportedHost();
  const target = activeWorkspaceEditor();
  const question = typeof providedQuestion === 'string'
    ? providedQuestion
    : await vscode.window.showInputBox({
      title: 'Maestro: Chat About Active File',
      prompt: `Question for ${target.logicalPath}`,
      ignoreFocusOut: true,
      validateInput: value => value.trim() === '' ? 'Enter a question.' : undefined
    });
  if (question === undefined) {
    return;
  }
  if (question.trim() === '') {
    throw new PreviewError('question_empty');
  }
  const runtime = resolveRuntime(target.folder, { configRequired: true });
  showResolution(runtime);
  await launch(target.folder, runtime, commands.buildChatInvocation(runtime, target.logicalPath, question), {
    command: 'chat',
    logicalPath: target.logicalPath
  });
}

async function replaceSelection(providedInstruction) {
  requireSupportedHost();
  const target = activeWorkspaceEditor({ mutation: true });
  if (target.editor.selections.length !== 1) {
    throw new PreviewError('selection_multiple');
  }
  requireWholeLineSelection(target.editor);
  const lines = commands.inclusiveSelectedLines(target.editor.selection);
  const instruction = typeof providedInstruction === 'string'
    ? providedInstruction
    : await vscode.window.showInputBox({
      title: 'Maestro: Mutate Selection',
      prompt: `Instruction for ${target.logicalPath}:${lines}`,
      placeHolder: 'Describe only the replacement for the selected lines.',
      ignoreFocusOut: true,
      validateInput: value => value.trim() === '' ? 'Enter a mutation instruction.' : undefined
    });
  if (instruction === undefined) {
    return;
  }
  if (instruction.trim() === '') {
    throw new PreviewError('instruction_empty');
  }
  const runtime = resolveRuntime(target.folder, { configRequired: true });
  showResolution(runtime);
  await launch(target.folder, runtime, commands.buildMutationInvocation(runtime, target.logicalPath, lines, instruction), {
    command: 'mutation',
    logicalPath: target.logicalPath
  });
  void vscode.window.showInformationMessage(
    'Review Maestro\'s complete preview in the terminal, then allow once or deny there.'
  );
}

function requireWholeLineSelection(editor) {
  const selection = editor.selection;
  if (selection.isEmpty) {
    throw new PreviewError('selection_empty');
  }
  const endsAtNextLineStart = selection.end.character === 0 && selection.end.line > selection.start.line;
  const endsAtLineEnd = selection.end.character === editor.document.lineAt(selection.end.line).text.length;
  if (selection.start.character !== 0 || (!endsAtNextLineStart && !endsAtLineEnd)) {
    throw new PreviewError('selection_partial');
  }
}

async function doctor() {
  requireSupportedHost();
  const folder = await selectWorkspaceFolder();
  if (!folder) {
    return;
  }
  const runtime = resolveRuntime(folder, { configRequired: true });
  showResolution(runtime);
  await launch(folder, runtime, commands.buildDoctorInvocation(runtime), { command: 'doctor' });
}

async function version() {
  requireSupportedHost();
  const folder = await selectWorkspaceFolder();
  if (!folder) {
    return;
  }
  const runtime = resolveRuntime(folder);
  showResolution(runtime);
  await launch(folder, runtime, commands.buildVersionInvocation(runtime), { command: 'version' });
}

async function openSetupGuide() {
  await vscode.commands.executeCommand('workbench.action.openWalkthrough', WALKTHROUGH_ID, false);
}

async function openNativeChat() {
  await vscode.commands.executeCommand('workbench.action.chat.open', { query: '@maestro ', isPartialQuery: true });
}

async function openTroubleshooting() {
  const extension = vscode.extensions.getExtension(EXTENSION_ID);
  if (!extension) {
    throw new PreviewError('command_unavailable');
  }
  const document = await vscode.workspace.openTextDocument(vscode.Uri.joinPath(extension.extensionUri, 'README.md'));
  await vscode.window.showTextDocument(document, { preview: true });
}

async function selectWorkspaceFolder() {
  const folders = vscode.workspace.workspaceFolders;
  if (!folders || folders.length === 0) {
    throw new PreviewError('workspace_not_open');
  }
  if (folders.length === 1) {
    requireLocalFolder(folders[0]);
    return folders[0];
  }
  const editor = vscode.window.activeTextEditor;
  const active = editor && editor.document.uri.scheme === 'file'
    ? vscode.workspace.getWorkspaceFolder(editor.document.uri)
    : undefined;
  if (!active) {
    const selected = await vscode.window.showQuickPick(
      folders.map(folder => ({ label: folder.name, description: folder.uri.fsPath, folder })),
      { title: 'Select the workspace folder Maestro should use', placeHolder: 'Workspace root' }
    );
    if (!selected) {
      return undefined;
    }
    requireLocalFolder(selected.folder);
    return selected.folder;
  }
  requireLocalFolder(active);
  return active;
}

function requireLocalFolder(folder) {
  if (!folder || folder.uri.scheme !== 'file') {
    throw new PreviewError('workspace_virtual');
  }
}

function validateTerminalName(value) {
  const name = typeof value === 'string' ? value.trim() : '';
  if (name === '' || name.length > 80 || /[\u0000-\u001f\u007f]/.test(name)) {
    throw new PreviewError('terminal_name_invalid');
  }
  return name;
}

async function launch(folder, runtime, invocation, metadata) {
  const definition = {
    type: 'maestro-preview',
    command: metadata.command,
    binaryOrigin: runtime.binaryOrigin,
    executionTarget: runtime.executionTarget,
    logicalPath: metadata.logicalPath,
    startedAt: Date.now()
  };
  const execution = new vscode.ProcessExecution(invocation.executable, invocation.args, {
    cwd: folder.uri.fsPath
  });
  const task = new vscode.Task(
    definition,
    folder,
    runtime.terminalName,
    'Maestro',
    execution,
    []
  );
  task.presentationOptions = {
    reveal: vscode.TaskRevealKind.Always,
    panel: vscode.TaskPanelKind.New,
    focus: true,
    echo: true,
    showReuseMessage: false,
    clear: false,
    close: false
  };
  writeEvent('command_started', {
    command: metadata.command,
    status: 'started',
    binary_origin: runtime.binaryOrigin,
    execution_target: runtime.executionTarget,
    logical_path: metadata.logicalPath
  });
  try {
    await vscode.tasks.executeTask(task);
  } catch {
    throw new PreviewError('command_unavailable');
  }
}

function onTaskEnded(event) {
  const definition = event.execution.task.definition;
  if (!definition || definition.type !== 'maestro-preview') {
    return;
  }
  const exitCode = Number.isInteger(event.exitCode) ? event.exitCode : -1;
  const duration = Math.max(0, Date.now() - Number(definition.startedAt || Date.now()));
  const denied = definition.command === 'mutation' && exitCode === 3;
  const passed = exitCode === 0;
  writeEvent(denied ? 'command_denied' : 'command_finished', {
    command: definition.command,
    status: denied ? 'denied' : passed ? 'passed' : 'failed',
    duration_ms: duration,
    exit_code: exitCode,
    binary_origin: definition.binaryOrigin,
    execution_target: definition.executionTarget,
    logical_path: definition.logicalPath,
    error_code: passed || denied ? undefined : 'cli_exit_nonzero'
  });
  if (!passed && !denied) {
    void presentError(new PreviewError('cli_exit_nonzero'));
  }
  if (passed && definition.command === 'doctor') {
    void vscode.commands.executeCommand('setContext', 'maestro.doctorPassed', true);
  }
  if (passed && definition.command === 'chat') {
    void vscode.commands.executeCommand('setContext', 'maestro.chatPassed', true);
  }
  if ((passed || denied) && definition.command === 'mutation') {
    void vscode.commands.executeCommand('setContext', 'maestro.mutationReviewed', true);
  }
  void refreshOnboardingStatus();
}

function showResolution(runtime) {
  const origin = runtime.binaryOrigin === 'setting' ? 'Settings' : 'extension host PATH';
  const config = runtime.configPath
    ? `; config from Settings: ${runtime.configPath}`
    : '';
  vscode.window.setStatusBarMessage(
    `Maestro — ${runtime.executionTarget}; binary from ${origin}: ${runtime.binaryPath}${config}`,
    10000
  );
}

async function nativeChatTarget() {
  requireSupportedHost();
  const editor = vscode.window.activeTextEditor;
  const folders = vscode.workspace.workspaceFolders;
  const activeFolder = editor && editor.document.uri.scheme === 'file'
    ? vscode.workspace.getWorkspaceFolder(editor.document.uri)
    : undefined;
  const decision = classifyWorkspaceContext({
    folderCount: folders ? folders.length : 0,
    hasEditor: Boolean(editor),
    activeScheme: editor && editor.document.uri.scheme,
    activeFolderFound: Boolean(activeFolder)
  });
  if (decision.kind === 'error') {
    throw new PreviewError(decision.code);
  }
  if (decision.kind === 'active') {
    const folder = activeFolder;
    requireLocalFolder(folder);
    if (editor.document.isDirty) {
      throw new PreviewError('file_dirty');
    }
    const selection = editor.selection && !editor.selection.isEmpty
      ? commands.inclusiveSelectedLines(editor.selection)
      : undefined;
    return {
      folder,
      logicalPath: commands.safeLogicalPath(folder.uri.fsPath, editor.document.uri.fsPath),
      selection,
      workspaceName: folder.name
    };
  }
  if (decision.kind === 'generic') {
    return { folder: undefined, logicalPath: undefined, selection: undefined, workspaceName: 'none' };
  }
  const folder = await selectWorkspaceFolder();
  if (!folder) {
    return undefined;
  }
  return { folder, logicalPath: undefined, selection: undefined, workspaceName: folder.name };
}

async function inspectEffectiveProfile(target, runtime, token) {
  const result = await runCli(commands.buildProfileInvocation(runtime, Boolean(target.folder)), {
    cwd: target.folder ? target.folder.uri.fsPath : os.homedir(),
    timeoutMs: 10000,
    token
  });
  if (result.exitCode !== 0) {
    throw classifyCliFailure(result.stderr);
  }
  const identity = parseProfileIdentity(result.stdout);
  if (identity.profile !== runtime.profile) {
    throw new PreviewError('profile_mismatch');
  }
  return identity;
}

function renderIdentity(stream, identity, target, mode) {
  const execution = currentExecutionEnvironment();
  const rows = [
    ['Profile', identity.profile],
    ['Chat model', identity.chatModel],
    ['Mutation model', identity.mutationModel],
    ['Extension host', execution.target],
    ['Workspace', target.workspaceName],
    ['Mode', mode]
  ];
  if (target.logicalPath) {
    rows.push(['Context', target.selection ? `${target.logicalPath}:${target.selection}` : target.logicalPath]);
  }
  const table = ['| Maestro | Active value |', '| --- | --- |', ...rows.map(([key, value]) => `| ${key} | ${escapeTable(value)} |`)].join('\n');
  stream.markdown(`${table}\n\n`);
}

async function handleNativeChat(request, stream, token) {
  try {
    const target = await nativeChatTarget();
    if (!target) {
      return { metadata: { command: request.command || '', status: 'canceled' } };
    }
    if (request.command === 'preview') {
      const mutation = activeWorkspaceEditor({ mutation: true });
      if (mutation.editor.selections.length !== 1) {
        throw new PreviewError('selection_multiple');
      }
      requireWholeLineSelection(mutation.editor);
      const selectedLines = commands.inclusiveSelectedLines(mutation.editor.selection);
      const runtime = resolveRuntime(mutation.folder, { configRequired: true });
      const identity = await inspectEffectiveProfile({
        folder: mutation.folder,
        logicalPath: mutation.logicalPath,
        selection: selectedLines,
        workspaceName: mutation.folder.name
      }, runtime, token);
      renderIdentity(stream, identity, {
        folder: mutation.folder,
        logicalPath: mutation.logicalPath,
        selection: selectedLines,
        workspaceName: mutation.folder.name
      }, 'preview → controlled mutation');
      await replaceSelection(request.prompt);
      stream.markdown('The authoritative preview and **allow once / deny** decision are open in the Maestro terminal. The extension cannot approve or apply the change.');
      return { metadata: { command: 'preview', status: 'terminal' } };
    }
    if (request.command === 'doctor') {
      if (!target.folder) {
        throw new PreviewError('workspace_not_open');
      }
      await doctor();
      stream.markdown('Doctor is running in the Maestro terminal, which contains the authoritative diagnostics.');
      return { metadata: { command: 'doctor', status: 'terminal' } };
    }

    const runtime = resolveRuntime(target.folder, { configRequired: true });
    const identity = await inspectEffectiveProfile(target, runtime, token);
    renderIdentity(stream, identity, target, request.command === 'status' ? 'diagnostics' : 'chat');
    if (request.command === 'status') {
      stream.progress('Checking the local provider and qualified models…');
      const diagnostic = await runCli(commands.buildCapturedDoctorInvocation(runtime, Boolean(target.folder)), {
        cwd: target.folder ? target.folder.uri.fsPath : os.homedir(),
        timeoutMs: 30000,
        token
      });
      const issue = classifyDoctorOutput(diagnostic.stdout, diagnostic.stderr);
      if (issue) {
        throw new PreviewError(issue);
      }
      stream.markdown('Configuration, provider, Direct Chat model, and Controlled Mutation model are available.');
      stream.button({ command: 'maestro.doctor', title: 'Run full Doctor in terminal' });
      return { metadata: { command: 'status', status: 'ready' } };
    }

    if (typeof request.prompt !== 'string' || request.prompt.trim() === '') {
      throw new PreviewError('question_empty');
    }
    const question = target.selection
      ? `Focus on lines ${target.selection} of ${target.logicalPath}.\n\n${request.prompt.trim()}`
      : request.prompt.trim();
    stream.progress('Asking the local Maestro model…');
    writeEvent('chat_request_started', {
      command: 'native_chat', status: 'started', binary_origin: runtime.binaryOrigin,
      execution_target: runtime.executionTarget,
      logical_path: target.logicalPath
    });
    const started = Date.now();
    const result = await runCli(commands.buildCapturedChatInvocation(runtime, target.logicalPath, Boolean(target.folder)), {
      cwd: target.folder ? target.folder.uri.fsPath : os.homedir(),
      input: question,
      token
    });
    if (result.exitCode !== 0) {
      throw classifyCliFailure(result.stderr);
    }
    const response = parseChatEnvelope(result.stdout);
    if (response.model !== identity.chatModel) {
      throw new PreviewError('cli_incompatible');
    }
    const markdown = new vscode.MarkdownString();
    markdown.isTrusted = false;
    markdown.supportHtml = false;
    markdown.appendText(response.content);
    stream.markdown(markdown);
    writeEvent('chat_request_finished', {
      command: 'native_chat', status: 'passed', duration_ms: Date.now() - started,
      execution_target: runtime.executionTarget,
      binary_origin: runtime.binaryOrigin, logical_path: target.logicalPath
    });
    await vscode.commands.executeCommand('setContext', 'maestro.chatPassed', true);
    return { metadata: { command: '', status: 'passed' } };
  } catch (error) {
    const normalized = normalizeError(error);
    writeEvent('chat_request_rejected', { command: 'native_chat', status: 'rejected', error_code: normalized.code });
    const message = new vscode.MarkdownString();
    message.appendMarkdown(`**${normalized.title}**  \n`);
    message.appendText(`${normalized.cause} Next: ${normalized.action}.`);
    stream.markdown(message);
    const action = chatAction(normalized.action);
    if (action) {
      stream.button(action);
    }
    return { metadata: { command: request.command || '', status: 'rejected', errorCode: normalized.code } };
  }
}

function chatAction(action) {
  const actions = {
    'Open Settings': { command: 'workbench.action.openSettings', title: 'Open Maestro Settings', arguments: [`@ext:${EXTENSION_ID}`] },
    'Open Setup Guide': { command: 'maestro.openSetupGuide', title: 'Open Setup Guide' },
    'Open Troubleshooting': { command: 'maestro.openTroubleshooting', title: 'Open Troubleshooting' },
    'Run Doctor': { command: 'maestro.doctor', title: 'Run Doctor' },
    'Save File': { command: 'workbench.action.files.save', title: 'Save File' }
  };
  return actions[action];
}

function escapeTable(value) {
  return String(value).replace(/\\/g, '\\\\').replace(/\|/g, '\\|').replace(/[\r\n]/g, ' ');
}

function writeEvent(eventCode, fields) {
  outputChannel.appendLine(formatEvent(eventCode, fields));
}

async function presentError(error) {
  writeEvent('command_rejected', { status: 'rejected', error_code: error.code });
  const selected = await vscode.window.showErrorMessage(
    `${error.title} [${error.code}] ${error.cause} Next: ${error.action}.`,
    error.action
  );
  if (selected === SETTINGS_ACTION) {
    await vscode.commands.executeCommand('workbench.action.openSettings', `@ext:${EXTENSION_ID}`);
  } else if (selected === GUIDE_ACTION) {
    await openSetupGuide();
  } else if (selected === 'Save File') {
    await vscode.commands.executeCommand('workbench.action.files.save');
  } else if (selected === 'Manage Workspace Trust') {
    await vscode.commands.executeCommand('workbench.trust.manage');
  } else if (selected === 'Run Doctor') {
    await doctor();
  } else if (selected === 'Open Troubleshooting') {
    await openTroubleshooting();
  }
}

function statusWorkspaceFolder() {
  const folders = vscode.workspace.workspaceFolders;
  if (!folders || folders.length === 0) {
    return undefined;
  }
  const editor = vscode.window.activeTextEditor;
  const active = editor && editor.document.uri.scheme === 'file'
    ? vscode.workspace.getWorkspaceFolder(editor.document.uri)
    : undefined;
  return active || folders[0];
}

async function refreshOnboardingStatus() {
  if (!statusItem) {
    return;
  }
  const folder = statusWorkspaceFolder();
  let execution;
  try {
    execution = currentExecutionEnvironment();
  } catch {
    statusItem.hide();
    await setReadinessContexts(false, false);
    return;
  }
  if (!vscode.workspace.isTrusted || !folder || folder.uri.scheme !== 'file') {
    statusItem.hide();
    await setReadinessContexts(false, false);
    return;
  }
  const result = inspectOnboarding({
    settings: settingsFor(folder.uri),
    workspaceRoot: folder.uri.fsPath,
    hostPath: process.env.PATH,
    hostCwd: process.cwd(),
    environment: process.env
  });
  await setReadinessContexts(result.binaryReady, result.configReady);
  statusItem.name = 'Maestro onboarding status';
  if (result.state === ONBOARDING_STATES.BINARY_MISSING) {
    statusItem.text = '$(warning) Maestro: Binary missing';
    statusItem.tooltip = `No executable Maestro CLI was found in ${execution.target}. Open the binaryPath setting.`;
    statusItem.command = {
      command: 'workbench.action.openSettings',
      title: 'Set Maestro Binary Path',
      arguments: [`@ext:${EXTENSION_ID} maestro.binaryPath`]
    };
  } else if (result.state === ONBOARDING_STATES.CONFIG_MISSING) {
    statusItem.text = '$(warning) Maestro: Config missing';
    statusItem.tooltip = `No readable Maestro configuration was found in ${execution.target}. Open the setup guide.`;
    statusItem.command = 'maestro.openSetupGuide';
  } else {
    statusItem.text = '$(check) Maestro: Ready';
    statusItem.tooltip = `Maestro CLI and configuration are available in ${execution.target}. Run Doctor.`;
    statusItem.command = 'maestro.doctor';
  }
  statusItem.show();
}

async function setReadinessContexts(binaryReady, configReady) {
  await Promise.all([
    vscode.commands.executeCommand('setContext', 'maestro.binaryReady', binaryReady),
    vscode.commands.executeCommand('setContext', 'maestro.configReady', configReady)
  ]);
}

function guarded(handler) {
  return async (...args) => {
    try {
      await handler(...args);
    } catch (error) {
      void presentError(normalizeError(error));
    }
  };
}

function activate(context) {
  outputChannel = vscode.window.createOutputChannel('Maestro');
  statusItem = vscode.window.createStatusBarItem('maestro.onboarding', vscode.StatusBarAlignment.Left, 50);
  context.subscriptions.push(
    outputChannel,
    statusItem,
    vscode.tasks.onDidEndTaskProcess(onTaskEnded),
    vscode.workspace.onDidChangeConfiguration(event => {
      if (event.affectsConfiguration('maestro')) {
        void refreshOnboardingStatus();
      }
    }),
    vscode.workspace.onDidChangeWorkspaceFolders(() => void refreshOnboardingStatus()),
    vscode.window.onDidChangeActiveTextEditor(() => void refreshOnboardingStatus())
  );
  const registrations = [
    ['maestro.askActiveFile', askActiveFile],
    ['maestro.replaceSelection', replaceSelection],
    ['maestro.doctor', doctor],
    ['maestro.version', version],
    ['maestro.openSetupGuide', openSetupGuide],
    ['maestro.openChat', openNativeChat],
    ['maestro.openTroubleshooting', openTroubleshooting]
  ];
  for (const [name, handler] of registrations) {
    context.subscriptions.push(vscode.commands.registerCommand(name, guarded(handler)));
  }
  const participant = vscode.chat.createChatParticipant(
    CHAT_PARTICIPANT_ID,
    (request, _chatContext, stream, token) => handleNativeChat(request, stream, token)
  );
  participant.iconPath = vscode.Uri.joinPath(context.extensionUri, 'media', 'icon.png');
  context.subscriptions.push(participant);
  void Promise.all([
    vscode.commands.executeCommand('setContext', 'maestro.doctorPassed', false),
    vscode.commands.executeCommand('setContext', 'maestro.chatPassed', false),
    vscode.commands.executeCommand('setContext', 'maestro.mutationReviewed', false),
    refreshOnboardingStatus()
  ]);
  if (process.env.MAESTRO_VSCODE_TEST_MODE === '1') {
    return Object.freeze({
      executionEnvironment: currentExecutionEnvironment(),
      handleNativeChat
    });
  }
}

function deactivate() {}

module.exports = { activate, deactivate };
