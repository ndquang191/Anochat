import { toast } from "sonner";
import type { AdminOverviewDTO, ApiResponse, UserStateResponse, ProfileDTO, BannedWordDTO, ReportGroupPageDTO, BannedUserPageDTO, MessagePageDTO } from "@/types";
import { translateStored } from "@/lib/i18n";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const API_HEALTH_CHECK_TIMEOUT_MS = 8_000;

let refreshPromise: Promise<boolean> | null = null;

function clearAuthCookies() {
	const past = "Thu, 01 Jan 1970 00:00:00 UTC";
	document.cookie = `has_session=;expires=${past};path=/;`;
}

function doRefresh(): Promise<boolean> {
	if (!refreshPromise) {
		refreshPromise = fetch(`${API_BASE}/auth/refresh`, {
			method: "POST",
			credentials: "include",
		})
			.then((r) => r.ok)
			.catch(() => false)
			.finally(() => {
				refreshPromise = null;
			});
	}
	return refreshPromise;
}

async function apiCall<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
	const config: RequestInit = {
		...options,
		credentials: "include",
		headers: {
			"Content-Type": "application/json",
			...options.headers,
		},
	};

	let response: Response;
	try {
		response = await fetch(`${API_BASE}${endpoint}`, config);
	} catch {
		throw new Error(translateStored("apiUnavailable"));
	}

	if (response.status === 401) {
		const errorData = await response.json().catch(() => ({}));

		if (errorData.code === "account_suspended") {
			toast.error(translateStored("accountSuspended"));
			clearAuthCookies();
			window.location.href = "/login";
			return new Promise(() => {});
		}

		const refreshed = await doRefresh();

		if (refreshed) {
			const retryResponse = await fetch(`${API_BASE}${endpoint}`, config);

			if (retryResponse.ok) {
				return retryResponse.json();
			}

			if (retryResponse.status === 401) {
				clearAuthCookies();
				window.location.href = "/login";
				return new Promise(() => {});
			}

			const retryError = await retryResponse.json().catch(() => ({}));
			const retryMessage = retryError.message || `HTTP error! status: ${retryResponse.status}`;
			throw new Error(retryMessage);
		}

		clearAuthCookies();
		window.location.href = "/login";
		return new Promise(() => {});
	}

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		const errorMessage = errorData.message || errorData.error || `HTTP error! status: ${response.status}`;
		throw new Error(errorMessage);
	}

	return response.json();
}

export const queueAPI = {
	join: async () => {
		const result = await apiCall<{ message: string }>("/queue/join", { method: "POST" });
		toast.success(translateStored("joinQueueSuccess"));
		return result;
	},

	leave: async () => {
		const result = await apiCall<{ message: string }>("/queue/leave", { method: "POST" });
		toast.success(translateStored("leaveQueueSuccess"));
		return result;
	},
};

export const userAPI = {
	getState: async () => {
		return apiCall<UserStateResponse>("/user/state");
	},

	requestBanReview: () =>
		apiCall<void>("/user/ban-review", { method: "POST" }),

	updateProfile: async (data: {
		nickname?: string | null;
		age?: number | null;
		is_male?: boolean | null;
		is_hidden?: boolean;
	}) => {
		const result = await apiCall<ProfileDTO>("/profile", {
			method: "PUT",
			body: JSON.stringify(data),
		});
		toast.success(translateStored("profileUpdated"));
		return result;
	},
};

export const authAPI = {
	devLogin: (user: "a" | "b") => apiCall<void>("/auth/dev", { method: "POST", body: JSON.stringify({ user }) }),
	logout: async () => {
		const result = await apiCall<{ message: string }>("/auth/logout", { method: "POST" });
		toast.success(translateStored("logoutSuccess"));
		return result;
	},

	assertAvailable: async () => {
		const controller = new AbortController();
		const timeout = window.setTimeout(
			() => controller.abort(),
			API_HEALTH_CHECK_TIMEOUT_MS
		);

		try {
			const response = await fetch(`${API_BASE}/healthz`, {
				credentials: "include",
				signal: controller.signal,
			});
			if (!response.ok) {
				throw new Error("API health check failed");
			}
		} catch {
			throw new Error(translateStored("apiUnavailable"));
		} finally {
			window.clearTimeout(timeout);
		}
	},

	getGoogleAuthUrl: () => {
		return `${API_BASE}/auth/google`;
	},
};

export const moderationAPI = {
	listWords: () => apiCall<BannedWordDTO[]>("/admin/words"),
	addWord: (word: string, category?: string) =>
		apiCall<BannedWordDTO>("/admin/words", { method: "POST", body: JSON.stringify({ word, category: category || "General" }) }),
	updateWord: (id: string, word: string, category: string) =>
		apiCall<void>(`/admin/words/${id}`, { method: "PUT", body: JSON.stringify({ word, category }) }),
	deleteWord: (id: string) =>
		apiCall<void>(`/admin/words/${id}`, { method: "DELETE" }),
	listReports: (options: { before?: string; query?: string; limit?: number } = {}) => {
		const params = new URLSearchParams({
			status: "pending",
			limit: String(options.limit ?? 20),
		});
		if (options.before) params.set("before", options.before);
		if (options.query) params.set("query", options.query);
		return apiCall<ReportGroupPageDTO>(`/admin/reports?${params.toString()}`);
	},
	banUser: (userId: string) =>
		apiCall<void>(`/admin/users/${userId}/ban`, { method: "POST" }),
	createReport: (reportedUserId: string, roomId: string) =>
		apiCall<void>("/report", {
			method: "POST",
			body: JSON.stringify({ reported_user_id: reportedUserId, room_id: roomId }),
		}),
	getReportMessages: (reportId: string) =>
		apiCall<{ id: string; sender_id: string; content: string; created_at: number }[]>(
			`/admin/reports/${reportId}/messages`
		),
	listBannedUsers: (options: { before?: string; query?: string; limit?: number } = {}) => {
		const params = new URLSearchParams({ limit: String(options.limit ?? 20) });
		if (options.before) params.set("before", options.before);
		if (options.query) params.set("query", options.query);
		return apiCall<BannedUserPageDTO>(`/admin/users/banned?${params.toString()}`);
	},
	unbanUser: (userId: string) =>
		apiCall<void>(`/admin/users/${userId}/unban`, { method: "POST" }),
};

export const adminAPI = {
	getOverview: () => apiCall<AdminOverviewDTO>("/admin/overview"),
};

export const roomAPI = {
	getMessages: (roomId: string, before?: string, limit: number = 50) => {
		const params = new URLSearchParams({ limit: String(limit) });
		if (before) params.set("before", before);
		return apiCall<MessagePageDTO>(`/rooms/${roomId}/messages?${params.toString()}`);
	},
	leaveRoom: async () => {
		const result = await apiCall<{ success: boolean; message: string }>("/room/leave", { method: "POST" });
		toast.success(translateStored("leaveChatRoomSuccess"));
		return result;
	},
};
