import { toast } from "sonner";
import type { ApiResponse, UserStateResponse, ProfileDTO, BannedWordDTO, ReportDTO } from "@/types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

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

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		const errorMessage = errorData.message || `HTTP error! status: ${response.status}`;
		toast.error(errorMessage);
		throw new Error(errorMessage);
	}

	return response.json();
}

export const queueAPI = {
	join: async () => {
		try {
			const result = await apiCall<{ message: string }>("/queue/join", {
				method: "POST",
			});
			toast.success("Đã tham gia hàng chờ thành công!");
			return result;
		} catch (error) {
			throw error;
		}
	},

	leave: async () => {
		try {
			const result = await apiCall<{ message: string }>("/queue/leave", {
				method: "POST",
			});
			toast.success("Đã rời khỏi hàng chờ!");
			return result;
		} catch (error) {
			throw error;
		}
	},
};

export const userAPI = {
	getState: async () => {
		return apiCall<UserStateResponse>("/user/state");
	},

	updateProfile: async (data: {
		age?: number | null;
		city?: string;
		is_male?: boolean;
		is_hidden?: boolean;
	}) => {
		try {
			const result = await apiCall<ProfileDTO>("/profile", {
				method: "PUT",
				body: JSON.stringify(data),
			});
			toast.success("Cập nhật thông tin thành công!");
			return result;
		} catch (error) {
			throw error;
		}
	},
};

export const authAPI = {
	logout: async () => {
		try {
			const result = await apiCall<{ message: string }>("/auth/logout", {
				method: "POST",
			});
			toast.success("Đăng xuất thành công!");
			return result;
		} catch (error) {
			throw error;
		}
	},

	getGoogleAuthUrl: () => {
		return `${API_BASE}/auth/google`;
	},
};

export const moderationAPI = {
	listWords: () => apiCall<BannedWordDTO[]>("/admin/words"),
	addWord: (word: string) =>
		apiCall<BannedWordDTO>("/admin/words", { method: "POST", body: JSON.stringify({ word }) }),
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
};

export const roomAPI = {
	leaveRoom: async () => {
		try {
			const result = await apiCall<{ success: boolean; message: string }>("/room/leave", {
				method: "POST",
			});
			toast.success("Đã rời phòng chat!");
			return result;
		} catch (error) {
			throw error;
		}
	},
};
