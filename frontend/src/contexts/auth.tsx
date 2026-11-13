"use client";

import { createContext, useContext, useCallback, useState, useEffect, ReactNode } from "react";
import { UserDTO, RoomDTO, MessageDTO, MatchSettingsDTO } from "@/types";
import { setCookieMaxAge, getCookie, deleteCookie } from "@/lib/cookies";
import { authAPI, userAPI } from "@/lib/api";
import { resetWebSocketClient } from "@/lib/websocket";
import { useUserState, USER_STATE_KEY } from "@/hooks/queries/use-user-state";
import { queryClient } from "@/lib/query-client";
import { Loader2 } from "lucide-react";
import { useLanguage } from "@/contexts/theme";
import { usePWA } from "@/contexts/pwa";

interface AuthContextType {
	isAuthenticated: boolean;
	user: UserDTO | null;
	room: RoomDTO | null;
	rooms: RoomDTO[];
	maxActiveRooms: number;
	selectRoom: (roomId: string) => void;
	unreadByRoom: Record<string, number>;
	markRoomUnread: (roomId: string) => void;
	clearRoomUnread: (roomId: string) => void;
	messages: MessageDTO[];
	messagesNextCursor: string | null;
	messagesHasMore: boolean;
	inQueue: boolean;
	isBanned: boolean;
	banCount: number;
	reviewRequestCount: number;
	reviewRequested: boolean;
	matchSettings: MatchSettingsDTO | null;
	loading: boolean;
	hasError: boolean;
	login: () => Promise<void>;
	logout: () => Promise<void>;
	checkAuth: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
	const [mounted, setMounted] = useState(false);
	const [sessionPresent, setSessionPresent] = useState(false);
	const [isLoggingOut, setIsLoggingOut] = useState(false);
	const [activeRoomId, setActiveRoomId] = useState<string | null>(null);
	const [unreadByRoom, setUnreadByRoom] = useState<Record<string, number>>({});
	const { t } = useLanguage();
	const { unsubscribe, refreshPushConfig } = usePWA();
	useEffect(() => {
		const hasSession = !!getCookie("has_session");
		// Remove PII cookies created by older versions of the login flow.
		deleteCookie("user_info");
		deleteCookie("temp_user_data");
		setMounted(true);
		setSessionPresent(hasSession);
	}, []);

	const { data, isLoading, isError } = useUserState(
		mounted && sessionPresent
	);

	const login = useCallback(async () => {
		// This is only a frontend routing hint. Authentication still relies on
		// the backend's HttpOnly access and refresh token cookies.
		try {
			const response = await userAPI.getState();
			if (!response.data?.user) {
				throw new Error("No authenticated user state");
			}
			setCookieMaxAge("has_session", "1", response.data.session_max_age_seconds);
			queryClient.setQueryData(USER_STATE_KEY, response.data);
			setSessionPresent(true);
			await refreshPushConfig();
		} catch (error) {
			deleteCookie("has_session");
			queryClient.removeQueries({ queryKey: USER_STATE_KEY });
			setSessionPresent(false);
			throw error;
		}
	}, [refreshPushConfig]);

	const logout = useCallback(async () => {
		setIsLoggingOut(true);

		try {
			await unsubscribe();
		} catch {
		}
		try {
			await authAPI.logout();
		} catch {
		}

		resetWebSocketClient();
		deleteCookie("has_session");

		// Keep the current user state behind the loading overlay until the hard
		// navigation starts, so the chat never flashes an empty state.
		window.location.replace("/login");
	}, [unsubscribe]);

	const checkAuth = useCallback(async () => {
		try {
			await queryClient.invalidateQueries({ queryKey: USER_STATE_KEY });
		} catch {
			await logout();
		}
	}, [logout]);

	const isAuthenticated = sessionPresent && !isError && !!data?.user;
	const user = data?.user ?? null;
	const rooms = data?.rooms ?? (data?.room ? [data.room] : []);
	const room = rooms.find((item) => item.id === activeRoomId) ?? rooms[0] ?? null;
	const selectRoom = useCallback((roomId: string) => {
		setActiveRoomId(roomId);
		setUnreadByRoom((current) => {
			if (!current[roomId]) return current;
			const next = { ...current };
			delete next[roomId];
			return next;
		});
	}, []);
	const markRoomUnread = useCallback((roomId: string) => {
		setUnreadByRoom((current) => ({
			...current,
			[roomId]: (current[roomId] ?? 0) + 1,
		}));
	}, []);
	const clearRoomUnread = useCallback((roomId: string) => {
		setUnreadByRoom((current) => {
			if (!current[roomId]) return current;
			const next = { ...current };
			delete next[roomId];
			return next;
		});
	}, []);

	useEffect(() => {
		if (room && room.id !== activeRoomId) selectRoom(room.id);
		if (!room && activeRoomId) setActiveRoomId(null);
	}, [room, activeRoomId, selectRoom]);

	const value: AuthContextType = {
		isAuthenticated,
		user,
		room,
		rooms,
		maxActiveRooms: data?.max_active_rooms ?? 3,
		selectRoom,
		unreadByRoom,
		markRoomUnread,
		clearRoomUnread,
		messages: data?.messages ?? [],
		messagesNextCursor: data?.messages_next_cursor ?? null,
		messagesHasMore: data?.messages_has_more ?? false,
		inQueue: data?.in_queue ?? false,
		isBanned: data?.is_banned ?? false,
		banCount: data?.ban_count ?? 0,
		reviewRequestCount: data?.review_request_count ?? 0,
		reviewRequested: data?.review_requested ?? false,
		matchSettings: data?.match_settings ?? null,
		loading: !mounted || (sessionPresent && isLoading),
		hasError: sessionPresent && isError,
		login,
		logout,
		checkAuth,
	};

	return (
		<AuthContext.Provider value={value}>
			{children}
			{isLoggingOut && (
				<div
					className="fixed inset-0 z-[100] flex items-center justify-center bg-background/80 backdrop-blur-sm"
					role="status"
					aria-live="polite"
				>
					<div className="flex flex-col items-center gap-3">
						<Loader2 className="size-8 animate-spin text-primary" aria-hidden="true" />
						<p className="text-sm font-medium">{t("loggingOut")}</p>
					</div>
				</div>
			)}
		</AuthContext.Provider>
	);
}

export function useAuth() {
	const context = useContext(AuthContext);
	if (!context) {
		throw new Error("useAuth must be used within AuthProvider");
	}
	return context;
}
