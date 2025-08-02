// lib/socket.ts
import { io, Socket } from "socket.io-client";

const socket: Socket = io(process.env.NEXT_PUBLIC_API_URL!, {
	transports: ["websocket"],
	withCredentials: true,
	autoConnect: true,
});

export default socket;
