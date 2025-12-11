#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# PID file locations
BACKEND_PID_FILE=".backend.pid"
FRONTEND_PID_FILE=".frontend.pid"

echo -e "${BLUE}Starting Anochat application...${NC}"

# Start Backend
echo -e "${GREEN}Starting Backend (Go API)...${NC}"
cd api
go run cmd/server/main.go > ../backend.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > "../$BACKEND_PID_FILE"
cd ..
echo -e "${GREEN}Backend started with PID: $BACKEND_PID${NC}"
echo -e "Backend logs: backend.log"

# Wait a bit for backend to initialize
sleep 2

# Start Frontend
echo -e "${GREEN}Starting Frontend (Next.js)...${NC}"
cd frontend
npm run dev > ../frontend.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > "../$FRONTEND_PID_FILE"
cd ..
echo -e "${GREEN}Frontend started with PID: $FRONTEND_PID${NC}"
echo -e "Frontend logs: frontend.log"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Both services are now running!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Backend PID: $BACKEND_PID (saved to $BACKEND_PID_FILE)"
echo -e "Frontend PID: $FRONTEND_PID (saved to $FRONTEND_PID_FILE)"
echo ""
echo -e "To view logs in real-time:"
echo -e "  Backend:  tail -f backend.log"
echo -e "  Frontend: tail -f frontend.log"
echo ""
echo -e "To stop the services, run: ./stop.sh"
echo ""
