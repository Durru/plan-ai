#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Plan-AI — Install Script
# Local-first continuous implementation planning for AI-assisted projects.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Durru/plan-ai/main/scripts/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/Durru/plan-ai/main/scripts/install.sh | bash -s -- go
#   curl -fsSL https://raw.githubusercontent.com/Durru/plan-ai/main/scripts/install.sh | bash -s -- git
# ============================================================================

GITHUB_OWNER="Durru"
GITHUB_REPO="plan-ai"
BINARY_NAME="plan-ai"

setup_colors() {
    if [ -t 1 ] && [ "${TERM:-}" != "dumb" ]; then
        RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
        BLUE='\033[0;34m'; CYAN='\033[0;36m'; BOLD='\033[1m'
        DIM='\033[2m'; NC='\033[0m'
    else
        RED=''; GREEN=''; YELLOW=''; BLUE=''; CYAN=''; BOLD=''; DIM=''; NC=''
    fi
}

info()    { echo -e "${BLUE}[info]${NC}    $*"; }
success() { echo -e "${GREEN}[ok]${NC}      $*"; }
warn()    { echo -e "${YELLOW}[warn]${NC}    $*"; }
error()   { echo -e "${RED}[error]${NC}   $*" >&2; }
fatal()   { error "$@"; exit 1; }
step()    { echo -e "\n${CYAN}${BOLD}==>${NC} ${BOLD}$*${NC}"; }

detect_platform() {
    case "$(uname -s)" in
        Darwin) OS="darwin"; GORELEASER_OS="darwin" ;;
        Linux)  OS="linux";  GORELEASER_OS="linux" ;;
        *)      fatal "Unsupported OS. Only macOS and Linux are supported." ;;
    esac
    case "$(uname -m)" in
        x86_64|amd64)   ARCH="amd64" ;;
        arm64|aarch64)  ARCH="arm64" ;;
        *)              fatal "Unsupported architecture. Only amd64 and arm64 are supported." ;;
    esac
    success "Platform: ${OS}/${ARCH}"
}

get_latest_version() {
    info "Fetching latest release..."
    local response http_code
    response="$(curl -sL -w "\n%{http_code}" "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest")"
    http_code="$(echo "$response" | tail -n1)"
    if [ "$http_code" = "404" ]; then
        warn "No GitHub releases yet — falling back to go install"
        install_go
        exit 0
    fi
    if [ "$http_code" != "200" ]; then
        fatal "GitHub API returned HTTP $http_code. Try: curl ... | bash -s -- go"
    fi
    LATEST_VERSION="$(echo "$response" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)"
    if [ -z "$LATEST_VERSION" ]; then
        fatal "Could not determine latest version from GitHub API"
    fi
    VERSION_NUMBER="${LATEST_VERSION#v}"
    success "Latest version: ${LATEST_VERSION}"
}

install_binary() {
    step "Installing pre-built binary"
    get_latest_version

    local archive="plan-ai_${VERSION_NUMBER}_${GORELEASER_OS}_${ARCH}.tar.gz"
    local url="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/${archive}"
    local checksums_url="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/checksums.txt"

    local tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    info "Downloading ${archive}..."
    curl -sfL -o "${tmpdir}/${archive}" "$url" || fatal "Failed to download ${url}"

    local file_size="$(wc -c < "${tmpdir}/${archive}" | tr -d '[:space:]')"
    [ "$file_size" -lt 1000 ] && fatal "Downloaded file is too small (${file_size} bytes)"
    success "Downloaded (${file_size} bytes)"

    info "Verifying checksum..."
    if curl -sL -o "${tmpdir}/checksums.txt" "$checksums_url" 2>/dev/null; then
        local expected="$(grep "${archive}" "${tmpdir}/checksums.txt" 2>/dev/null | awk '{print $1}' || true)"
        if [ -n "$expected" ]; then
            local actual
            actual="$(sha256sum "${tmpdir}/${archive}" 2>/dev/null | awk '{print $1}')" || \
            actual="$(shasum -a 256 "${tmpdir}/${archive}" 2>/dev/null | awk '{print $1}')" || \
            warn "No sha256sum/shasum found — skipping checksum"
            [ "$actual" = "$expected" ] || fatal "Checksum mismatch!"
            success "Checksum verified"
        fi
    fi

    info "Extracting..."
    tar -xzf "${tmpdir}/${archive}" -C "$tmpdir" || fatal "Extract failed"

    local install_dir="${HOME}/.local/bin"
    if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
        install_dir="/usr/local/bin"
    fi
    mkdir -p "$install_dir"
    cp "${tmpdir}/${BINARY_NAME}" "${install_dir}/${BINARY_NAME}"
    chmod +x "${install_dir}/${BINARY_NAME}"
    success "Installed to ${install_dir}/${BINARY_NAME}"

    if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
        warn "${install_dir} is not in your PATH"
        echo -e "  ${DIM}Add: export PATH=\"\$PATH:${install_dir}\"${NC}"
    fi
}

