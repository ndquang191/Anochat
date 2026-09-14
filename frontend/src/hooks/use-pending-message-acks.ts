"use client";

import { useCallback, useEffect, useRef } from "react";

export function usePendingMessageAcks() {
	const timersRef = useRef<Map<string, number>>(new Map());

	const clear = useCallback((messageId: string) => {
		const timer = timersRef.current.get(messageId);
		if (timer === undefined) return;
		window.clearTimeout(timer);
		timersRef.current.delete(messageId);
	}, []);

	const clearAll = useCallback(() => {
		for (const timer of timersRef.current.values()) {
			window.clearTimeout(timer);
		}
		timersRef.current.clear();
	}, []);

	const schedule = useCallback(
		(messageId: string, timeoutMs: number, onTimeout: () => void) => {
			clear(messageId);
			const timer = window.setTimeout(() => {
				timersRef.current.delete(messageId);
				onTimeout();
			}, timeoutMs);
			timersRef.current.set(messageId, timer);
		},
		[clear],
	);

	useEffect(() => clearAll, [clearAll]);

	return {
		clearPendingAck: clear,
		clearAllPendingAcks: clearAll,
		schedulePendingAck: schedule,
	};
}
