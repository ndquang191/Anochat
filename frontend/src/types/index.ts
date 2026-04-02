export interface UserDTO {
	id: string;
	email?: string;
	name?: string;
	avatar_url?: string;
	profile?: ProfileDTO;
}

export interface ProfileDTO {
	age?: number;
	is_male?: boolean;
	is_hidden: boolean;
}

export interface RoomDTO {
	id: string;
	user1_id: string;
	user2_id: string;
	partner?: UserDTO;
}

export interface MessageDTO {
	id: string;
	room_id: string;
	sender_id: string;
	content: string;
	created_at: number;
}

export interface UserStateResponse {
	profile?: ProfileDTO;
	room?: RoomDTO;
	messages?: MessageDTO[];
	in_queue?: boolean;
}

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	error?: string;
	message?: string;
}

export const AdminUserID = "8d2e7280-bdc8-47b2-8508-8911b5c9f796";

export interface BannedWordDTO {
	id: string;
	word: string;
	created_at: number;
}

export interface ReportDTO {
	id: string;
	reporter_id: string;
	reported_user_id: string;
	reported_user_name?: string;
	room_id: string;
	status: "pending" | "reviewed";
	created_at: number;
}

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const MAX_MESSAGE_LENGTH = 1000;
export const MIN_AGE = 13;
export const MAX_AGE = 100;
