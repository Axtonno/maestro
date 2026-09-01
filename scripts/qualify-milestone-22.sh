#!/usr/bin/env bash

set -euo pipefail

if (($# != 1)); then
    printf 'usage: %s /path/to/extracted-candidate\n' "$0" >&2
    exit 2
fi

root="$(cd "$1" && pwd -P)"
binary="$root/maestro"
config="$root/configs/maestro.chat.example.yaml"
fixture="$root/fixtures/laravel-v1"
working="$(mktemp -d "${TMPDIR:-/tmp}/maestro-m22-live.XXXXXXXX")"
cleanup() { rm -rf -- "$working"; }
trap cleanup EXIT

"$binary" version --diagnostic >"$working/version.out"
"$binary" doctor --mode chat --config "$config" >"$working/doctor.out"

before="$(find "$fixture" -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum)"
"$binary" chat --config "$config" \
    "Quali endpoint dichiara questo progetto?" \
    >"$working/nofile.out" 2>"$working/nofile.err"
"$binary" chat --config "$config" --file routes/api.php \
    "Quali endpoint, controller e action sono dichiarati?" \
    >"$working/complete.out" 2>"$working/complete.err"
"$binary" chat --stream --config "$config" --file routes/api.php \
    "Quali endpoint, controller e action sono dichiarati?" \
    >"$working/stream.out" 2>"$working/stream.err"

set +e
traversal="$($binary chat --config "$config" --file ../outside.php Domanda 2>&1)"
traversal_code=$?
set -e
after="$(find "$fixture" -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum)"

[[ "$before" == "$after" ]]
[[ $traversal_code -eq 2 && "$traversal" == 'chat failed: file_not_allowed' ]]
grep -Fq $'pass\tgeneration\tgeneration_controls_available' "$working/doctor.out"
for output in "$working/complete.out" "$working/stream.out"; do
    grep -Fq 'POST' "$output"
    grep -Fq '/orders' "$output"
    grep -Fq 'OrderController' "$output"
    grep -Fq 'store' "$output"
done
grep -Fq $'num_predict_requested\t512' "$working/complete.out"
grep -Fq $'residency_requested\t5m0s' "$working/complete.out"
grep -Fq $'truncated\tfalse' "$working/complete.out"
grep -Fq $'finish_reason\tstop' "$working/complete.out"
grep -Fq $'truncated\tfalse' "$working/stream.out"
grep -Fq $'finish_reason\tstop' "$working/stream.out"

if grep -Ev '^progress\tstate=generating elapsed_ms=[0-9]+$' \
    "$working/complete.err" "$working/stream.err" | grep -q .; then
    printf 'live stderr contains a non-allowlisted line\n' >&2
    exit 1
fi

printf 'milestone-22 live gate passed\n'
printf 'fixture_digest=%s\n' "$after"
printf 'complete_heartbeats=%s\n' "$(wc -l <"$working/complete.err")"
printf 'stream_heartbeats=%s\n' "$(wc -l <"$working/stream.err")"
tail -n 12 "$working/complete.out"
tail -n 12 "$working/stream.out"
