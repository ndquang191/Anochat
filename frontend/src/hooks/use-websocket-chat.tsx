"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { getWebSocketClient, ChatMessage, WebSocketMessage } from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";

export interface UseWebSocketChatProps {
	userId: string;
	onMatchFound?: (roomId: string) => void;
	onPartnerLeft?: () => void;
}

export function useWebSocketChat({ userId, onMatchFound, onPartnerLeft }: UseWebSocketChatProps) {
	const [messages, setMessages] = useState<ChatMessage[]>([]);
	const [isConnected, setIsConnected] = useState(false);
	const [roomId, setRoomId] = useState<string | null>(null);
	const [isPartnerTyping, setIsPartnerTyping] = useState(false);
	const wsClient = useRef(getWebSocketClient());
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
		if (room?.id && !roomId && !hasRehydratedRef.current) {
			setRoomId(room.id);
			hasRehydratedRef.current = true;
		}
	}, [room, roomId]);

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
			invalidateUserState();

			if (onMatchFoundRef.current) {
				onMatchFoundRef.current(room_id);
			}
		};

		const handleRoomJoined = () => {
		};

		const handleReceiveMessage = (message: WebSocketMessage) => {
			const chatMessage = message.payload as unknown as ChatMessage;
			setMessages((prev) => {
				if (prev.some((m) => m.id === chatMessage.id)) {
					return prev;
				}
				return [...prev, chatMessage];
			});
		};

		const handlePartnerLeft = () => {
			setRoomId(null);
			setMessages([]);
			hasJoinedRoomRef.current = null;
			if (onPartnerLeftRef.current) {
				onPartnerLeftRef.current();
			}
			invalidateUserState();
		};

		const handlePartnerTyping = (message: WebSocketMessage) => {
			const is_typing = message.payload.is_typing as boolean;
			setIsPartnerTyping(is_typing);

			if (is_typing) {
				setTimeout(() => setIsPartnerTyping(false), 3000);
			}
		};

		const handleRoomLeft = () => {
			setRoomId(null);
			setMessages([]);
			hasJoinedRoomRef.current = null;
			invalidateUserState();
		};

		client.on("connected", handleConnected);
		client.on("disconnected", handleDisconnected);
		client.on("match_found", handleMatchFound);
		client.on("room_joined", handleRoomJoined);
		client.on("receive_message", handleReceiveMessage);
		client.on("partner_left", handlePartnerLeft);
		client.on("partner_typing", handlePartnerTyping);
		client.on("room_left", handleRoomLeft);

		return () => {
			client.off("connected", handleConnected);
			client.off("disconnected", handleDisconnected);
			client.off("match_found", handleMatchFound);
			client.off("room_joined", handleRoomJoined);
			client.off("receive_message", handleReceiveMessage);
			client.off("partner_left", handlePartnerLeft);
			client.off("partner_typing", handlePartnerTyping);
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
				sender_id: userId,
				content,
				created_at: Math.floor(Date.now() / 1000),
			};
			setMessages((prev) => [...prev, optimisticMessage]);
		},
		[roomId, isConnected, userId]
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
		setMessages([]);
		hasJoinedRoomRef.current = null;
		invalidateUserState();
	}, [roomId, isConnected, invalidateUserState]);

	const sendTypingIndicator = useCallback(
		(isTyping: boolean) => {
			if (!roomId || !isConnected) {
				return;
			}

			const client = wsClient.current;
			client.send("typing", {
				is_typing: isTyping,
			});
		},
		[roomId, isConnected]
	);

	return {
		messages,
		sendMessage,
		isConnected,
		roomId,
		leaveRoom,
		joinRoom,
		isPartnerTyping,
		sendTypingIndicator,
	};
}
