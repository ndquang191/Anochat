"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { getWebSocketClient, ChatMessage, WebSocketMessage } from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { playMessageSound, playLeaveSound } from "@/hooks/use-sound-notification";

export interface UseWebSocketChatProps {
	userId: string;
	initialMessages?: ChatMessage[];
	onMatchFound?: (roomId: string) => void;
	onPartnerLeft?: () => void;
}

export function useWebSocketChat({ userId, initialMessages, onMatchFound, onPartnerLeft }: UseWebSocketChatProps) {
	const [messages, setMessages] = useState<ChatMessage[]>(initialMessages ?? []);
	const [isConnected, setIsConnected] = useState(false);
	const [roomId, setRoomId] = useState<string | null>(null);
	const [partnerLeft, setPartnerLeft] = useState(false);
	const wsClient = useRef(getWebSocketClient());
	const userIdRef = useRef(userId);
	const onMatchFoundRef = useRef(onMatchFound);
	const onPartnerLeftRef = useRef(onPartnerLeft);
	const { room } = useAuth();
	const invalidateUserState = useInvalidateUserState();
	const hasRehydratedRef = useRef(false);
	const hasJoinedRoomRef = useRef<string | null>(null);

	useEffect(() => {
		onMatchFoundRef.current = onMatchFound;
		onPartnerLeftRef.current = onPartnerLeft;
	}, [onMatchFound, onPartnerLeft]);

	useEffect(() => {
		userIdRef.current = userId;
	}, [userId]);

	useEffect(() => {
		if (room?.id && !roomId && !hasRehydratedRef.current) {
			setRoomId(room.id);
			setPartnerLeft(false);
			hasRehydratedRef.current = true;
		}
	}, [room, roomId]);

	useEffect(() => {
		if (!initialMessages?.length) return;
		setMessages((prev) => {
			const existingIds = new Set(prev.map((m) => m.id));
			const merged = [...prev, ...initialMessages.filter((m) => !existingIds.has(m.id))];
			return merged.sort((a, b) => a.created_at - b.created_at);
		});
	}, [initialMessages]);

	useEffect(() => {
		if (isConnected && roomId && hasJoinedRoomRef.current !== roomId) {
			const client = wsClient.current;
			client.send("join_room", { room_id: roomId });
			hasJoinedRoomRef.current = roomId;
		}
	}, [isConnected, roomId]);

	useEffect(() => {
		if (!userId) {
			return;
		}

		const client = wsClient.current;

		if (!client.isConnected()) {
			client
				.connect()
				.then(() => {
					setIsConnected(true);
				})
				.catch((error) => {
					console.error("Failed to connect to WebSocket:", error);
					setIsConnected(false);
				});
		} else {
			setIsConnected(true);
		}

		const handleConnected = () => {
			// Reset join tracking so join_room is re-sent after reconnection
			hasJoinedRoomRef.current = null;
			setIsConnected(true);
		};

		const handleDisconnected = () => {
			setIsConnected(false);
		};

		const handleMatchFound = (message: WebSocketMessage) => {
			const room_id = message.payload.room_id as string;
			setRoomId(room_id);
			setPartnerLeft(false);
			setMessages([]);
			invalidateUserState();
			if (onMatchFoundRef.current) {
				onMatchFoundRef.current(room_id);
			}
		};

		const handleRoomRejoined = (message: WebSocketMessage) => {
			const room_id = message.payload.room_id as string;
			setRoomId(room_id);
			setPartnerLeft(false);
			hasJoinedRoomRef.current = room_id;
		};

		const handleReceiveMessage = (message: WebSocketMessage) => {
			const chatMessage = message.payload as unknown as ChatMessage;
			if (chatMessage.sender_id !== userIdRef.current) {
				playMessageSound();
			}
			setMessages((prev) => {
				if (prev.some((m) => m.id === chatMessage.id)) {
					return prev;
				}
				return [...prev, chatMessage];
			});
		};

		const handlePartnerLeft = () => {
			setRoomId(null);
			setPartnerLeft(true);
			hasJoinedRoomRef.current = null;
			playLeaveSound();
			if (onPartnerLeftRef.current) {
				onPartnerLeftRef.current();
			}
			invalidateUserState();
		};

		const handleRoomLeft = () => {
			setRoomId(null);
			setPartnerLeft(false);
			setMessages([]);
			hasJoinedRoomRef.current = null;
			invalidateUserState();
		};

		client.on("connected", handleConnected);
		client.on("disconnected", handleDisconnected);
		client.on("match_found", handleMatchFound);
		client.on("room_rejoined", handleRoomRejoined);
		client.on("receive_message", handleReceiveMessage);
		client.on("partner_left", handlePartnerLeft);
		client.on("room_left", handleRoomLeft);

		return () => {
			client.off("connected", handleConnected);
			client.off("disconnected", handleDisconnected);
			client.off("match_found", handleMatchFound);
			client.off("room_rejoined", handleRoomRejoined);
			client.off("receive_message", handleReceiveMessage);
			client.off("partner_left", handlePartnerLeft);
			client.off("room_left", handleRoomLeft);
		};
	}, [userId, invalidateUserState]);

	const sendMessage = useCallback(
		(content: string) => {
			if (!roomId || !isConnected) {
				return;
			}

			const client = wsClient.current;
			client.send("send_message", {
				content,
			});

			// Optimistic update - show message immediately for the sender
			const optimisticMessage: ChatMessage = {
				id: crypto.randomUUID(),
				room_id: roomId,
				sender_id: userIdRef.current,
				content,
				created_at: Math.floor(Date.now() / 1000),
			};
			setMessages((prev) => [...prev, optimisticMessage]);
		},
		[roomId, isConnected]
	);

	const joinRoom = useCallback(
		(targetRoomId: string) => {
			if (!isConnected) {
				return;
			}

			const client = wsClient.current;
			client.send("join_room", {
				room_id: targetRoomId,
			});
		},
		[isConnected]
	);

	const leaveRoom = useCallback(() => {
		if (!roomId || !isConnected) {
			return;
		}

		const client = wsClient.current;
		client.send("leave_room", {
			room_id: roomId,
		});

		setRoomId(null);
		setPartnerLeft(false);
		setMessages([]);
		hasJoinedRoomRef.current = null;
		invalidateUserState();
	}, [roomId, isConnected, invalidateUserState]);

	return {
		messages,
		sendMessage,
		isConnected,
		roomId,
		partnerLeft,
		leaveRoom,
		joinRoom,
	};
}
