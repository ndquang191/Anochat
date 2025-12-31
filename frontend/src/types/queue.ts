// Queue-related types
export interface QueueStatus {
	is_in_queue: boolean;
	position: number;
	category: string;
	joined_at: string;
	expires_at: string;
}

export interface QueueStats {
	totalMale: number;
	totalFemale: number;
	estimatedWaitTime: string;
}

export interface MatchStats {
	totalMatches: number;
	maleWaitTime: string;
	femaleWaitTime: string;
	lastMatchTime: string;
}
