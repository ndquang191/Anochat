"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { queueAPI } from "@/lib/api";
import type { ApiResponse, QueueStatus, QueueStats, MatchStats } from "@/types";

interface UseQueueReturn {
	isInQueue: boolean;
	queueStatus: QueueStatus | null;
	queueStats: QueueStats | null;
	matchStats: MatchStats | null;
	isLoading: boolean;
	joinQueue: (category?: string) => Promise<ApiResponse<QueueStatus> | undefined>;
	leaveQueue: () => Promise<void>;
	refreshQueueStatus: () => Promise<QueueStatus | null>;
	refreshQueueStats: () => Promise<QueueStats | null>;
	refreshMatchStats: () => Promise<MatchStats | null>;
}

// Constants
const POLLING_INTERVAL = 5000; // 5 seconds
const DEFAULT_CATEGORY = "polite";

// Custom hook for managing queue state
export function useQueue(): UseQueueReturn {
	// State management
	const [state, setState] = useState({
		isInQueue: false,
		queueStatus: null as QueueStatus | null,
		queueStats: null as QueueStats | null,
		matchStats: null as MatchStats | null,
		isLoading: false,
	});

	// Refs for cleanup and preventing memory leaks
	const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
	const isMountedRef = useRef(true);

	// Update state safely (only if component is mounted)
	const safeSetState = useCallback(
		(updates: Partial<typeof state>) => {
			if (isMountedRef.current) {
				setState((prev) => ({ ...prev, ...updates }));
			}
		},
		[] // No dependencies - uses functional setState
	);

	// Create optimistic queue status
	const createOptimisticStatus = useCallback(
		(category: string): QueueStatus => ({
			is_in_queue: true,
			position: 0,
			category,
			joined_at: new Date().toISOString(),
			expires_at: "", // No expiry time - handled by backend WebSocket
		}),
		[]
	);

	// Join queue
	const joinQueue = useCallback(
		async (category: string = DEFAULT_CATEGORY) => {
			safeSetState({ isLoading: true });

			// Optimistic update for better UX
			const optimisticStatus = createOptimisticStatus(category);
			safeSetState({
				isInQueue: true,
				queueStatus: optimisticStatus,
			});

			try {
				const response = await queueAPI.join(category);

				if (response?.data) {
					safeSetState({
						isInQueue: true,
						queueStatus: response.data,
						isLoading: false,
					});
				} else {
					// If no data, still set loading to false
					safeSetState({ isLoading: false });
				}

				return response;
			} catch (error) {
				console.error("Failed to join queue:", error);

				// Revert optimistic update on failure
				safeSetState({
					isInQueue: false,
					queueStatus: null,
					isLoading: false,
				});

				throw error;
			}
		},
		[safeSetState, createOptimisticStatus]
	);

	// Leave queue
	const leaveQueue = useCallback(async () => {
		safeSetState({ isLoading: true });

		try {
			const response = await queueAPI.leave();

			safeSetState({
				isInQueue: false,
				queueStatus: null,
				isLoading: false,
			});

			return response;
		} catch (error) {
			console.error("Failed to leave queue:", error);
			safeSetState({ isLoading: false });
			throw error;
		}
	}, [safeSetState]);

	// Refresh queue status
	const refreshQueueStatus = useCallback(async () => {
		try {
			const response = await queueAPI.getStatus();
			console.log("[useQueue] refreshQueueStatus response:", response);
			if (response.success && response.data) {
				console.log("[useQueue] Setting isInQueue:", response.data.is_in_queue);
				safeSetState({
					isInQueue: response.data.is_in_queue,
					queueStatus: response.data,
				});
				return response.data;
			} else {
				console.log("[useQueue] No data in response, setting isInQueue to false");
				safeSetState({
					isInQueue: false,
					queueStatus: null,
				});
			}
			return null;
		} catch (error) {
			console.error("[useQueue] Failed to refresh queue status:", error);
			return null;
		}
	}, [safeSetState]);

	// Refresh queue stats
	const refreshQueueStats = useCallback(async () => {
		try {
			const response = await queueAPI.getStats();
			const stats = response?.data;

			if (stats) {
				safeSetState({ queueStats: stats });
			}

			return stats;
		} catch (error) {
			console.error("Failed to refresh queue stats:", error);
			return null;
		}
	}, [safeSetState]);

	// Refresh match stats
	const refreshMatchStats = useCallback(async () => {
		try {
			const response = await queueAPI.getMatchStats();
			const stats = response?.data;

			if (stats) {
				safeSetState({ matchStats: stats });
			}

			return stats;
		} catch (error) {
			console.error("Failed to refresh match stats:", error);
			return null;
		}
	}, [safeSetState]);

	// Initialize queue status on mount (for F5 refresh scenarios)
	useEffect(() => {
		refreshQueueStatus();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []); // Only run once on mount

	// Setup polling when in queue
	useEffect(() => {
		// Clear any existing interval
		if (pollingIntervalRef.current) {
			clearInterval(pollingIntervalRef.current);
			pollingIntervalRef.current = null;
		}

		// Start polling if in queue
		if (state.isInQueue) {
			pollingIntervalRef.current = setInterval(() => {
				refreshQueueStatus();
			}, POLLING_INTERVAL);
		}

		// Cleanup on unmount or when leaving queue
		return () => {
			if (pollingIntervalRef.current) {
				clearInterval(pollingIntervalRef.current);
				pollingIntervalRef.current = null;
			}
		};
	}, [state.isInQueue, refreshQueueStatus]);

	// Cleanup on unmount
	useEffect(() => {
		return () => {
			isMountedRef.current = false;
			if (pollingIntervalRef.current) {
				clearInterval(pollingIntervalRef.current);
			}
		};
	}, []);

	return {
		isInQueue: state.isInQueue,
		queueStatus: state.queueStatus,
		queueStats: state.queueStats,
		matchStats: state.matchStats,
		isLoading: state.isLoading,
		joinQueue,
		leaveQueue,
		refreshQueueStatus,
		refreshQueueStats,
		refreshMatchStats,
	};
}

// Export hook return type for external use
export type { UseQueueReturn };
