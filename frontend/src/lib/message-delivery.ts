import type { ChatMessage, MessageDeliveryStatus } from "@/lib/websocket";

export function updateMessageDeliveryStatus(
	messages: ChatMessage[],
	id: string,
	status: MessageDeliveryStatus,
	createdAt?: number
): ChatMessage[] {
	let changed = false;
	const updated = messages.map((message) => {
		if (message.id !== id) {
			return message;
		}
		changed = true;
		return {
			...message,
			status,
			created_at: createdAt ?? message.created_at,
		};
	});
	return changed ? updated : messages;
}

export function reconcileAuthoritativeMessages(
	current: ChatMessage[],
	authoritative: ChatMessage[]
): ChatMessage[] {
	const authoritativeById = new Map(
		authoritative.map((message) => [message.id, message])
	);
	const currentIds = new Set(current.map((message) => message.id));
	const reconciled = current.map((message) => {
		const persisted = authoritativeById.get(message.id);
		return persisted ? { ...persisted, status: "sent" as const } : message;
	});
	for (const message of authoritative) {
		if (!currentIds.has(message.id)) {
			reconciled.push({ ...message, status: "sent" });
		}
	}
	return reconciled.sort((a, b) => a.created_at - b.created_at);
}
