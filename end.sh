#!/bin/bash

tmux kill-session -t anochat-dev 2>/dev/null && echo "Stopped." || echo "Not running."
