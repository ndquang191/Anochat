#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# PID file locations
BACKEND_PID_FILE=".backend.pid"
FRONTEND_PID_FILE=".frontend.pid"

echo -e "${BLUE}Stopping Anochat application...${NC}"

# Stop Backend
if [ -f "$BACKEND_PID_FILE" ]; then
    BACKEND_PID=$(cat "$BACKEND_PID_FILE")
    if ps -p $BACKEND_PID > /dev/null 2>&1; then
        echo -e "${YELLOW}Stopping Backend (PID: $BACKEND_PID)...${NC}"
        kill $BACKEND_PID
        echo -e "${GREEN}Backend stopped.${NC}"
    else
        echo -e "${RED}Backend process (PID: $BACKEND_PID) not found.${NC}"
    fi
    rm "$BACKEND_PID_FILE"
else
    echo -e "${YELLOW}No backend PID file found.${NC}"
fi

# Stop Frontend
if [ -f "$FRONTEND_PID_FILE" ]; then
    FRONTEND_PID=$(cat "$FRONTEND_PID_FILE")
    if ps -p $FRONTEND_PID > /dev/null 2>&1; then
        echo -e "${YELLOW}Stopping Frontend (PID: $FRONTEND_PID)...${NC}"
        kill $FRONTEND_PID
        echo -e "${GREEN}Frontend stopped.${NC}"
    else
        echo -e "${RED}Frontend process (PID: $FRONTEND_PID) not found.${NC}"
    fi
    rm "$FRONTEND_PID_FILE"
else
    echo -e "${YELLOW}No frontend PID file found.${NC}"
fi

echo ""
echo -e "${GREEN}All services stopped.${NC}"
