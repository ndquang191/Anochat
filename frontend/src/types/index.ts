export interface UserDTO {
	id: string;
	email?: string;
	name?: string;
	avatar_url?: string;
	profile?: ProfileDTO;
}

export interface ProfileDTO {
	age?: number;
	city?: string;
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
	user: UserDTO;
	room?: RoomDTO;
	messages?: MessageDTO[];
	is_new_user: boolean;
	in_queue?: boolean;
}

export interface ApiResponse<T> {
	success: boolean;
	data?: T;
	error?: string;
	message?: string;
}

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export const MAX_MESSAGE_LENGTH = 1000;
export const MIN_AGE = 13;
export const MAX_AGE = 100;
