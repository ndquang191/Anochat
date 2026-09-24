"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import {
	getWebSocketClient,
	ChatMessage,
	WebSocketMessage,
} from "@/lib/websocket";
import { useAuth } from "@/contexts/auth";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import {
	playMessageSound,
	playLeaveSound,
} from "@/hooks/use-sound-notification";
import { roomAPI } from "@/lib/api";
import { prependUniqueMessages } from "@/lib/message-history";
import {
	reconcileAuthoritativeMessages,
	updateMessageDeliveryStatus,
} from "@/lib/message-delivery";
import { usePendingMessageAcks } from "@/hooks/use-pending-message-acks";

const MESSAGE_ACK_TIMEOUT_MS = 10000;

export interface UseWebSocketChatProps {
	userId: string;
	initialMessages?: ChatMessage[];
	initialNextCursor?: string | null;
	initialHasMore?: boolean;
	onMatchFound?: (roomId: string) => void;
	onPartnerLeft?: () => void;
	onMessageReceived?: () => void;
}

export function useWebSocketChat({
	userId,
	initialMessages,
	initialNextCursor,
	initialHasMore = false,
	onMatchFound,
	onPartnerLeft,
	onMessageReceived,
}: UseWebSocketChatProps) {
	const [messages, setMessages] = useState<ChatMessage[]>(
		initialMessages ?? [],
	);
	const [isConnected, setIsConnected] = useState(false);
	const [hasConnectedOnce, setHasConnectedOnce] = useState(false);
	const [roomId, setRoomId] = useState<string | null>(null);
	const [partnerLeft, setPartnerLeft] = useState(false);
	const [nextCursor, setNextCursor] = useState<string | null>(
		initialNextCursor ?? null,
	);
	const [hasMoreMessages, setHasMoreMessages] = useState(
		initialHasMore && !!initialNextCursor,
	);
	const [isLoadingOlder, setIsLoadingOlder] = useState(false);
	const [loadOlderError, setLoadOlderError] = useState(false);
	const wsClient = useRef(getWebSocketClient());
	const userIdRef = useRef(userId);
	const onMatchFoundRef = useRef(onMatchFound);
	const onPartnerLeftRef = useRef(onPartnerLeft);
	const onMessageReceivedRef = useRef(onMessageReceived);
	const { room } = useAuth();
	const invalidateUserState = useInvalidateUserState();
	const hasRehydratedRef = useRef(false);
	const hasJoinedRoomRef = useRef<string | null>(null);
	const paginationRoomRef = useRef<string | null>(null);
	const isLoadingOlderRef = useRef(false);
	const activeRoomIdRef = useRef<string | null>(null);
	const messageIdsRef = useRef(
		new Set((initialMessages ?? []).map((message) => message.id)),
	);
	const { clearPendingAck, clearAllPendingAcks, schedulePendingAck } =
		usePendingMessageAcks();

	useEffect(() => {
		onMatchFoundRef.current = onMatchFound;
		onPartnerLeftRef.current = onPartnerLeft;
		onMessageReceivedRef.current = onMessageReceived;
	}, [onMatchFound, onPartnerLeft, onMessageReceived]);

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
		for (const message of initialMessages) {
			clearPendingAck(message.id);
			messageIdsRef.current.add(message.id);
		}
		setMessages((prev) =>
			reconcileAuthoritativeMessages(prev, initialMessages),
		);
	}, [initialMessages, clearPendingAck]);

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
					setHasConnectedOnce(true);
				})
				.catch((error) => {
					console.error("Failed to connect to WebSocket:", error);
					setIsConnected(false);
				});
		} else {
			setIsConnected(true);
			setHasConnectedOnce(true);
		}

		const handleConnected = () => {
			// Reset join tracking so join_room is re-sent after reconnection
			hasJoinedRoomRef.current = null;
			setIsConnected(true);
			setHasConnectedOnce(true);
		};

		const handleDisconnected = () => {
			setIsConnected(false);
			clearAllPendingAcks();
			setMessages((current) =>
				current.map((message) =>
					message.status === "pending"
						? { ...message, status: "failed" }
						: message,
				),
			);
		};

		const handleMatchFound = (message: WebSocketMessage) => {
			const room_id = message.payload.room_id as string;
			clearAllPendingAcks();
			activeRoomIdRef.current = room_id;
			setRoomId(room_id);
			setPartnerLeft(false);
			setMessages([]);
			messageIdsRef.current.clear();
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
			const chatMessage = {
				...(message.payload as unknown as ChatMessage),
				status: "sent" as const,
			};
			if (messageIdsRef.current.has(chatMessage.id)) {
				return;
			}
			messageIdsRef.current.add(chatMessage.id);
			if (chatMessage.sender_id !== userIdRef.current) {
				playMessageSound();
				onMessageReceivedRef.current?.();
			}
			setMessages((prev) => [...prev, chatMessage]);
		};

		const handleMessageAck = (message: WebSocketMessage) => {
			const id = message.payload.id;
			if (typeof id !== "string") {
				return;
			}
			const createdAt =
				typeof message.payload.created_at === "number"
					? message.payload.created_at
					: undefined;
			clearPendingAck(id);
			setMessages((current) =>
				updateMessageDeliveryStatus(current, id, "sent", createdAt),
			);
		};

		const handleMessageFailed = (message: WebSocketMessage) => {
			const id = message.payload.id;
			if (typeof id !== "string") {
				return;
			}
			clearPendingAck(id);
			setMessages((current) =>
				updateMessageDeliveryStatus(current, id, "failed"),
			);
		};

		const handlePartnerLeft = () => {
			clearAllPendingAcks();
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
			clearAllPendingAcks();
			activeRoomIdRef.current = null;
			setRoomId(null);
			setPartnerLeft(false);
			setMessages([]);
			messageIdsRef.current.clear();
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
		client.on("message_ack", handleMessageAck);
		client.on("message_failed", handleMessageFailed);
		client.on("partner_left", handlePartnerLeft);
		client.on("room_left", handleRoomLeft);

		return () => {
			client.off("connected", handleConnected);
			client.off("disconnected", handleDisconnected);
			client.off("match_found", handleMatchFound);
			client.off("room_rejoined", handleRoomRejoined);
			client.off("receive_message", handleReceiveMessage);
			client.off("message_ack", handleMessageAck);
			client.off("message_failed", handleMessageFailed);
			client.off("partner_left", handlePartnerLeft);
			client.off("room_left", handleRoomLeft);
		};
	}, [userId, invalidateUserState, clearPendingAck, clearAllPendingAcks]);

	const reconnect = useCallback(async (): Promise<boolean> => {
		try {
			await wsClient.current.connect();
			const connected = wsClient.current.isConnected();
			setIsConnected(connected);
			if (connected) {
				setHasConnectedOnce(true);
			}
			return connected;
		} catch (error) {
			console.error("Failed to reconnect to WebSocket:", error);
			setIsConnected(false);
			return false;
		}
	}, []);

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
				prependUniqueMessages(current, page.messages, requestedRoomId),
			);
			setNextCursor(page.next_cursor ?? null);
			setHasMoreMessages(page.has_more && !!page.next_cursor);
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

			const messageId = crypto.randomUUID();
			const optimisticMessage: ChatMessage = {
				id: messageId,
				room_id: roomId,
				sender_id: userIdRef.current,
				content,
				created_at: Math.floor(Date.now() / 1000),
				status: "pending",
			};
			messageIdsRef.current.add(messageId);
			setMessages((prev) => [...prev, optimisticMessage]);

			const client = wsClient.current;
			const sent = client.send("send_message", {
				id: messageId,
				content,
			});
			if (!sent) {
				setMessages((current) =>
					updateMessageDeliveryStatus(current, messageId, "failed"),
				);
				return;
			}

			schedulePendingAck(messageId, MESSAGE_ACK_TIMEOUT_MS, () => {
				setMessages((current) =>
					updateMessageDeliveryStatus(current, messageId, "failed"),
				);
			});
		},
		[roomId, isConnected, schedulePendingAck],
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
		[isConnected],
	);

	return {
		messages,
		sendMessage,
		isConnected,
		hasConnectedOnce,
		reconnect,
		roomId,
		partnerLeft,
		hasMoreMessages,
		isLoadingOlder,
		loadOlderError,
		loadOlderMessages,
		joinRoom,
	};
}
