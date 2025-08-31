#!/bin/bash

echo "Testing Onyx TUI launcher..."
echo "Current directory: $(pwd)"
echo "TERM: $TERM"
echo "TTY: $(tty)"

# Test if we're in a terminal
if [ -t 0 ] && [ -t 1 ]; then
    echo "✓ Running in a terminal"
else
    echo "✗ Not running in a terminal"
fi

# Check if the binary exists
if [ -f ~/.local/lib/onyx-tui/onyx-tui-bin ]; then
    echo "✓ Binary exists"
else
    echo "✗ Binary not found"
fi

# Check if bubbletea dependencies are met
echo ""
echo "Attempting to run TUI..."
cd ~/.local/lib/onyx-tui
export ONYX_WORK_DIR="$(pwd)"
./onyx-tui-bin