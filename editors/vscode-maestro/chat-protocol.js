'use strict';

const { PreviewError } = require('./errors');

const EXACT_VALUE = /^[^\u0000\r\n]{1,512}$/;

function parseProfileIdentity(encoded) {
  let value;
  try {
    value = JSON.parse(encoded);
  } catch {
    throw new PreviewError('cli_incompatible');
  }
  if (!value || value.schema_version !== 1 ||
      !valid(value.profile) || !valid(value.provider) ||
      !valid(value.chat_model) || !valid(value.mutation_model)) {
    throw new PreviewError('cli_incompatible');
  }
  return Object.freeze({
    profile: value.profile,
    provider: value.provider,
    chatModel: value.chat_model,
    mutationModel: value.mutation_model
  });
}

function parseChatEnvelope(encoded) {
  const marker = '\nresult\n';
  const offset = encoded.indexOf(marker);
  if (offset < 0) {
    throw new PreviewError('cli_incompatible');
  }
  const fields = Object.create(null);
  for (const line of encoded.slice(0, offset).split('\n')) {
    const separator = line.indexOf('\t');
    if (separator <= 0) {
      throw new PreviewError('cli_incompatible');
    }
    fields[line.slice(0, separator)] = line.slice(separator + 1);
  }
  const content = encoded.slice(offset + marker.length).replace(/\n$/, '');
  if (fields.mode !== 'chat' || fields.terminal !== 'completed' || !valid(fields.model) || content.trim() === '') {
    throw new PreviewError('cli_incompatible');
  }
  return Object.freeze({ model: fields.model, content, finishReason: fields.finish_reason || 'unknown' });
}

function classifyCliFailure(stderr) {
  const text = typeof stderr === 'string' ? stderr : '';
  if (/unknown command "profile"/.test(text)) {
    return new PreviewError('cli_incompatible');
  }
  if (/configuration\tkind=read_failed/.test(text)) {
    return new PreviewError('config_not_found');
  }
  if (/provider_unavailable/.test(text)) {
    return new PreviewError('provider_unavailable');
  }
  if (/model_identity_mismatch/.test(text)) {
    return new PreviewError('chat_model_missing');
  }
  return new PreviewError('cli_exit_nonzero');
}

function classifyDoctorOutput(stdout, stderr) {
  const rows = new Map();
  for (const line of String(stdout || '').split(/\r?\n/)) {
    const [status, name, detail] = line.split('\t');
    if (status && name && detail) {
      rows.set(name, { status, detail });
    }
  }
  const failed = name => rows.get(name) && rows.get(name).status === 'fail';
  if (failed('mutation_provider') || /provider_unavailable/.test(String(stderr || ''))) {
    return 'provider_unavailable';
  }
  if (failed('mutation_direct_chat_model') || failed('chat_model')) {
    return 'chat_model_missing';
  }
  if (failed('mutation_controlled_mutation_model')) {
    return 'mutation_model_missing';
  }
  return undefined;
}

function valid(value) {
  return typeof value === 'string' && EXACT_VALUE.test(value);
}

module.exports = { classifyCliFailure, classifyDoctorOutput, parseChatEnvelope, parseProfileIdentity };
