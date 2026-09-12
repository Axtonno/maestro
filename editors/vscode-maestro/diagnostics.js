'use strict';

const FIELD_ORDER = Object.freeze([
  'command',
  'status',
  'duration_ms',
  'exit_code',
  'binary_origin',
  'logical_path',
  'error_code'
]);
const SAFE_TOKEN = /^[a-z0-9_.:-]+$/;

function formatEvent(eventCode, fields = {}, now = new Date()) {
  if (!SAFE_TOKEN.test(eventCode)) {
    throw new Error('unsafe diagnostic event code');
  }
  const unexpected = Object.keys(fields).filter(key => !FIELD_ORDER.includes(key));
  if (unexpected.length > 0) {
    throw new Error(`diagnostic field is not allowlisted: ${unexpected[0]}`);
  }
  const entries = [];
  for (const key of FIELD_ORDER) {
    if (fields[key] === undefined) {
      continue;
    }
    const value = formatValue(key, String(fields[key]));
    if (value === undefined) {
      throw new Error(`unsafe diagnostic value: ${key}`);
    }
    entries.push(`${key}=${value}`);
  }
  return `${now.toISOString()} event=${eventCode}${entries.length ? ` ${entries.join(' ')}` : ''}`;
}

function formatValue(key, value) {
  if (key === 'logical_path') {
    const segments = value.split('/');
    if (value.length > 512 || value.startsWith('/') || segments.some(segment => segment === '..') || /[\u0000-\u001f\u007f]/.test(value)) {
      return undefined;
    }
    return encodeURIComponent(value).replaceAll('%2F', '/');
  }
  if (key === 'duration_ms' || key === 'exit_code') {
    return /^-?[0-9]+$/.test(value) ? value : undefined;
  }
  return value.length <= 80 && SAFE_TOKEN.test(value) ? value : undefined;
}

module.exports = { formatEvent };
