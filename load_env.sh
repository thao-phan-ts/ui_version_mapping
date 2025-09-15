#!/usr/bin/env bash

# Script to safely load environment variables from .env file
# Usage: source ./load_env.sh

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if .env file exists
if [ -f ".env" ]; then
    echo -e "${GREEN}📋 Loading environment variables from .env${NC}"
    
    # Load variables from .env file
    set -a  # Mark variables for export
    source .env
    set +a  # Stop marking for export
    
    # Verify critical variables
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        echo -e "${GREEN}✅ GITHUB_TOKEN loaded${NC} (${#GITHUB_TOKEN} characters)"
    else
        echo -e "${YELLOW}⚠️  GITHUB_TOKEN not found in .env${NC}"
    fi
    
    if [ -n "${TS_TOKEN_TEST:-}" ]; then
        echo -e "${GREEN}✅ TS_TOKEN_TEST loaded${NC} (${#TS_TOKEN_TEST} characters)"
    fi
    
    if [ -n "${TSOCIAL_ACCESS_TOKEN:-}" ]; then
        echo -e "${GREEN}✅ TSOCIAL_ACCESS_TOKEN loaded${NC} (${#TSOCIAL_ACCESS_TOKEN} characters)"
    fi
    
    echo -e "${GREEN}✅ Environment variables loaded successfully${NC}"
else
    echo -e "${RED}❌ .env file not found${NC}"
    echo ""
    echo "To create one:"
    echo "1. Copy the example file: cp env.example .env"
    echo "2. Edit .env and add your tokens"
    echo "3. Run: source ./load_env.sh"
    
    # Offer to create from example
    if [ -f "env.example" ]; then
        echo ""
        read -p "Would you like to create .env from env.example now? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            cp env.example .env
            echo -e "${GREEN}✅ Created .env from env.example${NC}"
            echo "Please edit .env and add your actual tokens"
        fi
    fi
fi 