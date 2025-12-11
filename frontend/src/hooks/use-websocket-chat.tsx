"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { getWebSocketClient, ChatMessage, WebSocketMessage } from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";

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
	const { checkAuth } = useAuth();

	// Keep refs up to date
	useEffect(() => {
		onMatchFoundRef.current = onMatchFound;
		onPartnerLeftRef.current = onPartnerLeft;
	}, [onMatchFound, onPartnerLeft]);

	useEffect(() => {
		// Only connect if user is authenticated
		if (!userId) {
			console.log("No user ID, skipping WebSocket connection");
			return;
		}

		const client = wsClient.current;

		// Connect to WebSocket only if not already connected
		if (!client.isConnected()) {
			client
				.connect()
				.then(() => {
					setIsConnected(true);
					console.log("Connected to WebSocket");
				})
				.catch((error) => {
					console.error("Failed to connect to WebSocket:", error);
					setIsConnected(false);
				});
		} else {
			// Already connected, just update state
			setIsConnected(true);
		}

		// Handle connected event
		const handleConnected = (message: WebSocketMessage) => {
			console.log("WebSocket connection confirmed:", message);
			setIsConnected(true);
		};

		// Handle match found
		const handleMatchFound = (message: WebSocketMessage) => {
			const { room_id, category } = message.payload;
			console.log("Match found:", room_id, category);
			setRoomId(room_id);

			// Join the room automatically
			client.send("join_room", { room_id });

			if (onMatchFoundRef.current) {
				onMatchFoundRef.current(room_id);
			}
		};

		// Handle room joined
		const handleRoomJoined = (message: WebSocketMessage) => {
			const { room_id } = message.payload;
			console.log("Joined room:", room_id);
		};

		// Handle incoming messages
		const handleReceiveMessage = (message: WebSocketMessage) => {
			const chatMessage = message.payload as ChatMessage;
			setMessages((prev) => {
				// Avoid duplicates
				if (prev.some((m) => m.id === chatMessage.id)) {
					return prev;
				}
				return [...prev, chatMessage];
			});
		};

		// Handle partner left
		const handlePartnerLeft = (message: WebSocketMessage) => {
			console.log("Partner left the room");
			setRoomId(null);
			setMessages([]);
			if (onPartnerLeftRef.current) {
				onPartnerLeftRef.current();
			}

			// Refresh user state to clear active room
			setTimeout(() => {
				checkAuth().catch(console.error);
			}, 300);
		};

		// Handle partner typing
		const handlePartnerTyping = (message: WebSocketMessage) => {
			const { is_typing } = message.payload;
			setIsPartnerTyping(is_typing);

			// Auto-hide typing indicator after 3 seconds
			if (is_typing) {
				setTimeout(() => setIsPartnerTyping(false), 3000);
			}
		};

		// Handle room left confirmation
		const handleRoomLeft = (message: WebSocketMessage) => {
			console.log("Left room successfully");
			setRoomId(null);
			setMessages([]);

			// Refresh user state to clear active room
			setTimeout(() => {
				checkAuth().catch(console.error);
			}, 300);
		};

		// Register event handlers
		client.on("connected", handleConnected);
		client.on("match_found", handleMatchFound);
		client.on("room_joined", handleRoomJoined);
		client.on("receive_message", handleReceiveMessage);
		client.on("partner_left", handlePartnerLeft);
		client.on("partner_typing", handlePartnerTyping);
		client.on("room_left", handleRoomLeft);

		// Cleanup on unmount
		return () => {
			client.off("connected", handleConnected);
			client.off("match_found", handleMatchFound);
			client.off("room_joined", handleRoomJoined);
			client.off("receive_message", handleReceiveMessage);
			client.off("partner_left", handlePartnerLeft);
			client.off("partner_typing", handlePartnerTyping);
			client.off("room_left", handleRoomLeft);
		};
	}, [userId]);

	const sendMessage = useCallback(
		(content: string) => {
			if (!roomId || !isConnected) {
				console.warn("Cannot send message: not in a room or not connected");
				return;
			}

			const client = wsClient.current;
			client.send("send_message", {
				content,
			});
		},
		[roomId, isConnected]
	);

	const joinRoom = useCallback(
		(targetRoomId: string) => {
			if (!isConnected) {
				console.warn("Cannot join room: not connected");
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
			console.warn("Cannot leave room: not in a room or not connected");
			return;
		}

		const client = wsClient.current;
		client.send("leave_room", {
			room_id: roomId,
		});

		setRoomId(null);
		setMessages([]);

		// Refresh user state to clear active room
		setTimeout(() => {
			checkAuth().catch(console.error);
		}, 500);
	}, [roomId, isConnected, checkAuth]);

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
