"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { getWebSocketClient, ChatMessage, WebSocketMessage } from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { playMessageSound, playLeaveSound } from "@/hooks/use-sound-notification";
import { roomAPI } from "@/lib/api";
import { prependUniqueMessages } from "@/lib/message-history";

export interface UseWebSocketChatProps {
	userId: string;
	initialMessages?: ChatMessage[];
	initialNextCursor?: string | null;
	initialHasMore?: boolean;
	onMatchFound?: (roomId: string) => void;
	onPartnerLeft?: () => void;
}

export function useWebSocketChat({
	userId,
	initialMessages,
	initialNextCursor,
	initialHasMore = false,
	onMatchFound,
	onPartnerLeft,
}: UseWebSocketChatProps) {
	const [messages, setMessages] = useState<ChatMessage[]>(initialMessages ?? []);
	const [isConnected, setIsConnected] = useState(false);
	const [roomId, setRoomId] = useState<string | null>(null);
	const [partnerLeft, setPartnerLeft] = useState(false);
	const [nextCursor, setNextCursor] = useState<string | null>(initialNextCursor ?? null);
	const [hasMoreMessages, setHasMoreMessages] = useState(
		initialHasMore && !!initialNextCursor
	);
	const [isLoadingOlder, setIsLoadingOlder] = useState(false);
	const [loadOlderError, setLoadOlderError] = useState(false);
	const wsClient = useRef(getWebSocketClient());
	const userIdRef = useRef(userId);
	const onMatchFoundRef = useRef(onMatchFound);
	const onPartnerLeftRef = useRef(onPartnerLeft);
	const { room } = useAuth();
	const invalidateUserState = useInvalidateUserState();
	const hasRehydratedRef = useRef(false);
	const hasJoinedRoomRef = useRef<string | null>(null);
	const paginationRoomRef = useRef<string | null>(null);
	const isLoadingOlderRef = useRef(false);
	const activeRoomIdRef = useRef<string | null>(null);

	useEffect(() => {
		onMatchFoundRef.current = onMatchFound;
		onPartnerLeftRef.current = onPartnerLeft;
	}, [onMatchFound, onPartnerLeft]);

	useEffect(() => {
		userIdRef.current = userId;
	}, [userId]);

	useEffect(() => {
		if (room?.id && !roomId && !hasRehydratedRef.current) {
			activeRoomIdRef.current = room.id;
			setRoomId(room.id);
			setPartnerLeft(false);
			hasRehydratedRef.current = true;
		}
	}, [room, roomId]);

	useEffect(() => {
		if (!roomId) {
			paginationRoomRef.current = null;
			setNextCursor(null);
			setHasMoreMessages(false);
			setLoadOlderError(false);
			return;
		}
		if (room?.id !== roomId || paginationRoomRef.current === roomId) {
			return;
		}

		paginationRoomRef.current = roomId;
		setNextCursor(initialNextCursor ?? null);
		setHasMoreMessages(initialHasMore && !!initialNextCursor);
		setLoadOlderError(false);
	}, [room?.id, roomId, initialNextCursor, initialHasMore]);

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
			activeRoomIdRef.current = room_id;
			setRoomId(room_id);
			setPartnerLeft(false);
			setMessages([]);
			paginationRoomRef.current = null;
			setNextCursor(null);
			setHasMoreMessages(false);
			setLoadOlderError(false);
			invalidateUserState();
			if (onMatchFoundRef.current) {
				onMatchFoundRef.current(room_id);
			}
		};

		const handleRoomRejoined = (message: WebSocketMessage) => {
			const room_id = message.payload.room_id as string;
			activeRoomIdRef.current = room_id;
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
			activeRoomIdRef.current = null;
			setRoomId(null);
			setPartnerLeft(true);
			setNextCursor(null);
			setHasMoreMessages(false);
			hasJoinedRoomRef.current = null;
			playLeaveSound();
			if (onPartnerLeftRef.current) {
				onPartnerLeftRef.current();
			}
			invalidateUserState();
		};

		const handleRoomLeft = () => {
			activeRoomIdRef.current = null;
			setRoomId(null);
			setPartnerLeft(false);
			setMessages([]);
			setNextCursor(null);
			setHasMoreMessages(false);
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

	const loadOlderMessages = useCallback(async (): Promise<boolean> => {
		if (
			!roomId ||
			!nextCursor ||
			!hasMoreMessages ||
			isLoadingOlderRef.current
		) {
			return false;
		}

		isLoadingOlderRef.current = true;
		setIsLoadingOlder(true);
		setLoadOlderError(false);
		const requestedRoomId = roomId;

		try {
			const response = await roomAPI.getMessages(requestedRoomId, nextCursor);
			if (activeRoomIdRef.current !== requestedRoomId) {
				return false;
			}
			if (!response.data) {
				throw new Error("Missing message page data");
			}
			const page = response.data;

			setMessages((current) =>
				prependUniqueMessages(current, page.messages, requestedRoomId)
			);
			setNextCursor(page.next_cursor ?? null);
			setHasMoreMessages(
				page.has_more && !!page.next_cursor
			);
			return true;
		} catch (error) {
			if (activeRoomIdRef.current !== requestedRoomId) {
				return false;
			}
			console.error("Failed to load older messages:", error);
			setLoadOlderError(true);
			return false;
		} finally {
			isLoadingOlderRef.current = false;
			setIsLoadingOlder(false);
		}
	}, [roomId, nextCursor, hasMoreMessages]);

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

		activeRoomIdRef.current = null;
		setRoomId(null);
		setPartnerLeft(false);
		setMessages([]);
		setNextCursor(null);
		setHasMoreMessages(false);
		hasJoinedRoomRef.current = null;
		invalidateUserState();
	}, [roomId, isConnected, invalidateUserState]);

	return {
		messages,
		sendMessage,
		isConnected,
		roomId,
		partnerLeft,
		hasMoreMessages,
		isLoadingOlder,
		loadOlderError,
		loadOlderMessages,
		leaveRoom,
		joinRoom,
	};
}
