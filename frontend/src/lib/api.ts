import { toast } from "sonner";
import type { ApiResponse, UserStateResponse, ProfileDTO, BannedWordDTO, ReportDTO, BannedUserDTO } from "@/types";
import { translateStored } from "@/lib/i18n";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

let refreshPromise: Promise<boolean> | null = null;

function clearAuthCookies() {
	const past = "Thu, 01 Jan 1970 00:00:00 UTC";
	document.cookie = `user_info=;expires=${past};path=/;`;
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

	const response = await fetch(`${API_BASE}${endpoint}`, config);

	if (response.status === 401) {
		const errorData = await response.json().catch(() => ({}));

		if (errorData.code === "account_suspended") {
			toast.error("Tài kho?n dã b? khóa");
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
			toast.error(retryMessage);
			throw new Error(retryMessage);
		}

		clearAuthCookies();
		window.location.href = "/login";
		return new Promise(() => {});
	}

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		const errorMessage = errorData.message || errorData.error || `HTTP error! status: ${response.status}`;
		toast.error(errorMessage);
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
	logout: async () => {
		const result = await apiCall<{ message: string }>("/auth/logout", { method: "POST" });
		toast.success(translateStored("logoutSuccess"));
		return result;
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
	listReports: () => apiCall<ReportDTO[]>("/admin/reports"),
	banUser: (userId: string) =>
		apiCall<void>(`/admin/users/${userId}/ban`, { method: "POST" }),
	createReport: (reportedUserId: string, roomId: string) =>
		apiCall<void>("/report", {
			method: "POST",
			body: JSON.stringify({ reported_user_id: reportedUserId, room_id: roomId }),
		}),
	getRoomMessages: (roomId: string) =>
		apiCall<{ id: string; sender_id: string; content: string; created_at: number }[]>(
			`/admin/rooms/${roomId}/messages`
		),
	listBannedUsers: () => apiCall<BannedUserDTO[]>("/admin/users/banned"),
	unbanUser: (userId: string) =>
		apiCall<void>(`/admin/users/${userId}/unban`, { method: "POST" }),
};

export const roomAPI = {
	leaveRoom: async () => {
		const result = await apiCall<{ success: boolean; message: string }>("/room/leave", { method: "POST" });
		toast.success(translateStored("leaveChatRoomSuccess"));
		return result;
	},
};
