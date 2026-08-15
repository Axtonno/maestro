#!/usr/bin/env bash

set -euo pipefail

usage() {
    printf 'usage: %s --version vX.Y.Z[-prerelease] [--status packaging-candidate|release-candidate|release] [--output directory]\n' "$0"
}

version=""
status="packaging-candidate"
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
mkdir -p "$root/configs" "$root/docs" "$root/fixtures/laravel-v1"

ldflags="-s -w -buildid= -X github.com/antonio-cafeo/maestro/internal/buildinfo.Version=${version} -X github.com/antonio-cafeo/maestro/internal/buildinfo.Commit=${commit}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=local GOENV=off GOFLAGS='' \
    GOCACHE="$working/go-cache" SOURCE_DATE_EPOCH="$source_date_epoch" \
    go build -mod=readonly -trimpath -buildvcs=false -ldflags "$ldflags" \
    -o "$root/maestro" ./cmd/maestro

cp LICENSE NOTICE THIRD_PARTY_LICENSES.txt README.md CHANGELOG.md SECURITY.md "$root/"
cp docs/installation.md docs/configuration.md docs/cli.md \
    docs/operational-experience.md docs/packaging-candidate.md \
    docs/quick-start.md docs/reference-agent-laravel.md docs/security-model.md \
    docs/compatibility.md docs/troubleshooting.md docs/known-issues.md \
    docs/v0.1.0-api-compatibility.md docs/laravel-plugin.md "$root/docs/"
mkdir -p "$root/docs/releases"
cp "$release_notes" "$root/docs/releases/"
cp configs/maestro.example.yaml "$root/configs/"
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
status=${status}
EOF

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
