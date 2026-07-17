export interface UserDTO {
	id: string;
	email?: string;
	name?: string;
	nickname?: string;
	avatar_url?: string;
	is_admin?: boolean;
	profile?: ProfileDTO;
}

export interface ProfileDTO {
	nickname?: string;
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
	is_admin: boolean;
}

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	error?: string;
	message?: string;
}

export interface BannedWordDTO {
	id: string;
	word: string;
	category: string;
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

export interface BannedUserDTO {
	id: string;
	name?: string;
	email?: string;
	created_at: number;
	last_room_id?: string;
}

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const MAX_MESSAGE_LENGTH = 1000;
export const MIN_AGE = Number(process.env.NEXT_PUBLIC_MIN_AGE ?? 10);
export const MAX_AGE = Number(process.env.NEXT_PUBLIC_MAX_AGE ?? 99);
