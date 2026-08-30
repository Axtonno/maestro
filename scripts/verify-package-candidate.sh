#!/usr/bin/env bash

set -euo pipefail

version="v0.2.0-pc.1"
status="packaging-candidate"
profile_kind="release"
while (($# > 0)); do
    case "$1" in
        --version)
            (($# >= 2)) || { printf 'missing --version value\n' >&2; exit 2; }
            version="$2"
            shift 2
            ;;
        --status)
            (($# >= 2)) || { printf 'missing --status value\n' >&2; exit 2; }
            status="$2"
            shift 2
            ;;
        --profile)
            (($# >= 2)) || { printf 'missing --profile value\n' >&2; exit 2; }
            profile_kind="$2"
            shift 2
            ;;
        *)
            printf 'usage: %s [--version vX.Y.Z[-prerelease]] [--status packaging-candidate|release-candidate|release] [--profile release|cpu-qualification]\n' "$0" >&2
            exit 2
            ;;
    esac
done
if [[ "$status" != "packaging-candidate" && "$status" != "release-candidate" && "$status" != "release" ]]; then
    printf 'status must be packaging-candidate, release-candidate or release\n' >&2
    exit 2
fi
if [[ "$status" == "release" && "$version" == *-* ]]; then
    printf 'release status requires a final vX.Y.Z version\n' >&2
    exit 2
fi
if [[ "$profile_kind" != "release" && "$profile_kind" != "cpu-qualification" ]]; then
    printf 'profile must be release or cpu-qualification\n' >&2
    exit 2
fi

schema_version="2"
chat_model="qwen3.5:9b"
chat_model_digest="6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7"
chat_num_predict="provider-default"
chat_residency="provider-default"
if [[ "$profile_kind" == "cpu-qualification" ]]; then
    schema_version="3"
    chat_model="qwen2.5-coder:7b"
    chat_model_digest="dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364"
    chat_num_predict="512"
    chat_residency="5m"
fi

repository="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
commit="$(git -C "$repository" rev-parse --verify HEAD)"
artifact="maestro-${version}-linux-amd64"
archive="${artifact}.tar.gz"
checksum="${archive}.sha256"
release_version="${version#v}"
release_notes="docs/releases/v${release_version%%-*}.md"
working="$(mktemp -d "${TMPDIR:-/tmp}/maestro-package-gate.XXXXXXXX")"
cleanup() {
    rm -rf -- "$working"
}
trap cleanup EXIT

mkdir -p "$working/first" "$working/second" "$working/extracted" \
    "$working/install/bin" "$working/empty"
"$repository/scripts/package-candidate.sh" --version "$version" --status "$status" --profile "$profile_kind" --output "$working/first" >/dev/null
"$repository/scripts/package-candidate.sh" --version "$version" --status "$status" --profile "$profile_kind" --output "$working/second" >/dev/null

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

for required in maestro LICENSE NOTICE THIRD_PARTY_LICENSES.txt README.md CHANGELOG.md SECURITY.md ARTIFACT-MANIFEST.txt \
    docs/installation.md docs/configuration.md docs/cli.md \
    docs/packaging-candidate.md docs/quick-start.md docs/security-model.md \
    docs/compatibility.md docs/troubleshooting.md docs/known-issues.md \
    "$release_notes" \
    configs/maestro.chat.example.yaml fixtures/laravel-v1/dataset.json \
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
cmp "$repository/CHANGELOG.md" "$root/CHANGELOG.md"
cmp "$repository/SECURITY.md" "$root/SECURITY.md"
grep -Fxq "artifact=${artifact}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "version=${version}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "commit=${commit}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "status=${status}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq 'profile=configs/maestro.chat.example.yaml' "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "profile_kind=${profile_kind}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "chat_model=${chat_model}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "chat_model_digest=${chat_model_digest}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq 'chat_num_ctx=4096' "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "chat_num_predict=${chat_num_predict}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq 'chat_thinking=false' "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq 'chat_temperature=0' "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "chat_residency=${chat_residency}" "$root/ARTIFACT-MANIFEST.txt"
grep -Fxq "version=${version}" "$root/docs/installation.md"
grep -Fxq "Stato: ${status}" "$root/docs/installation.md"
grep -Fq 'artifact="maestro-${version}-linux-amd64"' "$root/docs/installation.md"
if grep -R -Fq '@MAESTRO_' "$root"; then
    printf 'archive contains an unresolved documentation token\n' >&2
    exit 1
fi
grep -Fq '"id": "maestro-laravel-mini"' "$root/fixtures/laravel-v1/dataset.json"
grep -Fq '"version": "1.0.0"' "$root/fixtures/laravel-v1/dataset.json"
profile="$root/configs/maestro.chat.example.yaml"
grep -Eq "^[[:space:]]*version:[[:space:]]*${schema_version}[[:space:]]*$" "$profile"
grep -Eq "^[[:space:]]*model:[[:space:]]*${chat_model//./\\.}[[:space:]]*$" "$profile"
grep -Eq '^[[:space:]]*streaming:[[:space:]]*true[[:space:]]*$' "$profile"
grep -Eq '^[[:space:]]*num_ctx:[[:space:]]*4096[[:space:]]*$' "$profile"
grep -Eq '^[[:space:]]*thinking:[[:space:]]*"false"[[:space:]]*$' "$profile"
grep -Eq '^[[:space:]]*max_file_bytes:[[:space:]]*1048576[[:space:]]*$' "$profile"
grep -Eq '^[[:space:]]*max_output_bytes:[[:space:]]*1048576[[:space:]]*$' "$profile"
if [[ "$profile_kind" == "cpu-qualification" ]]; then
    grep -Eq '^[[:space:]]*num_predict:[[:space:]]*512[[:space:]]*$' "$profile"
    grep -Eq '^[[:space:]]*residency:[[:space:]]*5m[[:space:]]*$' "$profile"
fi
if grep -Eq 'workspace\.(write|patch)|^[[:space:]]*(agent|limits|context):' "$profile"; then
    printf 'published configuration exposes an unsupported agent or mutation surface\n' >&2
    exit 1
fi
grep -Eq '^[[:space:]]*workspace_mutate:[[:space:]]*deny[[:space:]]*$' \
    "$profile"
for unsupported in configs/maestro.mutating.example.yaml \
    configs/maestro.example.yaml configs/maestro.interaction.example.yaml \
    docs/mutation-qualification.md docs/mutation-benchmark.md \
    docs/reference-agent-laravel.md docs/operational-experience.md; do
    [[ ! -e "$root/$unsupported" ]] || {
        printf 'archive publishes unsupported mutation surface: %s\n' "$unsupported" >&2
        exit 1
    }
done

version_output="$("$root/maestro" version)"
grep -Fxq "maestro ${version}" <<<"$version_output"
grep -Fxq "commit ${commit}" <<<"$version_output"
diagnostic_output="$("$root/maestro" version --diagnostic)"
read -r binary_sha256 _ < <(sha256sum "$root/maestro")
grep -Fxq $'mode\tbinary_identity' <<<"$diagnostic_output"
grep -Fxq $'version\t'"${version}" <<<"$diagnostic_output"
grep -Fxq $'status\t'"${status}" <<<"$diagnostic_output"
grep -Fxq $'commit\t'"${commit}" <<<"$diagnostic_output"
grep -Fxq $'dirty\tfalse' <<<"$diagnostic_output"
grep -Fxq $'executable\t'"\"$root/maestro\"" <<<"$diagnostic_output"
grep -Fxq $'sha256\t'"${binary_sha256}" <<<"$diagnostic_output"
"$root/maestro" --help | grep -Fq 'usage: maestro <command>'

doctor_config="$root/configs/doctor-test.yaml"
cp "$profile" "$doctor_config"
sed -i 's#http://127.0.0.1:11434#http://127.0.0.1:1#' "$doctor_config"
set +e
doctor_output="$($root/maestro doctor --mode chat --config "$doctor_config" 2>&1)"
doctor_status=$?
set -e
[[ $doctor_status -eq 1 ]]
grep -Fq $'pass\tconfig\tschema_v'"${schema_version}"'_chat_valid' <<<"$doctor_output"
grep -Fq $'pass\tworkspace\troot_available' <<<"$doctor_output"
grep -Fq $'pass\tcomposition\tdirect_chat_provider' <<<"$doctor_output"
grep -Fq $'fail\tmodel\trequired_capability_unavailable' <<<"$doctor_output"
grep -Fq $'skip\tgeneration\tmodel_unavailable' <<<"$doctor_output"

set +e
containment_output="$($root/maestro chat --config "$profile" --file ../outside.php Question 2>&1)"
containment_status=$?
set -e
[[ $containment_status -eq 2 ]]
[[ "$containment_output" == 'chat failed: file_not_allowed' ]]

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

printf 'artifact verified: %s commit=%s status=%s profile=%s\n' "$archive" "$commit" "$status" "$profile_kind"
sha256sum "$working/first/$archive"
