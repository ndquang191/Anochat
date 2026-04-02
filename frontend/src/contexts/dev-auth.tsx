"use client";

import { ReactNode } from "react";
import { UserDTO, RoomDTO, MessageDTO } from "@/types";
import { AuthContext } from "./auth";

export const DEV_USER: UserDTO = {
	id: "dev-user-001",
	email: "dev@anochat.dev",
	name: "Dev User",
	profile: { age: 25, is_male: true, is_hidden: false },
};

export const DEV_PARTNER: UserDTO = {
	id: "dev-partner-001",
	name: "Partner User",
	profile: { age: 22, is_male: false, is_hidden: false },
};

export const DEV_ROOM: RoomDTO = {
	id: "dev-room-001",
	user1_id: "dev-user-001",
	user2_id: "dev-partner-001",
	partner: DEV_PARTNER,
};

export const DEV_MESSAGES: MessageDTO[] = [
	{ id: "m1", room_id: "dev-room-001", sender_id: "dev-partner-001", content: "Xin chào! Bạn khỏe không?", created_at: Date.now() - 60000 },
	{ id: "m2", room_id: "dev-room-001", sender_id: "dev-user-001", content: "Chào! Mình khỏe, cảm ơn bạn nhé 😊", created_at: Date.now() - 50000 },
	{ id: "m3", room_id: "dev-room-001", sender_id: "dev-partner-001", content: "Bạn đang làm gì vậy?", created_at: Date.now() - 30000 },
	{ id: "m4", room_id: "dev-room-001", sender_id: "dev-user-001", content: "Đang dev UI cho AnoChat 🛠️", created_at: Date.now() - 10000 },
];

export type DevState = "lobby" | "queue" | "chat";

interface DevAuthProviderProps {
	children: ReactNode;
	devState: DevState;
}

export function DevAuthProvider({ children, devState }: DevAuthProviderProps) {
	const room = devState === "chat" ? DEV_ROOM : null;
	const inQueue = devState === "queue";
	const messages = devState === "chat" ? DEV_MESSAGES : [];

	return (
		<AuthContext.Provider
			value={{
				isAuthenticated: true,
				user: DEV_USER,
				room,
				messages,
				inQueue,
				loading: false,
				login: () => {},
				logout: async () => {},
				checkAuth: async () => {},
			}}
		>
			{children}
		</AuthContext.Provider>
	);
}
