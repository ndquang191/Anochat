// ============================================================================
// CORE DATABASE TYPES
// Based on database.md schema
// ============================================================================

export interface User {
	id: string;
	name?: string;
	email?: string;
	avatar_url?: string;
	is_active: boolean;
	is_deleted: boolean;
	created_at: string;
	profile?: Profile;
	room?: Room;
	messages?: Message[];
}

export interface Profile {
	user_id: string;
	is_male?: boolean;
	age?: number;
	city?: string;
	is_hidden: boolean;
	updated_at: string;
	user?: User;
}

export interface Room {
	id: string;
	user1_id: string;
	user2_id: string;
	category: string;
	created_at: string;
	ended_at?: string;
	is_sensitive: boolean;
	user1_last_read_message_id?: string;
	user2_last_read_message_id?: string;
	is_deleted: boolean;
	user1?: User;
	user2?: User;
	messages?: Message[];
}

export interface Message {
	id: string;
	room_id: string;
	sender_id: string;
	content: string;
	created_at: string;
	room?: Room;
	sender?: User;
}

// ============================================================================
// API REQUEST/RESPONSE TYPES
// Based on API.md endpoints
// ============================================================================

// Auth Types
export interface AuthResponse {
	user: User;
	token: string;
}

export interface OAuthCallbackResponse {
	token: string;
	user: string; // URL encoded JSON string
}

// User & Profile Types
export interface GetUserResponse {
	user: User;
	room?: Room;
	messages?: Message[];
}

export interface GetProfileResponse {
	profile: Profile;
}

export interface UpdateProfileRequest {
	is_male?: boolean;
	age?: number;
	city?: string;
	is_hidden?: boolean;
}

export interface UpdateProfileResponse {
	profile: Profile;
}

// Room Types (for future use)
export interface CreateRoomRequest {
	user1_id: string;
	user2_id: string;
	category?: string;
}

export interface CreateRoomResponse {
	room: Room;
}

export interface GetRoomResponse {
	room: Room;
	messages: Message[];
}

export interface LeaveRoomRequest {
	room_id: string;
}

export interface LeaveRoomResponse {
	success: boolean;
}

export interface UpdateRoomRequest {
	room_id: string;
	is_sensitive?: boolean;
	ended_at?: string;
	user1_last_read_message_id?: string;
	user2_last_read_message_id?: string;
}

export interface UpdateRoomResponse {
	room: Room;
}

// Message Types (for future use)
export interface CreateMessageRequest {
	room_id: string;
	content: string;
}

export interface CreateMessageResponse {
	message: Message;
}

export interface UpdateMessageSeenRequest {
	message_id: string;
}

export interface UpdateMessageSeenResponse {
	success: boolean;
}

// ============================================================================
// FRONTEND STATE TYPES
// ============================================================================

// Auth State
export interface AuthState {
	isAuthenticated: boolean;
	user: User | null;
	token: string | null;
	loading: boolean;
}

export interface AuthContextType extends AuthState {
	login: (token: string, user: User) => void;
	logout: () => void;
	checkAuth: () => Promise<void>;
}

// Chat State

// ============================================================================
// UTILITY TYPES
// ============================================================================

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	error?: string;
	message?: string;
}

export interface PaginatedResponse<T> {
	data: T[];
	total: number;
	page: number;
	limit: number;
	hasNext: boolean;
	hasPrev: boolean;
}

export interface PaginationParams {
	page?: number;
	limit?: number;
	sort?: string;
	order?: "asc" | "desc";
}

// ============================================================================
// ENUMS
// ============================================================================

export enum RoomCategory {
	POLITE = "polite",
	FRIENDLY = "friendly",
	ROMANTIC = "romantic",
}

export enum MessageStatus {
	SENT = "sent",
	DELIVERED = "delivered",
	READ = "read",
}

export enum UserStatus {
	ONLINE = "online",
	OFFLINE = "offline",
	AWAY = "away",
	BUSY = "busy",
}

// ============================================================================
// CONSTANTS
// ============================================================================

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const DEFAULT_PAGE_SIZE = 20;
export const MAX_MESSAGE_LENGTH = 1000;
export const MIN_AGE = 13;
export const MAX_AGE = 100;

// ============================================================================
// TYPE GUARDS
// ============================================================================
