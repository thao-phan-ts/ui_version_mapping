#!/usr/bin/env bash

# UI Version Mapping Tool Setup Script
# Thay thế cho auto_sync.sh với cách tiếp cận tối ưu hơn

set -e

echo "🚀 UI Version Mapping Tool Setup"
echo "================================="

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

# Function to setup local development with submodules (optional)
setup_local_dev() {
    echo "📁 Setting up local development environment..."
    
    if [ ! -d "scripts/submodules" ]; then
        echo "📥 Cloning digital_journey repository..."
        mkdir -p scripts/submodules
        cd scripts/submodules
        
        # Clone with sparse checkout for better performance
        git clone --filter=blob:none --no-checkout --depth 1 --sparse \
            git@github.com:tsocial/digital_journey.git
        
        cd digital_journey
        git config core.sparseCheckout true
        echo "migration/sync/vietnam/tpbank/lender_configs" > .git/info/sparse-checkout
        git checkout
        
        cd ../../..
        echo "✅ Local configs available at: scripts/submodules/digital_journey/"
    else
        echo "✅ Local development environment already exists"
    fi
}

# Function to setup remote-only development
setup_remote_dev() {
    echo "🌐 Setting up remote development environment..."
    
    # Copy environment template
    if [ ! -f ".env" ]; then
        cp config.example.env .env
        echo "📝 Created .env file from template"
        echo "⚠️  Please edit .env file and add your GITHUB_TOKEN if needed"
    fi
    
    echo "✅ Remote development environment ready"
    echo "💡 Use -remote flag to fetch configs from GitHub API"
}

# Function to clean up old submodules
cleanup_old() {
    echo "🧹 Cleaning up old submodules..."
    if [ -d "scripts/submodules" ]; then
        rm -rf scripts/submodules
        echo "✅ Old submodules removed"
    fi
}

# Main setup logic
case "${1:-auto}" in
    "local")
        setup_local_dev
        ;;
    "remote")
        setup_remote_dev
        ;;
    "clean")
        cleanup_old
        ;;
    "auto")
        echo "🤖 Auto-detecting best setup method..."
        
        # Check if we have SSH access to GitHub
        if ssh -T git@github.com 2>&1 | grep -q "successfully authenticated"; then
            echo "✅ GitHub SSH access detected"
            setup_local_dev
        else
            echo "⚠️  No GitHub SSH access, using remote API mode"
            setup_remote_dev
        fi
        ;;
    *)
        echo "Usage: $0 [local|remote|clean|auto]"
        echo ""
        echo "Options:"
        echo "  local   - Setup with local git submodules (requires SSH access)"
        echo "  remote  - Setup for remote GitHub API access"
        echo "  clean   - Remove old submodules"
        echo "  auto    - Auto-detect best method (default)"
        echo ""
        echo "Examples:"
        echo "  $0 local    # Clone configs locally"
        echo "  $0 remote   # Use GitHub API"
        echo "  $0 clean    # Clean old setup"
        exit 1
        ;;
esac

echo ""
echo "🎉 Setup completed!"
echo ""
echo "Next steps:"
echo "1. Build the tool: make build"
echo "2. Run analysis: ./bin/ui-version-check -config 9054"
echo "3. Or run tests: make test"
echo ""
echo "For help: ./bin/ui-version-check -help" 