install_go() {
    step "Installing via go install"
    go install "github.com/${GITHUB_OWNER}/${GITHUB_REPO}/cmd/${BINARY_NAME}@latest" || \
        fatal "go install failed. Make sure Go 1.25+ is installed."

    local gobin="$(go env GOBIN)"
    [ -z "$gobin" ] && gobin="$(go env GOPATH)/bin"

    if [[ ":$PATH:" != *":${gobin}:"* ]]; then
        warn "${gobin} is not in your PATH"
        echo -e "  ${DIM}Add: export PATH=\"\$PATH:${gobin}\"${NC}"
    fi
    success "Installed via go install"
}

install_git() {
    step "Building from source"
    command -v git >/dev/null || fatal "git is required"
    command -v go >/dev/null || fatal "Go 1.25+ is required"

    local tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    git clone --depth 1 "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}.git" "$tmpdir" || \
        fatal "Failed to clone repository"
    cd "$tmpdir"
    go build -o "${BINARY_NAME}" ./cmd/plan-ai/ || fatal "Build failed"

    local install_dir="${HOME}/.local/bin"
    [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ] && install_dir="/usr/local/bin"
    mkdir -p "$install_dir"
    cp "${BINARY_NAME}" "${install_dir}/${BINARY_NAME}"
    chmod +x "${install_dir}/${BINARY_NAME}"

    if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
        warn "${install_dir} is not in your PATH"
        echo -e "  ${DIM}Add: export PATH=\"\$PATH:${install_dir}\"${NC}"
    fi
    success "Built and installed to ${install_dir}/${BINARY_NAME}"
}

integrate_opencode() {
    step "Configuring OpenCode integration"

    if ! command -v "$BINARY_NAME" &>/dev/null; then
        warn "Cannot find $BINARY_NAME in PATH — skipping OpenCode integration"
        return
    fi

    if "$BINARY_NAME" install --allow-real-opencode 2>&1; then
        success "Plan-AI registered in OpenCode (mcp.plan-ai)"
    else
        warn "OpenCode integration skipped (OpenCode not detected — standalone mode)"
        info  "Run 'plan-ai install --allow-real-opencode' later when OpenCode is available"
    fi
}

print_banner() {
    echo ""
    echo -e "${CYAN}${BOLD}"
    echo "  ____  _             _          _    ___ "
    echo " |  _ \| | __ _ _ __ | |_       / \  |_ _|"
    echo " | |_) | |/ _\` | '_ \| __|____ / _ \  | | "
    echo " |  __/| | (_| | | | | ||_____/ ___ \ | | "
    echo " |_|   |_|\__,_|_| |_|\__|   /_/   \_\___|"
    echo -e "${NC}"
    echo -e "  ${DIM}Local-first continuous implementation planning${NC}"
    echo ""
}

main() {
    setup_colors

    METHOD="${1:-go}"
    case "$METHOD" in
        binary|go|git) ;;
        -h|--help)
            echo "Usage: curl ... | bash [-s -- METHOD]"
            echo ""
            echo "Methods:"
            echo "  binary  Pre-built binary from GitHub Releases (default, auto-fallback to go)"
            echo "  go      go install from source"
            echo "  git     Clone and build from source"
            echo "  -h      Show this help"
            exit 0
            ;;
        *) fatal "Unknown method: $METHOD. Use: binary, go, or git" ;;
    esac

    print_banner

    case "$METHOD" in
        binary) detect_platform; install_binary ;;
        go)     install_go ;;
        git)    install_git ;;
    esac

    integrate_opencode

    echo ""
    echo -e "${GREEN}${BOLD}Plan-AI is ready!${NC}"
    echo ""
    echo -e "${BOLD}Next steps:${NC}"
    echo -e "  ${CYAN}1.${NC} cd your-project && ${BOLD}plan-ai init${NC}"
    echo -e "  ${CYAN}2.${NC} ${BOLD}plan-ai doctor${NC} to verify health"
    echo -e "  ${CYAN}3.${NC} ${BOLD}plan-ai intent create --description \"your idea\"${NC}"
    echo ""
    echo -e "${DIM}Docs: https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}${NC}"
    echo ""
}

main "$@"
