#!/usr/bin/env bash
# Cut a new SDK release.
#
#   ./scripts/release.sh 0.2.0      # or: ./scripts/release.sh v0.2.0
#
# Bumps the Version constant, runs the full check suite, then commits, tags, and
# pushes — the pushed vX.Y.Z tag is what publishes the module to the Go proxy.
# A GitHub release with auto-generated notes is created when the `gh` CLI is
# available.
#
# This does NOT regenerate models from the contract; if the API changed, run
# ./scripts/regenerate.sh first and commit that separately.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

die() { echo "error: $*" >&2; exit 1; }

# ---- 1. Parse + validate the version ------------------------------------
[ $# -eq 1 ] || die "usage: $0 <version>   (e.g. 0.2.0 or v0.2.0)"
RAW="$1"
BARE="${RAW#v}"                 # strip an optional leading v
TAG="v${BARE}"                  # canonical git tag

# SemVer: MAJOR.MINOR.PATCH with an optional -prerelease.
if ! printf '%s' "$BARE" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$'; then
  die "invalid semantic version: '$RAW' (want MAJOR.MINOR.PATCH, e.g. 1.2.3)"
fi

# A v2+ release requires the /vN module-path suffix in go.mod (Go module rule).
MAJOR="${BARE%%.*}"
if [ "$MAJOR" -ge 2 ] && ! grep -q "/v${MAJOR}\$" go.mod; then
  die "v${MAJOR} release needs the module path to end in /v${MAJOR} in go.mod (see Go module versioning)"
fi

# ---- 2. Preconditions ---------------------------------------------------
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
[ "$BRANCH" = "master" ] || die "must release from 'master' (on '$BRANCH')"
git diff-index --quiet HEAD -- || die "working tree is dirty; commit or stash first"
git fetch --tags --quiet origin || true
if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null || \
   git ls-remote --exit-code --tags origin "${TAG}" >/dev/null 2>&1; then
  die "tag ${TAG} already exists (releases are immutable — bump to a new version)"
fi

echo "==> Releasing ${TAG}"

# ---- 3. Bump version.go -------------------------------------------------
grep -q 'const Version = "' version.go || die "could not find Version constant in version.go"
UPDATED="$(sed "s/const Version = \".*\"/const Version = \"${BARE}\"/" version.go)"
printf '%s\n' "$UPDATED" > version.go

# Restore version.go if anything below fails before we commit.
restore() { git checkout -- version.go 2>/dev/null || true; }
trap restore ERR

# ---- 4. Verify ----------------------------------------------------------
echo "==> go build / vet / test"
gofmt -l . | grep -v '^models/' && die "gofmt: files need formatting" || true
go build ./...
go vet ./...
go test ./...

# ---- 5. Commit, tag, push ----------------------------------------------
trap - ERR
echo "==> commit + tag + push"
git add version.go
git commit -m "chore: release ${TAG}"
git tag -a "${TAG}" -m "${TAG}"
git push origin "${BRANCH}"
git push origin "${TAG}"

# ---- 6. GitHub release (best effort) -----------------------------------
if command -v gh >/dev/null 2>&1; then
  gh release create "${TAG}" --title "${TAG}" --generate-notes \
    && echo "==> GitHub release ${TAG} created"
else
  echo "==> gh not found; skipped GitHub release (tag is pushed, module is published)"
fi

echo "==> Done. Verify with:"
echo "    GOPROXY=proxy.golang.org go list -m github.com/crawlsnap/crawlsnap-go@${TAG}"
