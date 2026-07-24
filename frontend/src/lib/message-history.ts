import type { MessageDTO } from "@/types";
import type { ChatMessage } from "@/lib/websocket";

export function prependUniqueMessages(
	current: ChatMessage[],
	older: MessageDTO[],
	roomId: string
): ChatMessage[] {
	const seen = new Set(current.map((message) => message.id));
	const uniqueOlder: ChatMessage[] = [];

	for (const message of older) {
		if (seen.has(message.id)) continue;
		seen.add(message.id);
		uniqueOlder.push({
			...message,
			room_id: message.room_id || roomId,
		});
	}

	return uniqueOlder.length > 0 ? [...uniqueOlder, ...current] : current;
}
