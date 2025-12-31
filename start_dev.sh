#!/bin/bash

# Script to start development environment with tmux
# Frontend: bun run dev in frontend folder
# Backend: air command in api folder

# Check if tmux is installed
if ! command -v tmux &> /dev/null; then
    echo "tmux is not installed. Please install tmux first."
    echo "On Ubuntu/Debian: sudo apt install tmux"
    echo "On macOS: brew install tmux"
    exit 1
fi

# Check if frontend folder exists
if [ ! -d "frontend" ]; then
    echo "frontend folder not found!"
    exit 1
fi

# Check if api folder exists
if [ ! -d "api" ]; then
    echo "api folder not found!"
    exit 1
fi

# Kill any existing tmux session with the same name
tmux kill-session -t anochat-dev 2>/dev/null

# Create new tmux session
tmux new-session -d -s anochat-dev

# Create window for frontend
tmux new-window -t anochat-dev -n frontend
tmux send-keys -t anochat-dev:frontend "cd frontend" Enter
tmux send-keys -t anochat-dev:frontend "bun run dev" Enter

# Create window for backend
tmux new-window -t anochat-dev -n backend
tmux send-keys -t anochat-dev:backend "cd api" Enter
tmux send-keys -t anochat-dev:backend "air" Enter

# Switch to frontend window by default
tmux select-window -t anochat-dev:frontend

# Attach to the session
echo "Starting development environment..."
echo "Frontend: bun run dev"
echo "Backend: air"
echo ""
echo "Use Ctrl+B then D to detach from tmux"
echo "Use 'tmux attach -t anochat-dev' to reattach"
echo ""

tmux attach -t anochat-dev
