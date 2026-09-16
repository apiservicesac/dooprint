#!/usr/bin/env bash
#
# Dooprint release script.
#
# Bumps VERSION, moves the "Unreleased" notes of CHANGELOG.md to the new version, runs the checks,
# commits, tags vX.Y.Z and pushes. The Release workflow on GitHub then builds the binaries and the
# installers and publishes them.
#
# Usage:
#   scripts/release.sh patch|minor|major [--dry-run]
#   scripts/release.sh 1.2.3 [--dry-run]
#   scripts/release.sh --bump-only 1.2.3

set -euo pipefail
cd "$(dirname "$0")/.."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1" >&2; }

show_help() {
    sed -n '3,13p' "$0" | sed 's/^# \{0,1\}//'
}

DRY_RUN=false
BUMP_ONLY=false
TARGET=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true ;;
        --bump-only) BUMP_ONLY=true ;;
        -h|--help) show_help; exit 0 ;;
        *)
            if [[ -n "$TARGET" ]]; then
                log_error "Unknown argument: $1"; show_help; exit 1
            fi
            TARGET="$1"
            ;;
    esac
    shift
done

if [[ -z "$TARGET" ]]; then
    log_error "A release type (patch, minor, major) or a version is required"
    show_help
    exit 1
fi

run() {
    if $DRY_RUN; then
        log_info "[DRY RUN] $*"
    else
        "$@"
    fi
}

current_version() { tr -d '[:space:]' < VERSION; }

next_version() {
    local current="$1" kind="$2" major minor patch
    IFS='.' read -r major minor patch <<< "${current%%-*}"
    case "$kind" in
        patch) echo "$major.$minor.$((patch + 1))" ;;
        minor) echo "$major.$((minor + 1)).0" ;;
        major) echo "$((major + 1)).0.0" ;;
        *) echo "$kind" ;;
    esac
}

validate_version() {
    if [[ ! "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$ ]]; then
        log_error "Invalid version: $1 (expected X.Y.Z or X.Y.Z-suffix)"
        exit 1
    fi
}

update_changelog() {
    local version="$1" previous="$2" date
    date=$(date +%Y-%m-%d)
    if ! grep -q '^## \[Unreleased\]' CHANGELOG.md; then
        log_error "CHANGELOG.md has no '## [Unreleased]' section"
        exit 1
    fi
    if $DRY_RUN; then
        log_info "[DRY RUN] Would move the Unreleased notes to [$version] - $date"
        return
    fi
    awk -v version="$version" -v date="$date" '
        /^## \[Unreleased\]/ { print; print ""; print "## [" version "] - " date; next }
        { print }
    ' CHANGELOG.md > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md
    # Comparison links at the bottom.
    sed -i -E "s#^\[Unreleased\]: (.*)/compare/v[^ ]+\.\.\.HEAD\$#[Unreleased]: \1/compare/v$version...HEAD\n[$version]: \1/compare/v$previous...v$version#" CHANGELOG.md
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    git rev-parse --is-inside-work-tree >/dev/null 2>&1 || { log_error "Not in a git repository"; exit 1; }

    local branch
    branch=$(git branch --show-current)
    if [[ "$branch" != "main" ]]; then
        log_warning "Not on main (on: $branch)"
        read -r -p "Continue anyway? (y/N): " confirm
        [[ "$confirm" =~ ^[yY]$ ]] || exit 1
    fi

    if [[ -n "$(git status --porcelain)" ]]; then
        log_error "Uncommitted changes. Commit or stash them first."
        exit 1
    fi

    for tool in git docker make; do
        command -v "$tool" >/dev/null || { log_error "$tool is not installed"; exit 1; }
    done

    run git fetch --quiet --tags origin
    if [[ "$(git rev-parse HEAD)" != "$(git rev-parse '@{u}' 2>/dev/null || git rev-parse HEAD)" ]]; then
        log_error "The branch is not in sync with its remote. Pull or push first."
        exit 1
    fi
    log_success "Prerequisites check passed"
}

main() {
    local current new
    current=$(current_version)
    new=$(next_version "$current" "$TARGET")
    validate_version "$new"

    if $BUMP_ONLY; then
        echo "$new" > VERSION
        log_success "VERSION set to $new (was $current). Commit it with the CHANGELOG."
        exit 0
    fi

    if [[ "$new" == "$current" ]]; then
        log_error "Version $new is the current version"
        exit 1
    fi
    if git rev-parse -q --verify "refs/tags/v$new" >/dev/null; then
        log_error "Tag v$new already exists"
        exit 1
    fi

    log_info "Releasing dooprint: $current -> $new"
    $DRY_RUN && log_warning "DRY RUN: nothing will be changed"

    check_prerequisites

    if ! $DRY_RUN; then
        echo ""
        log_warning "About to release v$new and push it to origin"
        read -r -p "Continue? (y/N): " confirm
        [[ "$confirm" =~ ^[yY]$ ]] || { log_info "Release cancelled"; exit 0; }
    fi

    log_info "Running checks..."
    run make check test

    log_info "Updating VERSION and CHANGELOG.md..."
    $DRY_RUN || echo "$new" > VERSION
    update_changelog "$new" "$current"

    log_info "Committing and tagging..."
    run git add VERSION CHANGELOG.md
    run git commit -q -m "chore: release v$new"
    run git tag -a "v$new" -m "Release v$new"

    log_info "Pushing..."
    run git push -q origin HEAD
    run git push -q origin "v$new"

    local repo
    repo=$(git remote get-url origin | sed -E 's#.*github.com[:/](.+)\.git$#\1#; s#.*github.com[:/](.+)$#\1#')
    log_success "Release v$new pushed"
    echo ""
    log_info "The Release workflow builds and publishes it:"
    log_info "  https://github.com/$repo/actions"
    log_info "  https://github.com/$repo/releases/tag/v$new"
}

main
