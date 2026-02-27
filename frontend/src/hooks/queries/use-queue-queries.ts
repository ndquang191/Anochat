import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { queueAPI } from "@/lib/api";
import type { QueueStatus, QueueStats, MatchStats } from "@/types";

export const QUEUE_STATUS_KEY = ["queue-status"] as const;
export const QUEUE_STATS_KEY = ["queue-stats"] as const;
export const MATCH_STATS_KEY = ["match-stats"] as const;

export function useQueueStatus(isInQueue: boolean) {
	return useQuery<QueueStatus | null>({
		queryKey: QUEUE_STATUS_KEY,
		queryFn: async () => {
			const response = await queueAPI.getStatus();
			return response.data ?? null;
		},
		refetchInterval: isInQueue ? 5000 : false,
	});
}

export function useQueueStats() {
	return useQuery<QueueStats | null>({
		queryKey: QUEUE_STATS_KEY,
		queryFn: async () => {
			const response = await queueAPI.getStats();
			return response.data ?? null;
		},
	});
}

export function useMatchStats() {
	return useQuery<MatchStats | null>({
		queryKey: MATCH_STATS_KEY,
		queryFn: async () => {
			const response = await queueAPI.getMatchStats();
			return response.data ?? null;
		},
	});
}

export function useJoinQueue() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (category: string = "polite") => queueAPI.join(category),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: QUEUE_STATUS_KEY });
		},
	});
}

export function useLeaveQueue() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: () => queueAPI.leave(),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: QUEUE_STATUS_KEY });
		},
	});
}
