#!/bin/bash

# =============================================================================
# NPM SPEED OPTIMIZATION SCRIPT - Mimir Limit Optimizer
# =============================================================================
# This script optimizes npm for maximum speed and performance
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to optimize npm configuration
optimize_npm_config() {
    print_status "Optimizing npm configuration for maximum speed..."
    
    # Disable unnecessary features for speed
    npm config set progress false --location=user
    npm config set audit false --location=user
    npm config set fund false --location=user
    npm config set update-notifier false --location=user
    
    # Speed optimizations
    npm config set fetch-retries 1 --location=user
    npm config set fetch-retry-mintimeout 2000 --location=user
    npm config set fetch-retry-maxtimeout 5000 --location=user
    npm config set fetch-timeout 30000 --location=user
    
    # Use faster registry if available
    npm config set registry https://registry.npmjs.org/ --location=user
    
    # Set maxsockets for parallel downloads
    npm config set maxsockets 50 --location=user
    
    # Disable package-lock for faster installs (use with caution)
    # npm config set package-lock false --location=user
    
    print_success "npm configuration optimized!"
}

# Function to clean all npm caches
clean_npm_caches() {
    print_status "Cleaning all npm caches..."
    
    # Clean npm cache
    npm cache clean --force 2>/dev/null || true
    
    # Clean npm verify cache
    npm cache verify 2>/dev/null || true
    
    # Remove global npm cache directory
    local npm_cache_dir=$(npm config get cache 2>/dev/null || echo "")
    if [ -n "$npm_cache_dir" ] && [ -d "$npm_cache_dir" ]; then
        print_status "Cleaning cache directory: $npm_cache_dir"
        rm -rf "$npm_cache_dir" 2>/dev/null || true
    fi
    
    print_success "npm caches cleaned!"
}

# Function to optimize project-specific npm
optimize_project_npm() {
    print_status "Optimizing project npm setup..."
    
    cd ui || { print_error "ui directory not found!"; exit 1; }
    
    # Remove existing node_modules and lock file for fresh start
    print_status "Removing existing node_modules and lock files..."
    rm -rf node_modules package-lock.json 2>/dev/null || true
    
    # Create .npmrc for project-specific optimizations
    print_status "Creating optimized .npmrc..."
    cat > .npmrc << EOF
# Speed optimizations
progress=false
audit=false
fund=false
update-notifier=false

# Network optimizations
fetch-retries=1
fetch-retry-mintimeout=2000
fetch-retry-maxtimeout=5000
fetch-timeout=30000
maxsockets=50

# Registry optimization
registry=https://registry.npmjs.org/

# Cache optimization
prefer-offline=true
prefer-dedupe=true

# Installation optimizations
engine-strict=false
optional=false
EOF
    
    cd ..
    print_success "Project npm optimized!"
}

# Function to test npm speed
test_npm_speed() {
    print_status "Testing npm speed with optimizations..."
    
    cd ui || { print_error "ui directory not found!"; exit 1; }
    
    print_status "Running optimized npm install test..."
    time npm ci --silent --prefer-offline --no-audit --no-fund 2>/dev/null
    
    print_status "Testing npm build speed..."
    if [ -d "node_modules" ]; then
        time npm run build --silent 2>/dev/null || print_warning "Build failed, but dependencies are installed"
    fi
    
    cd ..
    print_success "npm speed test completed!"
}

# Function to show npm info
show_npm_info() {
    print_status "Current npm configuration:"
    echo "Node version: $(node --version)"
    echo "npm version: $(npm --version)"
    echo "Registry: $(npm config get registry)"
    echo "Cache directory: $(npm config get cache)"
    echo "Progress: $(npm config get progress)"
    echo "Audit: $(npm config get audit)"
    echo "Fund: $(npm config get fund)"
    echo "Max sockets: $(npm config get maxsockets)"
}

# Function to use yarn as alternative (faster option)
setup_yarn_alternative() {
    print_status "Setting up Yarn as faster alternative..."
    
    # Install yarn if not present
    if ! command -v yarn >/dev/null 2>&1; then
        print_status "Installing Yarn..."
        npm install -g yarn --silent
    fi
    
    cd ui || { print_error "ui directory not found!"; exit 1; }
    
    # Remove npm lock file and use yarn
    rm -f package-lock.json
    
    # Configure yarn for speed
    yarn config set network-timeout 30000
    yarn config set network-concurrency 50
    yarn config set disable-self-update-check true
    
    print_status "Installing dependencies with Yarn..."
    time yarn install --silent --prefer-offline --no-audit --no-fund-raising
    
    print_status "Testing Yarn build..."
    time yarn build --silent 2>/dev/null || print_warning "Build test completed"
    
    cd ..
    print_success "Yarn setup completed!"
    print_status "To use Yarn: cd ui && yarn install && yarn build"
}

# Function to show help
show_help() {
    cat << EOF
NPM Speed Optimization Script

USAGE:
    $0 [COMMAND]

COMMANDS:
    config      Optimize npm configuration only
    clean       Clean all npm caches
    project     Optimize project-specific npm setup
    test        Test npm speed after optimizations
    info        Show current npm configuration
    yarn        Setup Yarn as faster alternative
    all         Run all optimizations (config + clean + project + test)
    help        Show this help

EXAMPLES:
    $0 all      # Complete npm optimization
    $0 yarn     # Switch to Yarn for faster builds
    $0 test     # Test current npm speed

OPTIMIZATIONS APPLIED:
    ✓ Disable progress bars, audits, fund messages
    ✓ Reduce network timeouts and retries
    ✓ Increase parallel connections
    ✓ Clean all caches
    ✓ Use offline-first approach
    ✓ Project-specific .npmrc configuration
EOF
}

# Main execution
main() {
    local command="${1:-help}"
    
    case $command in
        config)
            optimize_npm_config
            ;;
        clean)
            clean_npm_caches
            ;;
        project)
            optimize_project_npm
            ;;
        test)
            test_npm_speed
            ;;
        info)
            show_npm_info
            ;;
        yarn)
            setup_yarn_alternative
            ;;
        all)
            print_status "Running complete npm optimization..."
            optimize_npm_config
            clean_npm_caches
            optimize_project_npm
            test_npm_speed
            print_success "Complete npm optimization finished!"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@" 