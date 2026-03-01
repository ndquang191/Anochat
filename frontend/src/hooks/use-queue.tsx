"use client";

import { useState, useCallback } from "react";
import { queueAPI } from "@/lib/api";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";

export function useQueue() {
	const [isLoading, setIsLoading] = useState(false);
	const invalidateUserState = useInvalidateUserState();

	const joinQueue = useCallback(async () => {
		setIsLoading(true);
		try {
			const result = await queueAPI.join();
			invalidateUserState();
			return result;
		} finally {
			setIsLoading(false);
		}
	}, [invalidateUserState]);

	const leaveQueue = useCallback(async () => {
		setIsLoading(true);
		try {
			const result = await queueAPI.leave();
			invalidateUserState();
			return result;
		} finally {
			setIsLoading(false);
		}
	}, [invalidateUserState]);

	return { joinQueue, leaveQueue, isLoading };
}
