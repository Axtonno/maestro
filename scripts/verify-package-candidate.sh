#!/usr/bin/env bash

set -euo pipefail

version="v0.1.0-pc.1"
if (($# == 2)) && [[ "$1" == "--version" ]]; then
    version="$2"
elif (($# != 0)); then
    printf 'usage: %s [--version vX.Y.Z-prerelease]\n' "$0" >&2
    exit 2
fi

repository="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
commit="$(git -C "$repository" rev-parse --verify HEAD)"
artifact="maestro-${version}-linux-amd64"
archive="${artifact}.tar.gz"
checksum="${archive}.sha256"
working="$(mktemp -d "${TMPDIR:-/tmp}/maestro-package-gate.XXXXXXXX")"
cleanup() {
    rm -rf -- "$working"
}
trap cleanup EXIT

mkdir -p "$working/first" "$working/second" "$working/extracted" \
    "$working/install/bin" "$working/empty"
"$repository/scripts/package-candidate.sh" --version "$version" --output "$working/first" >/dev/null
"$repository/scripts/package-candidate.sh" --version "$version" --output "$working/second" >/dev/null

cmp "$working/first/$archive" "$working/second/$archive"
cmp "$working/first/$checksum" "$working/second/$checksum"
(
    cd "$working/first"
    sha256sum -c "$checksum"
)

tar -tzf "$working/first/$archive" >"$working/archive.list"
if grep -Eq '(^/|(^|/)\.\.(/|$))' "$working/archive.list"; then
    printf 'archive contains an unsafe path\n' >&2
    exit 1
fi
tar -xzf "$working/first/$archive" -C "$working/extracted"
root="$working/extracted/$artifact"

for required in maestro LICENSE NOTICE THIRD_PARTY_LICENSES.txt README.md ARTIFACT-MANIFEST.txt \
    docs/installation.md docs/configuration.md docs/cli.md \
    docs/operational-experience.md docs/packaging-candidate.md \
    configs/maestro.example.yaml fixtures/laravel-v1/dataset.json \
    fixtures/laravel-v1/artisan fixtures/laravel-v1/composer.json; do
    [[ -e "$root/$required" ]] || {
        printf 'archive is missing %s\n' "$required" >&2
        exit 1
    }
done
[[ -x "$root/maestro" && -x "$root/fixtures/laravel-v1/artisan" ]]
cmp "$repository/LICENSE" "$root/LICENSE"
cmp "$repository/NOTICE" "$root/NOTICE"
cmp "$repository/THIRD_PARTY_LICENSES.txt" "$root/THIRD_PARTY_LICENSES.txt"
grep -Fxq "artifact=${artifact}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "version=${version}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "commit=${commit}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "version=${version}" "$root/docs/installation.md"
grep -Fq 'artifact="maestro-${version}-linux-amd64"' "$root/docs/installation.md"
if grep -R -Fq '@MAESTRO_VERSION@' "$root"; then
    printf 'archive contains an unresolved version token\n' >&2
    exit 1
fi
grep -Fq '"id": "maestro-laravel-mini"' "$root/fixtures/laravel-v1/dataset.json"
grep -Fq '"version": "1.0.0"' "$root/fixtures/laravel-v1/dataset.json"

version_output="$($root/maestro version)"
grep -Fxq "maestro ${version}" <<<"$version_output"
grep -Fxq "commit ${commit}" <<<"$version_output"
"$root/maestro" --help | grep -Fq 'usage: maestro <command>'

doctor_config="$root/configs/doctor-test.yaml"
cp "$root/configs/maestro.example.yaml" "$doctor_config"
sed -i 's#http://127.0.0.1:11434#http://127.0.0.1:1#' "$doctor_config"
set +e
doctor_output="$($root/maestro doctor --config "$doctor_config" 2>&1)"
doctor_status=$?
set -e
[[ $doctor_status -eq 1 ]]
grep -Fq $'pass\tconfig\tschema_v1_valid' <<<"$doctor_output"
grep -Fq $'pass\tworkspace\troot_available' <<<"$doctor_output"
grep -Fq $'pass\tlaravel\tworkspace_detected' <<<"$doctor_output"
grep -Fq $'fail\tprovider\tinstance_probe_failed' <<<"$doctor_output"

install -m 0755 "$root/maestro" "$working/install/bin/maestro"
(
    cd "$working/empty"
    "$working/install/bin/maestro" version | grep -Fxq "maestro ${version}"
    "$working/install/bin/maestro" --help | grep -Fq 'usage: maestro <command>'
)

if find "$root/fixtures" -type l -o -name vendor -o -name node_modules -o \
    -name .git -o -name .env | grep -q .; then
    printf 'archive fixture contains a forbidden entry\n' >&2
    exit 1
fi
if grep -aR -Fq "$repository" "$root"; then
    printf 'archive exposes the build workspace path\n' >&2
    exit 1
fi
if grep -aERq -- '-----BEGIN [A-Z ]*PRIVATE KEY-----|AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9]{20,}|sk-[A-Za-z0-9]{20,}' "$root"; then
    printf 'archive contains credential-shaped data\n' >&2
    exit 1
fi

printf 'packaging candidate verified: %s commit=%s\n' "$archive" "$commit"
sha256sum "$working/first/$archive"
