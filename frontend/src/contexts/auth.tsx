"use client";

import { createContext, useContext, useCallback, ReactNode } from "react";
import { UserDTO, RoomDTO, MessageDTO } from "@/types";
import { setCookie, getCookie, deleteCookie } from "@/lib/cookies";
import { authAPI } from "@/lib/api";
import { resetWebSocketClient } from "@/lib/websocket";
import { useUserState, USER_STATE_KEY } from "@/hooks/queries/use-user-state";
import { queryClient } from "@/lib/query-client";

interface AuthContextType {
	isAuthenticated: boolean;
	user: UserDTO | null;
	room: RoomDTO | null;
	messages: MessageDTO[];
	loading: boolean;
	login: (user: UserDTO) => void;
	logout: () => Promise<void>;
	checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
	const hasCookie = !!getCookie("user_info");
	const { data, isLoading, isError } = useUserState(hasCookie);

	const login = useCallback((user: UserDTO) => {
		setCookie("user_info", JSON.stringify(user), 7);
		queryClient.invalidateQueries({ queryKey: USER_STATE_KEY });
	}, []);

	const logout = useCallback(async () => {
		try {
			await authAPI.logout();
		} catch {
		}

		resetWebSocketClient();
		deleteCookie("user_info");
		queryClient.removeQueries({ queryKey: USER_STATE_KEY });

		window.location.href = "/login";
	}, []);

	const checkAuth = useCallback(async () => {
		try {
			await queryClient.invalidateQueries({ queryKey: USER_STATE_KEY });
		} catch {
			await logout();
		}
	}, [logout]);

	const isAuthenticated = hasCookie && !isError && !!data;

	const value: AuthContextType = {
		isAuthenticated,
		user: data?.user ?? null,
		room: data?.room ?? null,
		messages: data?.messages ?? [],
		loading: hasCookie && isLoading,
		login,
		logout,
		checkAuth,
	};

	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
	const context = useContext(AuthContext);
	if (!context) {
		throw new Error("useAuth must be used within AuthProvider");
	}
	return context;
}
