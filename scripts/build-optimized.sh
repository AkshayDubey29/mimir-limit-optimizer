#!/bin/bash

# =============================================================================
# OPTIMIZED BUILD SCRIPT - Mimir Limit Optimizer
# =============================================================================
# This script demonstrates optimized build processes using the ignore files
# for maximum build speed and efficiency
# =============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_REGISTRY="${DOCKER_REGISTRY:-}"
IMAGE_NAME="${IMAGE_NAME:-mimir-limit-optimizer}"
VERSION="${VERSION:-$(git rev-parse --short HEAD 2>/dev/null || echo 'dev')}"
BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
GIT_COMMIT="$(git rev-parse HEAD 2>/dev/null || echo 'unknown')"

# Function to print colored output
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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to build UI only
build_ui() {
    print_status "Building UI with optimized npm operations..."
    
    cd ui
    
    # Clean npm cache for faster installs
    if [ -d "node_modules" ]; then
        print_status "Cleaning existing node_modules..."
        rm -rf node_modules
    fi
    
    # Use npm ci for faster, reproducible builds
    print_status "Installing dependencies with npm ci (faster than npm install)..."
    time npm ci --prefer-offline --no-audit --no-fund --silent
    
    # Build with production optimizations
    print_status "Building UI for production..."
    time npm run build
    
    # Show build output size
    if [ -d "build" ]; then
        print_success "UI build completed!"
        print_status "Build size analysis:"
        du -sh build/
        find build -name "*.js" -o -name "*.css" | xargs ls -lh
    fi
    
    cd ..
}

# Function to build Go binary
build_go() {
    print_status "Building Go binary with optimizations..."
    
    # Clean previous builds
    rm -f bin/mimir-limit-optimizer
    mkdir -p bin
    
    # Build with optimizations
    print_status "Compiling Go binary..."
    time CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
        -a -installsuffix cgo \
        -o bin/mimir-limit-optimizer \
        main.go
    
    # Show binary info
    if [ -f "bin/mimir-limit-optimizer" ]; then
        print_success "Go binary built successfully!"
        print_status "Binary info:"
        ls -lh bin/mimir-limit-optimizer
        file bin/mimir-limit-optimizer
    fi
}

# Function to build Docker image
build_docker() {
    local dockerfile="${1:-Dockerfile.optimized}"
    local tag="${2:-${IMAGE_NAME}:${VERSION}}"
    
    print_status "Building Docker image with optimized Dockerfile..."
    print_status "Using dockerfile: ${dockerfile}"
    print_status "Tag: ${tag}"
    
    # Show build context size (should be small due to .dockerignore)
    print_status "Analyzing build context size..."
    docker build --dry-run -f "${dockerfile}" . >/dev/null 2>&1 || true
    
    # Build with BuildKit for better performance
    print_status "Building with Docker BuildKit for better performance..."
    time DOCKER_BUILDKIT=1 docker build \
        -f "${dockerfile}" \
        --build-arg VERSION="${VERSION}" \
        --build-arg BUILD_DATE="${BUILD_DATE}" \
        --build-arg GIT_COMMIT="${GIT_COMMIT}" \
        -t "${tag}" \
        .
    
    # Show image info
    if docker image inspect "${tag}" >/dev/null 2>&1; then
        print_success "Docker image built successfully!"
        print_status "Image info:"
        docker image inspect "${tag}" --format='{{.Size}}' | numfmt --to=iec
        docker image ls "${tag}"
    fi
}

# Function to run performance tests
test_performance() {
    print_status "Running performance tests..."
    
    # Test UI build performance
    if [ -d "ui" ]; then
        print_status "Testing UI build speed..."
        cd ui
        time npm run build >/dev/null 2>&1
        cd ..
    fi
    
    # Test Go build performance
    print_status "Testing Go build speed..."
    time go build -o /tmp/test-binary main.go >/dev/null 2>&1
    rm -f /tmp/test-binary
    
    print_success "Performance tests completed!"
}

