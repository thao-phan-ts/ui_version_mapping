#!/usr/bin/env bash

# UI Version Mapping Tool Demo
# Showcase both local and remote functionality

set -e

echo "🎯 UI Version Mapping Tool - Enhanced Demo"
echo "=========================================="
echo ""

# Build the tool first
echo "🔨 Building the tool..."
make build
echo ""

# Demo 1: Local Analysis (if submodules exist)
if [ -d "scripts/submodules/digital_journey" ]; then
    echo "📁 Demo 1: Local Analysis"
    echo "-------------------------"
    echo "Using local submodules for analysis..."
    ./bin/ui-version-check -config 9054 -mode ab-testing
    echo ""
    
    echo "Journey analysis with local configs..."
    ./bin/ui-version-check -config 9054 -mode journey
    echo ""
else
    echo "📁 Demo 1: Local Analysis (Skipped)"
    echo "-----------------------------------"
    echo "⚠️  Local submodules not found. Run 'make setup-local' first."
    echo ""
fi

# Demo 2: Remote Analysis (if GitHub token is available)
if [ -n "$GITHUB_TOKEN" ]; then
    echo "🌐 Demo 2: Remote GitHub API Analysis"
    echo "------------------------------------"
    echo "Using GitHub API to fetch configs remotely..."
    ./bin/ui-version-check -config 9054 -remote -mode ab-testing
    echo ""
    
    echo "Journey analysis with remote configs..."
    ./bin/ui-version-check -config 9054 -remote -mode journey
    echo ""
    
    echo "Complete analysis with remote configs..."
    ./bin/ui-version-check -config 9054 -remote -mode complete
    echo ""
else
    echo "🌐 Demo 2: Remote Analysis (Skipped)"
    echo "-----------------------------------"
    echo "⚠️  GITHUB_TOKEN not set. Export your token to test remote functionality."
    echo "   export GITHUB_TOKEN=your_token_here"
    echo ""
fi

# Demo 3: Different Config Paths
echo "🔄 Demo 3: Different Config Paths"
echo "---------------------------------"
echo "Testing different lender config paths..."

if [ -n "$GITHUB_TOKEN" ]; then
    echo "Analyzing WIN configs remotely..."
    ./bin/ui-version-check -config 9012 -config-path win -remote -mode ab-testing
    echo ""
else
    echo "⚠️  Skipped (requires GITHUB_TOKEN)"
fi

# Demo 4: Help and Features
echo "❓ Demo 4: Tool Features"
echo "----------------------"
echo "Showing tool help and features..."
./bin/ui-version-check -help
echo ""

# Summary
echo "📊 Demo Summary"
echo "==============="
echo "✅ Enhanced Architecture Features:"
echo "   • Smart config provider selection (local/remote)"
echo "   • No dependency on auto_sync.sh"
echo "   • GitHub API integration with recursive loading"
echo "   • Environment-based configuration"
echo "   • Service-oriented architecture"
echo "   • Zero redundant functions"
echo ""
echo "🚀 Usage Examples:"
echo "   • Local:  ./bin/ui-version-check -config 9054"
echo "   • Remote: ./bin/ui-version-check -config 9054 -remote"
echo "   • Custom: ./bin/ui-version-check -config 9012 -config-path win"
echo ""
echo "🎉 Demo completed! The tool is ready for production use." 