#!/bin/bash

tmux kill-session -t anochat-dev 2>/dev/null

tmux new-session -d -s anochat-dev

tmux new-window -t anochat-dev -n ui
tmux send-keys -t anochat-dev:ui "cd frontend && bun run dev" Enter

tmux new-window -t anochat-dev -n backend
tmux send-keys -t anochat-dev:backend "cd api && air" Enter

tmux select-window -t anochat-dev:backend
tmux attach -t anochat-dev
