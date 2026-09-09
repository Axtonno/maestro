#!/usr/bin/env bash

set -euo pipefail

usage() {
    printf 'usage: %s --version vX.Y.Z[-prerelease] [--status packaging-candidate|release-candidate|release] [--profile release|cpu-qualification|mutation-productization] [--output directory]\n' "$0"
}

version=""
status="packaging-candidate"
profile_kind="release"
output="dist"
while (($# > 0)); do
    case "$1" in
        --version)
            (($# >= 2)) || { usage >&2; exit 2; }
            version="$2"
            shift 2
            ;;
        --status)
            (($# >= 2)) || { usage >&2; exit 2; }
            status="$2"
            shift 2
            ;;
        --profile)
            (($# >= 2)) || { usage >&2; exit 2; }
            profile_kind="$2"
            shift 2
            ;;
        --output)
            (($# >= 2)) || { usage >&2; exit 2; }
            output="$2"
            shift 2
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            printf 'unknown argument: %s\n' "$1" >&2
            usage >&2
            exit 2
            ;;
    esac
done

if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
    printf 'version must match vX.Y.Z or vX.Y.Z-prerelease\n' >&2
    exit 2
fi
if [[ "$status" != "packaging-candidate" && "$status" != "release-candidate" && "$status" != "release" ]]; then
    printf 'status must be packaging-candidate, release-candidate or release\n' >&2
    exit 2
fi
if [[ "$status" == "release" && "$version" == *-* ]]; then
    printf 'release status requires a final vX.Y.Z version\n' >&2
    exit 2
fi
if [[ "$profile_kind" != "release" && "$profile_kind" != "cpu-qualification" && "$profile_kind" != "mutation-productization" ]]; then
    printf 'profile must be release, cpu-qualification or mutation-productization\n' >&2
	exit 2
fi

profile_source="configs/maestro.chat.example.yaml"
profile_target="configs/maestro.chat.example.yaml"
chat_model="qwen3.5:9b"
chat_model_digest="6488c96fa5faab64bb65cbd30d4289e20e6130ef535a93ef9a49f42eda893ea7"
chat_num_predict="1024"
chat_residency="5m"
chat_streaming="true"
mutation_model=""
mutation_model_digest=""
mutation_prompt=""
mutation_prompt_sha256=""
mutation_schema=""
mutation_schema_sha256=""
if [[ "$profile_kind" == "cpu-qualification" ]]; then
	profile_source="configs/maestro.milestone-21-candidate.yaml"
	chat_model="qwen2.5-coder:7b"
	chat_model_digest="dae161e27b0e90dd1856c8bb3209201fd6736d8eb66298e75ed87571486f4364"
	chat_num_predict="512"
	chat_residency="5m"
elif [[ "$profile_kind" == "mutation-productization" ]]; then
	profile_source="configs/maestro.v0.5.0-candidate.yaml"
	profile_target="configs/maestro.v0.5.0-candidate.yaml"
	chat_streaming="false"
	mutation_model="qwen2.5-coder:14b"
	mutation_model_digest="9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849"
	mutation_prompt="mutation-host-bound-model-selection-v1"
	mutation_prompt_sha256="594659d52ec6142a5ef79c36dc0db4899e7ef1bb3f99d05017410f68bc1ba732"
	mutation_schema="host-bound-mutation-decision-v1"
	mutation_schema_sha256="bc3432a8f19867eec8e153adaa4434b688974cf34d24b6bd770e887e0dd7557d"
fi

repository="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$repository"

for command in git go tar gzip sha256sum mktemp; do
    command -v "$command" >/dev/null || {
        printf 'required command is unavailable: %s\n' "$command" >&2
        exit 1
    }
done
tar_version="$(tar --version)"
[[ "$tar_version" == *"GNU tar"* ]] || {
    printf 'GNU tar is required for normalized archives\n' >&2
    exit 1
}

if ! git diff --quiet || ! git diff --cached --quiet ||
    [[ -n "$(git ls-files --others --exclude-standard)" ]]; then
    printf 'packaging requires a clean worktree with no untracked inputs\n' >&2
    exit 1
fi

commit="$(git rev-parse --verify HEAD)"
source_date_epoch="$(git show -s --format=%ct HEAD)"
go_version="$(go env GOVERSION)"
artifact="maestro-${version}-linux-amd64"
archive="${artifact}.tar.gz"
checksum="${archive}.sha256"
release_version="${version#v}"
release_notes="docs/releases/v${release_version%%-*}.md"

if [[ ! -f "$release_notes" ]]; then
    printf 'release notes are unavailable: %s\n' "$release_notes" >&2
    exit 1
fi

if [[ "$output" != /* ]]; then
    output="$repository/$output"
fi
mkdir -p "$output"
if [[ -e "$output/$archive" || -e "$output/$checksum" ]]; then
    printf 'refusing to overwrite existing artifact in %s\n' "$output" >&2
    exit 1
fi

working="$(mktemp -d "${TMPDIR:-/tmp}/maestro-package.XXXXXXXX")"
cleanup() {
    rm -rf -- "$working"
}
trap cleanup EXIT

root="$working/$artifact"
mkdir -p "$root/configs" "$root/docs/reports" "$root/fixtures/laravel-v1"

ldflags="-s -w -buildid= -X github.com/antonio-cafeo/maestro/internal/buildinfo.Version=${version} -X github.com/antonio-cafeo/maestro/internal/buildinfo.Commit=${commit} -X github.com/antonio-cafeo/maestro/internal/buildinfo.Status=${status}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=local GOENV=off GOFLAGS='' \
    GOCACHE="$working/go-cache" SOURCE_DATE_EPOCH="$source_date_epoch" \
    go build -mod=readonly -trimpath -buildvcs=false -ldflags "$ldflags" \
    -o "$root/maestro" ./cmd/maestro

cp LICENSE NOTICE THIRD_PARTY_LICENSES.txt README.md CHANGELOG.md SECURITY.md "$root/"
cp docs/installation.md docs/configuration.md docs/cli.md \
	docs/packaging-candidate.md docs/quick-start.md docs/security-model.md \
	docs/compatibility.md docs/troubleshooting.md docs/known-issues.md \
	docs/install-and-try.md docs/controlled-mutation-support.md \
	docs/current-capabilities.md docs/validation.md \
	"$root/docs/"
cp docs/milestone-38-field-adoption-freeze.yaml "$root/docs/"
cp docs/v0.5.0-public-baseline-freeze.yaml "$root/docs/"
cp docs/reports/milestone-38-final.md \
	docs/reports/milestone-38-environment.yaml \
	docs/reports/milestone-38-live-runs.json \
	"$root/docs/reports/"
if [[ "$profile_kind" == "mutation-productization" ]]; then
	mkdir -p "$root/docs/prompts" "$root/docs/schemas"
	cp docs/prompts/mutation-host-bound-model-selection-v1.txt "$root/docs/prompts/"
	cp docs/schemas/host-bound-mutation-decision-v1.schema.json "$root/docs/schemas/"
fi
mkdir -p "$root/docs/releases"
cp "$release_notes" "$root/docs/releases/"
cp "$profile_source" "$root/$profile_target"
if [[ "$profile_kind" == "cpu-qualification" ]]; then
	sed -i 's#root: ../internal/benchmark/developer/testdata/laravel-v1#root: ../fixtures/laravel-v1#' \
		"$root/configs/maestro.chat.example.yaml"
fi
cp -R internal/benchmark/developer/testdata/laravel-v1/. "$root/fixtures/laravel-v1/"
sed -i "s/@MAESTRO_VERSION@/${version}/g" "$root/docs/installation.md"
sed -i "s/@MAESTRO_STATUS@/${status}/g" "$root/docs/installation.md"
sed -i "s/@MAESTRO_VERSION@/${version}/g" "$root/docs/quick-start.md"

if find "$root/fixtures" -type l -o -name vendor -o -name node_modules -o \
    -name .git -o -name .env | grep -q .; then
    printf 'release fixture contains a forbidden entry\n' >&2
    exit 1
fi

cat >"$root/ARTIFACT-MANIFEST.txt" <<EOF
artifact=${artifact}
version=${version}
commit=${commit}
platform=linux/amd64
go=${go_version}
license=Apache-2.0
fixture=maestro-laravel-mini@1.0.0
profile=${profile_target}
profile_kind=${profile_kind}
chat_model=${chat_model}
chat_model_digest=${chat_model_digest}
chat_num_ctx=4096
chat_num_predict=${chat_num_predict}
chat_thinking=false
chat_temperature=0
chat_residency=${chat_residency}
chat_streaming=${chat_streaming}
status=${status}
EOF
if [[ "$profile_kind" == "mutation-productization" ]]; then
	cat >>"$root/ARTIFACT-MANIFEST.txt" <<EOF
controlled_mutation_enabled=true
controlled_mutation_model=${mutation_model}
controlled_mutation_model_digest=${mutation_model_digest}
controlled_mutation_num_ctx=4096
controlled_mutation_num_predict=1024
controlled_mutation_thinking=false
controlled_mutation_temperature=0
controlled_mutation_residency=5m
controlled_mutation_prompt=${mutation_prompt}
controlled_mutation_prompt_sha256=${mutation_prompt_sha256}
controlled_mutation_schema=${mutation_schema}
controlled_mutation_schema_sha256=${mutation_schema_sha256}
EOF
fi

find "$root" -type d -exec chmod 0755 {} +
find "$root" -type f -exec chmod 0644 {} +
chmod 0755 "$root/maestro" "$root/fixtures/laravel-v1/artisan"

LC_ALL=C tar --sort=name --format=ustar --mtime="@${source_date_epoch}" \
    --owner=0 --group=0 --numeric-owner -C "$working" -cf - "$artifact" |
    gzip -n -9 >"$working/$archive"
(
    cd "$working"
    sha256sum "$archive" >"$checksum"
)

mv "$working/$archive" "$output/$archive"
mv "$working/$checksum" "$output/$checksum"
printf '%s\n%s\n' "$output/$archive" "$output/$checksum"
