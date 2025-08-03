"use client";

import { useState, useCallback, useEffect } from "react";
import { queueAPI } from "@/lib/api";

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

export function useQueue() {
	const [isInQueue, setIsInQueue] = useState(false);
	const [queueStatus, setQueueStatus] = useState<QueueStatus | null>(null);
	const [queueStats, setQueueStats] = useState<QueueStats | null>(null);
	const [matchStats, setMatchStats] = useState<MatchStats | null>(null);
	const [isLoading, setIsLoading] = useState(false);

	// Join queue
	const joinQueue = useCallback(async (category: string = "polite") => {
		setIsLoading(true);

		try {
			const data = await queueAPI.join(category);
			setIsInQueue(true);
			setQueueStatus(data.data);
			return data;
		} catch (error) {
			// Error already handled by toast in API client
			console.error("Queue join error:", error);
		} finally {
			setIsLoading(false);
		}
	}, []);

	// Leave queue
	const leaveQueue = useCallback(async () => {
		setIsLoading(true);

		try {
			const data = await queueAPI.leave();
			setIsInQueue(false);
			setQueueStatus(null);
			return data;
		} catch (error) {
			// Error already handled by toast in API client
			console.error("Queue leave error:", error);
		} finally {
			setIsLoading(false);
		}
	}, []);

	// Get queue status
	const getQueueStatus = useCallback(async () => {
		try {
			const data = await queueAPI.getStatus();
			const status = data.data;

			setIsInQueue(status.isInQueue);
			setQueueStatus(status.isInQueue ? status : null);
			return status;
		} catch (err) {
			console.error("Lỗi khi lấy queue status:", err);
			return null;
		}
	}, []);

	// Get queue stats
	const getQueueStats = useCallback(async () => {
		try {
			const data = await queueAPI.getStats();
			setQueueStats(data.data);
			return data.data;
		} catch (err) {
			console.error("Lỗi khi lấy queue stats:", err);
			return null;
		}
	}, []);

	// Get match stats
	const getMatchStats = useCallback(async () => {
		try {
			const data = await queueAPI.getMatchStats();
			setMatchStats(data.data);
			return data.data;
		} catch (err) {
			console.error("Lỗi khi lấy match stats:", err);
			return null;
		}
	}, []);

	// Auto-refresh queue status when in queue
	useEffect(() => {
		if (isInQueue) {
			const interval = setInterval(() => {
				getQueueStatus();
			}, 5000); // Refresh every 5 seconds

			return () => clearInterval(interval);
		}
	}, [isInQueue, getQueueStatus]);

	// Initial load
	useEffect(() => {
		getQueueStatus();
	}, [getQueueStatus]);

	return {
		isInQueue,
		queueStatus,
		queueStats,
		matchStats,
		isLoading,
		joinQueue,
		leaveQueue,
		getQueueStatus,
		getQueueStats,
		getMatchStats,
	};
}
