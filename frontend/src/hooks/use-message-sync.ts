"use client";

import { useCallback, useEffect, useRef } from "react";
import { roomAPI } from "@/lib/api";
import type { MessageDTO } from "@/types";

const MESSAGE_SYNC_INTERVAL_MS = 5000;
const FOREGROUND_SYNC_STALE_MS = 30000;

interface UseMessageSyncProps {
	roomId: string | null;
	isConnected: boolean;
	onMessagesSynced: (messages: MessageDTO[]) => void;
}

/**
 * Keeps the latest message page in sync when WebSocket delivery is unavailable
 * or may have been interrupted while the app was in the background.
 */
export function useMessageSync({
	roomId,
	isConnected,
	onMessagesSynced,
}: UseMessageSyncProps) {
	const activeRoomIdRef = useRef(roomId);
	const isSyncingRef = useRef(false);
	const syncRequestedRef = useRef(false);
	const lastSyncAttemptAtRef = useRef(Date.now());
	const wasConnectedRef = useRef(isConnected);

	const syncLatestMessages = useCallback(async () => {
		syncRequestedRef.current = true;
		if (isSyncingRef.current) return;

		isSyncingRef.current = true;
		try {
			while (syncRequestedRef.current) {
				syncRequestedRef.current = false;
				const requestedRoomId = activeRoomIdRef.current;
				if (!requestedRoomId) continue;

				lastSyncAttemptAtRef.current = Date.now();
				try {
					const response = await roomAPI.getMessages(requestedRoomId);
					if (
						activeRoomIdRef.current === requestedRoomId &&
						response.data
					) {
						onMessagesSynced(response.data.messages);
					}
				} catch {
					// A failed fallback sync can safely retry on the next trigger without
					// disrupting WebSocket delivery or the chat UI.
				}
			}
		} finally {
			isSyncingRef.current = false;
		}
	}, [onMessagesSynced]);

	useEffect(() => {
		activeRoomIdRef.current = roomId;
		lastSyncAttemptAtRef.current = Date.now();
		if (roomId) {
			void syncLatestMessages();
		}
	}, [roomId, syncLatestMessages]);

	useEffect(() => {
		const connectionEstablished = isConnected && !wasConnectedRef.current;
		wasConnectedRef.current = isConnected;

		// Reconcile after every connection is established, including the first.
		// This closes the gap between loading /user/state and joining the room.
		if (connectionEstablished && roomId) {
			void syncLatestMessages();
		}
	}, [isConnected, roomId, syncLatestMessages]);

	useEffect(() => {
		if (!roomId || isConnected) return;

		const interval = window.setInterval(() => {
			if (document.visibilityState === "visible") {
				void syncLatestMessages();
			}
		}, MESSAGE_SYNC_INTERVAL_MS);

		return () => window.clearInterval(interval);
	}, [roomId, isConnected, syncLatestMessages]);

	useEffect(() => {
		if (!roomId) return;

		const syncWhenReturningToApp = () => {
			const isStale =
				Date.now() - lastSyncAttemptAtRef.current >=
				FOREGROUND_SYNC_STALE_MS;
			if (document.visibilityState === "visible" && isStale) {
				void syncLatestMessages();
			}
		};

		document.addEventListener("visibilitychange", syncWhenReturningToApp);
		window.addEventListener("focus", syncWhenReturningToApp);

		return () => {
			document.removeEventListener("visibilitychange", syncWhenReturningToApp);
			window.removeEventListener("focus", syncWhenReturningToApp);
		};
	}, [roomId, syncLatestMessages]);
}
