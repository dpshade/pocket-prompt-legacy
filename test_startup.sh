#!/bin/bash

# Test startup performance with async loading
cd /Users/dps/Developer/My\ Repos/pocket-prompt

echo "Testing startup performance..."
export POCKET_PROMPT_DIR="$(pwd)/test-library"

# Build the application
echo "Building application..."
go build -o pocket-prompt-test .

echo ""
echo "Starting application with test data..."
echo "The UI should appear immediately, then load data asynchronously."
echo "Press 'q' to quit after observing the startup behavior."
echo ""

# Run the application
./pocket-prompt-test