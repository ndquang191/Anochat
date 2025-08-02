"use client";

import { useEffect, useState, useCallback } from "react";
import socket from "@/lib/socket";

interface UseSocketChatProps {
	username: string;
	gender?: string;
	category?: string;
}

export interface ChatMessage {
	id: string;
	content: string;
	user: {
		name: string;
	};
	createdAt: string;
}

export function useSocketChat({ username, gender = "unknown", category = "general" }: UseSocketChatProps) {
	const [messages, setMessages] = useState<ChatMessage[]>([]);
	const [isConnected, setIsConnected] = useState(false);
	const [roomId, setRoomId] = useState<string | null>(null);
	const [status, setStatus] = useState<string>("");

	useEffect(() => {
		// Lấy roomId từ localStorage khi component mount trên client
		const storedRoomId =
			typeof window !== "undefined" ? localStorage.getItem(`chat_room_${username}`) : null;
		if (storedRoomId) {
			setRoomId(storedRoomId);
		}

		socket.connect();

		socket.on("partner_found", ({ roomId, partner, messages = [] }) => {
			setRoomId(roomId);
			setStatus(`Đã kết nối với ${partner}`);
			setMessages(messages);
			if (typeof window !== "undefined") {
				localStorage.setItem(`chat_room_${username}`, roomId);
			}
		});

		socket.on("waiting", ({ message }) => {
			setStatus(message);
		});

		socket.on("error", ({ message }) => {
			setStatus(message);
			if (message.includes("Phòng chat không tồn tại")) {
				setRoomId(null);
				setMessages([]);
				if (typeof window !== "undefined") {
					localStorage.removeItem(`chat_room_${username}`);
				}
			}
		});

		socket.on("partner_disconnected", ({ message }) => {
			setStatus(message);
		});

		socket.on("partner_reconnected", ({ message }) => {
			setStatus(message);
		});

		socket.on("partner_left", ({ message }) => {
			setStatus(message);
			setRoomId(null);
			setMessages([]);
			if (typeof window !== "undefined") {
				localStorage.removeItem(`chat_room_${username}`);
			}
		});

		socket.on("connect", () => {
			setIsConnected(true);
			if (storedRoomId) {
				socket.emit("rejoin_room", { roomId: storedRoomId, username });
			}
		});

		socket.on("message", (message: ChatMessage) => {
			setMessages((prev) => {
				if (!prev.some((m) => m.id === message.id)) {
					return [...prev, message];
				}
				return prev;
			});
		});

		socket.on("rejoin_failed", ({ message }) => {
			setStatus(message);
			setRoomId(null);
			setMessages([]);
			if (typeof window !== "undefined") {
				localStorage.removeItem(`chat_room_${username}`);
			}
		});

		return () => {
			socket.off("partner_found");
			socket.off("waiting");
			socket.off("error");
			socket.off("partner_disconnected");
			socket.off("partner_reconnected");
			socket.off("partner_left");
			socket.off("connect");
			socket.off("message");
			socket.off("rejoin_failed");
			socket.disconnect();
		};
	}, [username]);

	const sendMessage = useCallback(
		(content: string) => {
			if (!roomId) return;
			const message: ChatMessage = {
				id: crypto.randomUUID(),
				content,
				user: { name: username },
				createdAt: new Date().toISOString(),
			};
			socket.emit("send_message", { roomId, ...message });
		},
		[roomId, username]
	);

	const findPartner = useCallback(() => {
		if (isConnected && !roomId) {
			socket.emit("find_partner", { username, gender, category });
			setStatus("Đang tìm partner...");
		}
	}, [isConnected, roomId, username, gender, category]);

	const leaveRoom = useCallback(() => {
		if (isConnected && roomId) {
			socket.emit("leave_room", { roomId, username });
			setRoomId(null);
			setMessages([]);
			setStatus("Bạn đã rời phòng");
			if (typeof window !== "undefined") {
				localStorage.removeItem(`chat_room_${username}`);
			}
		}
	}, [isConnected, roomId, username]);

	return { messages, sendMessage, isConnected, roomId, findPartner, leaveRoom, status };
}
