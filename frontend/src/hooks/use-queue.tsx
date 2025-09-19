"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { queueAPI } from "@/lib/api";
import { ApiResponse } from "@/types";

// Types
interface QueueStatus {
	isInQueue: boolean;
	position: number;
	category: string;
	joinedAt: string;
	expiresAt: string;
}

interface QueueStats {
	totalMale: number;
	totalFemale: number;
	estimatedWaitTime: string;
}

interface MatchStats {
	totalMatches: number;
	maleWaitTime: string;
	femaleWaitTime: string;
	lastMatchTime: string;
}

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
const QUEUE_EXPIRY_TIME = 5 * 60 * 1000; // 5 minutes

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
				console.log("🔧 safeSetState called with:", updates, "current state:", state);
				setState((prev) => {
					const newState = { ...prev, ...updates };
					console.log("🔄 State updated from:", prev, "to:", newState);
					return newState;
				});
			}
		},
		[state]
	);

	// Create optimistic queue status
	const createOptimisticStatus = useCallback(
		(category: string): QueueStatus => ({
			isInQueue: true,
			position: 0,
			category,
			joinedAt: new Date().toISOString(),
			expiresAt: new Date(Date.now() + QUEUE_EXPIRY_TIME).toISOString(),
		}),
		[]
	);

	// Join queue
	const joinQueue = useCallback(
		async (category: string = DEFAULT_CATEGORY) => {
			// Prevent duplicate join requests
			if (state.isInQueue || state.isLoading) {
				console.warn("Already in queue or loading");
				return;
			}

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
		[state.isInQueue, state.isLoading, safeSetState, createOptimisticStatus]
	);

	// Leave queue
	const leaveQueue = useCallback(async () => {
		if (!state.isInQueue || state.isLoading) {
			return;
		}

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
	}, [state.isInQueue, state.isLoading, safeSetState]);

	// Refresh queue status
	const refreshQueueStatus = useCallback(async () => {
		try {
			const response = await queueAPI.getStatus();
			if (response.success && response.data) {
				safeSetState({
					isInQueue: response.data.isInQueue,
					queueStatus: response.data,
				});
			}
		} catch (error) {
			console.error("Failed to refresh queue status:", error);
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

	// Setup polling when in queue
	useEffect(() => {
		console.log("🔄 Polling useEffect triggered");
		console.log("  - state.isInQueue =", state.isInQueue);

		// Clear any existing interval
		if (pollingIntervalRef.current) {
			console.log("🛑 Clearing existing polling interval");
			clearInterval(pollingIntervalRef.current);
			pollingIntervalRef.current = null;
		}

		// Start polling if in queue
		if (state.isInQueue) {
			pollingIntervalRef.current = setInterval(() => {
				console.log("📡 Polling: calling refreshQueueStatus...");
				refreshQueueStatus();
			}, POLLING_INTERVAL);
		} else {
			console.log("⏸️ Not in queue, no polling needed");
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

// Optional: Export types for external use
export type { QueueStatus, QueueStats, MatchStats, UseQueueReturn };
