#!/usr/bin/env bash

set -euo pipefail  # Exit on error, undefined variables, and pipe failures

# ============================================================================
# Configuration
# ============================================================================
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="${SCRIPT_DIR}/submodules"
readonly GITHUB_ORG="tsocial"

# Repository configurations - parallel arrays
REPOS=("digital_journey" "decision_engine")
CHECKOUT_DIRS=("migration" "etc/production migration")
# Default versions (branch/tag/commit) - can be overridden via environment variables
REPO_VERSIONS=("master" "master")



# Color codes for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# ============================================================================
# Helper Functions
# ============================================================================

# Print colored output
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_separator() {
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

# Check if directory is a git repository
is_git_repo() {
    local dir="$1"
    [ -d "$dir/.git" ] || git -C "$dir" rev-parse --git-dir &>/dev/null
}

# Create directory if it doesn't exist
ensure_directory() {
    local dir="$1"
    if [ ! -d "$dir" ]; then
        log_info "Creating directory: $dir"
        mkdir -p "$dir"
    fi
}

# Get version for a repository (from environment variable or default)
get_repo_version() {
    local repo_name="$1"
    local default_version="$2"
    
    # Convert repo name to uppercase and replace hyphens with underscores for env var
    local env_var_name=$(echo "$repo_name" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
    env_var_name="${env_var_name}_VERSION"
    
    # Get version from environment variable or use default
    local version=$(eval echo "\${$env_var_name:-$default_version}")
    echo "$version"
}

# ============================================================================
# Core Functions
# ============================================================================

# Initialize and sparse checkout a repository
sparse_checkout_repo() {
    local repo_name="$1"
    local checkout_dirs="$2"
    local version="$3"
    local target_dir="${4:-$repo_name}"
    
    print_separator
    log_info "Setting up sparse checkout for: $repo_name"
    log_info "Directories to checkout: $checkout_dirs"
    log_info "Version/Branch: $version"
    
    # Clean up existing directory
    if [ -d "$target_dir" ]; then
        log_warn "Removing existing directory: $target_dir"
        rm -rf "$target_dir"
    fi
    
    # Determine repository URL based on environment
    local repo_url
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        # Directly use authenticated URL in CI with token
        repo_url="https://x-access-token:${GITHUB_TOKEN}@github.com/${GITHUB_ORG}/${repo_name}.git"
        log_info "Using authenticated HTTPS URL"
    elif [ -n "${CI:-}" ] || [ -n "${GITHUB_ACTIONS:-}" ]; then
        # CI without token, try public access
        repo_url="https://github.com/${GITHUB_ORG}/${repo_name}.git"
        log_info "Using public HTTPS URL (CI environment)"
    else
        # Local development, use SSH
        repo_url="git@github.com:${GITHUB_ORG}/${repo_name}.git"
        log_info "Using SSH URL (local environment)"
    fi
    
    # Clone with sparse checkout
    log_info "Cloning repository: ${repo_name}"
    
    # Disable terminal prompts for git
    export GIT_TERMINAL_PROMPT=0
    export GIT_ASKPASS=/bin/echo
    
    # Clone and suppress token in output
    git clone \
        --filter=blob:none \
        --no-checkout \
        --sparse \
        "$repo_url" \
        "$target_dir" 2>&1 | sed 's/x-access-token:[^@]*@/x-access-token:***@/g' || {
        log_error "Failed to clone ${repo_name}"
        return 1
    }
    
    cd "$target_dir"
    
    # Configure sparse checkout
    git config core.sparseCheckout true
    
    # Add directories to sparse checkout
    for dir in $checkout_dirs; do
        log_info "Adding to sparse checkout: $dir"
        git sparse-checkout add "$dir"
    done
    
    # Fetch all refs if checking out a specific version
    if [ "$version" != "master" ] && [ "$version" != "main" ]; then
        log_info "Fetching all refs for version checkout..."
        git fetch --all --tags
    fi
    
    # Checkout specified version
    log_info "Checking out version: $version"
    git checkout "$version" || {
        log_error "Failed to checkout $version, trying as remote branch..."
        git checkout -b "$version" "origin/$version" || {
            log_error "Failed to checkout version: $version"
            exit 1
        }
    }
    
    cd ..
    log_info "✓ Successfully set up $repo_name at version $version"
}

# Initialize all repositories with sparse checkout
init_repositories() {
    log_info "Initializing repositories with sparse checkout..."
    
    ensure_directory "$ROOT_DIR"
    cd "$ROOT_DIR"
    
    local num_repos=${#REPOS[@]}
    for ((i=0; i<num_repos; i++)); do
        local repo="${REPOS[$i]}"
        local dirs="${CHECKOUT_DIRS[$i]}"
        local default_version="${REPO_VERSIONS[$i]}"
        local version=$(get_repo_version "$repo" "$default_version")
        
        sparse_checkout_repo "$repo" "$dirs" "$version"
    done
    
    cd "$SCRIPT_DIR"
    log_info "✓ All repositories initialized"
}

# Update existing repositories
update_repositories() {
    log_info "Updating existing repositories..."
    
    if [ ! -d "$ROOT_DIR" ]; then
        log_error "Submodules directory not found. Run with '1' argument to initialize first."
        exit 1
    fi
    
    for dir in "$ROOT_DIR"/*; do
        [ ! -d "$dir" ] && continue
        
        local dir_name=$(basename "$dir")
        
        print_separator
        
        if is_git_repo "$dir"; then
            log_info "Updating: $dir_name"
            cd "$dir"
            
            # Find the repo index to get its version
            local repo_index=-1
            for ((i=0; i<${#REPOS[@]}; i++)); do
                if [ "${REPOS[$i]}" = "$dir_name" ]; then
                    repo_index=$i
                    break
                fi
            done
            
            if [ $repo_index -ne -1 ]; then
                local default_version="${REPO_VERSIONS[$repo_index]}"
                local version=$(get_repo_version "$dir_name" "$default_version")
                
                # Check for uncommitted changes
                if ! git diff-index --quiet HEAD -- 2>/dev/null; then
                    log_warn "Uncommitted changes detected in $dir_name, skipping update"
                else
                    # Get current branch/tag
                    local current_ref=$(git symbolic-ref -q --short HEAD || git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
                    
                    if [ "$current_ref" != "$version" ]; then
                        log_warn "Switching from $current_ref to $version"
                        git fetch --all --tags
                        git checkout "$version" || log_error "Failed to checkout $version"
                    else
                        log_info "Already on $version, pulling latest changes..."
                        git pull origin "$version" || log_info "No remote tracking branch, skipping pull"
                    fi
                fi
            else
                log_warn "$dir_name not found in repository list, skipping version check"
                git pull || log_info "Pull failed, repository might be on a detached HEAD"
            fi
            
            cd "$SCRIPT_DIR"
        else
            log_warn "$dir_name is not a git repository, skipping"
        fi
    done
    
    log_info "✓ Repository updates complete"
}


# Display usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Synchronize and manage sparse checkouts of multiple repositories.

OPTIONS:
    1           Initialize all repositories with sparse checkout
    (no args)   Update existing repositories
    -h, --help  Show this help message

REPOSITORIES & VERSIONS:
EOF
    local num_repos=${#REPOS[@]}
    for ((i=0; i<num_repos; i++)); do
        local repo="${REPOS[$i]}"
        local default_version="${REPO_VERSIONS[$i]}"
        local version=$(get_repo_version "$repo" "$default_version")
        echo "    - $repo: ${CHECKOUT_DIRS[$i]}"
        echo "        Version: $version (default: $default_version)"
    done
    
    cat << EOF

VERSION CONTROL:
    You can override the version for each repository using environment variables:
    - DIGITAL_JOURNEY_VERSION: Version for digital_journey (default: master)
    - DECISION_ENGINE_VERSION: Version for decision_engine (default: master)
    
    Examples:
        DIGITAL_JOURNEY_VERSION=v1.2.3 $0 1    # Init with specific tag
        DECISION_ENGINE_VERSION=develop $0     # Update with specific branch
        
    Versions can be:
    - Branch names (master, develop, feature/xyz)
    - Tags (v1.2.3, release-2024)
    - Commit hashes (abc123def)

CONFIGURATION:
    Root Directory: $ROOT_DIR
    GitHub Organization: $GITHUB_ORG
    
EOF
}

# ============================================================================
# Main Execution
# ============================================================================

# Setup git authentication if in CI environment
setup_git_auth() {
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        log_info "Configuring git authentication for CI environment"
        
        # Method 1: URL rewriting (primary)
        git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
        git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "git@github.com:"
        git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com/".insteadOf "ssh://git@github.com/"
        
        # Method 2: Credential helper as backup
        git config --global credential.helper store
        echo "https://x-access-token:${GITHUB_TOKEN}@github.com" > ~/.git-credentials
        
        # Method 3: Set authorization header
        git config --global http.https://github.com/.extraheader "Authorization: token ${GITHUB_TOKEN}"
        
        # Disable terminal prompts
        export GIT_TERMINAL_PROMPT=0
        export GIT_ASKPASS=/bin/echo
        
        log_info "Git authentication configured successfully"
    fi
}

main() {
    # Setup authentication first if needed
    setup_git_auth
    
    case "${1:-}" in
        1)
            init_repositories
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        "")
            update_repositories
            ;;
        *)
            log_error "Invalid argument: $1"
            show_usage
            exit 1
            ;;
    esac
    
    print_separator
    log_info "✓ All operations completed successfully"
}

# Run main function
main "$@"

