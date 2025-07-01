#!/bin/bash

set -e

echo "Building Mimir Limit Optimizer UI..."

# Navigate to UI directory
cd ui

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing UI dependencies..."
    npm install
fi

# Build the React app
echo "Building React app..."
npm run build

echo "UI build completed successfully!"

# Verify build directory exists
if [ ! -d "build" ]; then
    echo "Error: Build directory not found!"
    exit 1
fi

echo "Build artifacts created in ui/build/" 