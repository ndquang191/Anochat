import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import { userAPI } from "@/lib/api";
import type { UserStateResponse } from "@/types";

export const USER_STATE_KEY = ["user-state"] as const;

export function useUserState(enabled: boolean = true) {
	return useQuery<UserStateResponse>({
		queryKey: USER_STATE_KEY,
		queryFn: async () => {
			const response = await userAPI.getState();
			if (!response.data) {
				throw new Error("No user state data");
			}
			return response.data;
		},
		enabled,
	});
}

export function useInvalidateUserState() {
	const queryClient = useQueryClient();
	return useCallback(() => {
		queryClient.invalidateQueries({ queryKey: USER_STATE_KEY });
	}, [queryClient]);
}
