'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const {
  classifyCliFailure,
  classifyDoctorOutput,
  parseChatEnvelope,
  parseProfileIdentity
} = require('../chat-protocol');

test('parses the exact profile identity contract', () => {
  assert.deepEqual(parseProfileIdentity(JSON.stringify({
    schema_version: 1,
    profile: 'recommended',
    provider: 'ollama',
    chat_model: 'qwen3.5:9b',
    mutation_model: 'qwen2.5-coder:14b'
  })), {
    profile: 'recommended',
    provider: 'ollama',
    chatModel: 'qwen3.5:9b',
    mutationModel: 'qwen2.5-coder:14b'
  });
  assert.throws(() => parseProfileIdentity('{"schema_version":2}'), error => error.code === 'cli_incompatible');
});

test('separates the stable chat envelope from untrusted result text', () => {
  assert.deepEqual(parseChatEnvelope([
    'mode\tchat',
    'terminal\tcompleted',
    'model\tqwen3.5:9b',
    'finish_reason\tstop',
    'result',
    '# Answer',
    'Use `code`.',
    ''
  ].join('\n')), {
    model: 'qwen3.5:9b',
    content: '# Answer\nUse `code`.',
    finishReason: 'stop'
  });
  assert.throws(() => parseChatEnvelope('result\nmissing envelope'), error => error.code === 'cli_incompatible');
});

test('maps CLI and Doctor failures to actionable setup states', () => {
  assert.equal(classifyCliFailure('unknown command "profile"').code, 'cli_incompatible');
  assert.equal(classifyCliFailure('profile failed: invalid_request\nconfiguration\tkind=read_failed').code, 'config_not_found');
  assert.equal(classifyCliFailure('profile failed: invalid_request').code, 'cli_exit_nonzero');
  assert.equal(classifyCliFailure('chat failed: provider_unavailable').code, 'provider_unavailable');
  assert.equal(classifyDoctorOutput([
    'pass\tmutation_provider\tollama_available',
    'fail\tmutation_direct_chat_model\tmodel_or_digest_mismatch',
    'pass\tmutation_controlled_mutation_model\tqualified_model_digest'
  ].join('\n')), 'chat_model_missing');
  assert.equal(classifyDoctorOutput([
    'pass\tmutation_provider\tollama_available',
    'pass\tmutation_direct_chat_model\tqualified_model_digest',
    'fail\tmutation_controlled_mutation_model\tmodel_or_digest_mismatch'
  ].join('\n')), 'mutation_model_missing');
  assert.equal(classifyDoctorOutput('fail\tmutation_provider\tprovider_unavailable'), 'provider_unavailable');
});
