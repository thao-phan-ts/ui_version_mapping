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

# ============================================================================
# Core Functions
# ============================================================================

# Initialize and sparse checkout a repository
sparse_checkout_repo() {
    local repo_name="$1"
    local checkout_dirs="$2"
    local target_dir="${3:-$repo_name}"
    
    print_separator
    log_info "Setting up sparse checkout for: $repo_name"
    log_info "Directories to checkout: $checkout_dirs"
    
    # Clean up existing directory
    if [ -d "$target_dir" ]; then
        log_warn "Removing existing directory: $target_dir"
        rm -rf "$target_dir"
    fi
    
    # Clone with sparse checkout
    log_info "Cloning repository..."
    git clone \
        --filter=blob:none \
        --no-checkout \
        --depth 1 \
        --sparse \
        "git@github.com:${GITHUB_ORG}/${repo_name}.git" \
        "$target_dir"
    
    cd "$target_dir"
    
    # Configure sparse checkout
    git config core.sparseCheckout true
    
    # Add directories to sparse checkout
    for dir in $checkout_dirs; do
        log_info "Adding to sparse checkout: $dir"
        git sparse-checkout add "$dir"
    done
    
    # Checkout and pull latest
    git checkout
    git pull origin master
    
    cd ..
    log_info "✓ Successfully set up $repo_name"
}

# Initialize all repositories with sparse checkout
init_repositories() {
    log_info "Initializing repositories with sparse checkout..."
    
    ensure_directory "$ROOT_DIR"
    cd "$ROOT_DIR"
    
    local num_repos=${#REPOS[@]}
    for ((i=0; i<num_repos; i++)); do
        sparse_checkout_repo "${REPOS[$i]}" "${CHECKOUT_DIRS[$i]}"
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
            
            # Check for uncommitted changes
            if ! git diff-index --quiet HEAD -- 2>/dev/null; then
                log_warn "Uncommitted changes detected in $dir_name, skipping update"
            else
                git pull origin master || log_error "Failed to update $dir_name"
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

REPOSITORIES:
EOF
    local num_repos=${#REPOS[@]}
    for ((i=0; i<num_repos; i++)); do
        echo "    - ${REPOS[$i]}: ${CHECKOUT_DIRS[$i]}"
    done
    
    cat << EOF

CONFIGURATION:
    Root Directory: $ROOT_DIR
    GitHub Organization: $GITHUB_ORG
    
EOF
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
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

