import { toast } from "sonner";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Type definitions for API responses
interface ApiResponse<T> {
	success: boolean;
	message: string;
	data: T;
}

interface QueueStatus {
	is_in_queue: boolean;
	position: number;
	category: string;
	joined_at: string;
	expires_at: string;
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

interface UserState {
	id: string;
	email?: string;
	name?: string;
	avatar_url?: string;
	is_active: boolean;
	is_deleted: boolean;
	created_at: string;
	profile?: {
		user_id: string;
		is_male?: boolean;
		age?: number;
		city?: string;
		is_hidden: boolean;
		updated_at: string;
	};
}

interface UserProfile {
	id: string;
	email?: string;
	name?: string;
	avatar_url?: string;
	is_active: boolean;
	is_deleted: boolean;
	created_at: string;
	profile?: {
		user_id: string;
		is_male?: boolean;
		age?: number;
		city?: string;
		is_hidden: boolean;
		updated_at: string;
	};
}

// Generic API call function
async function apiCall<T>(endpoint: string, options: RequestInit = {}): Promise<ApiResponse<T>> {
	// Token is handled via HTTP-only cookies, so we don't need to manually add it
	const config: RequestInit = {
		...options,
		credentials: "include",
		headers: {
			"Content-Type": "application/json",
			...options.headers,
		},
	};

	console.log(`[API] Calling: ${API_BASE}${endpoint}`);
	const response = await fetch(`${API_BASE}${endpoint}`, config);
	console.log(`[API] Response status: ${response.status}`);

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		const errorMessage = errorData.message || `HTTP error! status: ${response.status}`;
		console.error(`[API] Error: ${errorMessage}`);
		toast.error(errorMessage);
		throw new Error(errorMessage);
	}

	return response.json();
}

// Queue API functions
export const queueAPI = {
	// Join queue
	join: async (category: string = "polite") => {
		try {
			const result = await apiCall<QueueStatus>("/queue/join", {
				method: "POST",
				body: JSON.stringify({ category }),
			});
			toast.success("Đã tham gia hàng chờ thành công!");
			return result;
		} catch (error) {
			// Error already shown by apiCall
			throw error;
		}
	},

	// Leave queue
	leave: async () => {
		try {
			const result = await apiCall<{ message: string }>("/queue/leave", {
				method: "POST",
			});
			toast.success("Đã rời khỏi hàng chờ!");
			return result;
		} catch (error) {
			// Error already shown by apiCall
			throw error;
		}
	},

	// Get queue status
	getStatus: async () => {
		return apiCall<QueueStatus>("/queue/status");
	},

	// Get queue stats
	getStats: async () => {
		return apiCall<QueueStats>("/queue/stats");
	},

	// Get match stats
	getMatchStats: async () => {
		return apiCall<MatchStats>("/queue/match-stats");
	},
};

// User API functions
export const userAPI = {
	// Get user state
	getState: async () => {
		return apiCall<UserState>("/user/state");
	},

	// Get user profile
	getProfile: async () => {
		return apiCall<UserProfile>("/profile");
	},

	// Update user profile
	updateProfile: async (data: {
		age?: number | null;
		city?: string;
		is_male?: boolean;
		is_hidden?: boolean;
	}) => {
		try {
			const result = await apiCall<UserProfile>("/profile", {
				method: "PUT",
				body: JSON.stringify(data),
			});
			toast.success("Cập nhật thông tin thành công!");
			return result;
		} catch (error) {
			// Error already shown by apiCall
			throw error;
		}
	},
};

// Auth API functions
export const authAPI = {
	// Logout
	logout: async () => {
		try {
			const result = await apiCall<{ message: string }>("/auth/logout", {
				method: "POST",
			});
			toast.success("Đăng xuất thành công!");
			return result;
		} catch (error) {
			// Error already shown by apiCall
			throw error;
		}
	},

	// Get Google OAuth URL
	getGoogleAuthUrl: () => {
		return `${API_BASE}/auth/google`;
	},
};

// Health check
export const healthAPI = {
	check: async () => {
		return apiCall<{ status: string }>("/healthz");
	},
};