# Function to show build context analysis
analyze_build_context() {
    print_status "Analyzing build context (what Docker sees)..."
    
    # Create a temporary tar to see what gets included
    tar --exclude-from=.dockerignore -cf /tmp/build-context.tar . 2>/dev/null || true
    
    if [ -f "/tmp/build-context.tar" ]; then
        print_status "Build context size: $(du -sh /tmp/build-context.tar | cut -f1)"
        print_status "Files included in build context:"
        tar -tf /tmp/build-context.tar | head -20
        echo "... (truncated, showing first 20 files)"
        rm -f /tmp/build-context.tar
    fi
}

# Function to clean build artifacts
clean() {
    print_status "Cleaning build artifacts..."
    
    # Clean Go artifacts
    rm -rf bin/
    go clean -cache -modcache -testcache 2>/dev/null || true
    
    # Clean UI artifacts
    if [ -d "ui" ]; then
        cd ui
        rm -rf node_modules/ build/ .cache/
        npm cache clean --force 2>/dev/null || true
        cd ..
    fi
    
    # Clean Docker artifacts
    docker system prune -f >/dev/null 2>&1 || true
    
    print_success "Cleanup completed!"
}

# Function to show help
show_help() {
    cat << EOF
Optimized Build Script for Mimir Limit Optimizer

USAGE:
    $0 [COMMAND] [OPTIONS]

COMMANDS:
    ui              Build UI only
    go              Build Go binary only
    docker          Build Docker image
    all             Build everything (UI + Go + Docker)
    test            Run performance tests
    analyze         Analyze build context
    clean           Clean all build artifacts
    help            Show this help

OPTIONS:
    --dockerfile    Specify Dockerfile to use (default: Dockerfile.optimized)
    --tag           Specify Docker tag (default: ${IMAGE_NAME}:${VERSION})
    --registry      Specify Docker registry

EXAMPLES:
    $0 all                                  # Build everything
    $0 docker --dockerfile Dockerfile      # Build with standard Dockerfile
    $0 docker --tag myapp:v1.0.0          # Build with custom tag
    
ENVIRONMENT VARIABLES:
    DOCKER_REGISTRY     Docker registry URL
    IMAGE_NAME          Docker image name (default: mimir-limit-optimizer)
    VERSION             Build version (default: git short hash)

PERFORMANCE OPTIMIZATIONS:
    ✓ Uses .dockerignore to reduce build context
    ✓ Uses .npmignore for faster npm operations
    ✓ Uses npm ci instead of npm install
    ✓ Multi-stage Docker builds for minimal images
    ✓ BuildKit for better Docker build performance
    ✓ Go build optimizations (-ldflags="-w -s")
    ✓ Parallel builds where possible
EOF
}

# Main execution
main() {
    local command="${1:-help}"
    shift || true
    
    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dockerfile)
                DOCKERFILE="$2"
                shift 2
                ;;
            --tag)
                TAG="$2"
                shift 2
                ;;
            --registry)
                DOCKER_REGISTRY="$2"
                shift 2
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Set defaults
    DOCKERFILE="${DOCKERFILE:-Dockerfile.optimized}"
    TAG="${TAG:-${IMAGE_NAME}:${VERSION}}"
    
    # Add registry prefix if specified
    if [ -n "${DOCKER_REGISTRY}" ]; then
        TAG="${DOCKER_REGISTRY}/${TAG}"
    fi
    
    # Check prerequisites
    if ! command_exists docker; then
        print_error "Docker is required but not installed"
        exit 1
    fi
    
    if ! command_exists go; then
        print_error "Go is required but not installed"
        exit 1
    fi
    
    if [ -d "ui" ] && ! command_exists npm; then
        print_error "npm is required but not installed"
        exit 1
    fi
    
    # Execute command
    case $command in
        ui)
            build_ui
            ;;
        go)
            build_go
            ;;
        docker)
            build_docker "$DOCKERFILE" "$TAG"
            ;;
        all)
            print_status "Building all components..."
            build_ui
            build_go
            build_docker "$DOCKERFILE" "$TAG"
            print_success "All builds completed!"
            ;;
        test)
            test_performance
            ;;
        analyze)
            analyze_build_context
            ;;
        clean)
            clean
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

# Run main function with all arguments
main "$@" 