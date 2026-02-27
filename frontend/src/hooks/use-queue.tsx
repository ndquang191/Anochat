"use client";

import { useState, useCallback } from "react";
import { queueAPI } from "@/lib/api";
import {
	useQueueStatus,
	useQueueStats,
	useMatchStats,
	QUEUE_STATUS_KEY,
} from "@/hooks/queries/use-queue-queries";
import { queryClient } from "@/lib/query-client";
import type { ApiResponse, QueueStatus, QueueStats, MatchStats } from "@/types";

interface UseQueueReturn {
	isInQueue: boolean;
	queueStatus: QueueStatus | null;
	queueStats: QueueStats | null;
	matchStats: MatchStats | null;
	isLoading: boolean;
	joinQueue: (category?: string) => Promise<ApiResponse<QueueStatus> | undefined>;
	leaveQueue: () => Promise<ApiResponse<{ message: string }> | void>;
	refreshQueueStatus: () => Promise<QueueStatus | null | undefined>;
	refreshQueueStats: () => Promise<QueueStats | null | undefined>;
	refreshMatchStats: () => Promise<MatchStats | null | undefined>;
}

const DEFAULT_CATEGORY = "polite";

export function useQueue(): UseQueueReturn {
	const [isInQueue, setIsInQueue] = useState(false);
	const [isLoading, setIsLoading] = useState(false);

	const { data: queueStatus } = useQueueStatus(isInQueue);
	const { data: queueStats } = useQueueStats();
	const { data: matchStats } = useMatchStats();

	const syncQueueState = useCallback((status: QueueStatus | null | undefined) => {
		setIsInQueue(!!status?.is_in_queue);
	}, []);

	const joinQueue = useCallback(
		async (category: string = DEFAULT_CATEGORY) => {
			setIsLoading(true);
			setIsInQueue(true);

			try {
				const response = await queueAPI.join(category);
				syncQueueState(response?.data);
				queryClient.invalidateQueries({ queryKey: QUEUE_STATUS_KEY });
				setIsLoading(false);
				return response;
			} catch (error) {
				setIsInQueue(false);
				setIsLoading(false);
				throw error;
			}
		},
		[syncQueueState]
	);

	const leaveQueue = useCallback(async () => {
		setIsLoading(true);
		try {
			const response = await queueAPI.leave();
			setIsInQueue(false);
			queryClient.invalidateQueries({ queryKey: QUEUE_STATUS_KEY });
			setIsLoading(false);
			return response;
		} catch (error) {
			setIsLoading(false);
			throw error;
		}
	}, []);

	const refreshQueueStatus = useCallback(async () => {
		try {
			const response = await queueAPI.getStatus();
			const status = response.data ?? null;
			syncQueueState(status);
			return status;
		} catch {
			return null;
		}
	}, [syncQueueState]);

	const refreshQueueStats = useCallback(async () => {
		try {
			const response = await queueAPI.getStats();
			return response.data ?? null;
		} catch {
			return null;
		}
	}, []);

	const refreshMatchStats = useCallback(async () => {
		try {
			const response = await queueAPI.getMatchStats();
			return response.data ?? null;
		} catch {
			return null;
		}
	}, []);

	return {
		isInQueue,
		queueStatus: queueStatus ?? null,
		queueStats: queueStats ?? null,
		matchStats: matchStats ?? null,
		isLoading,
		joinQueue,
		leaveQueue,
		refreshQueueStatus,
		refreshQueueStats,
		refreshMatchStats,
	};
}

export type { UseQueueReturn };
