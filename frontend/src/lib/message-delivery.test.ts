import { describe, expect, test } from "bun:test";

import {
	reconcileAuthoritativeMessages,
	updateMessageDeliveryStatus,
} from "./message-delivery";

const pendingMessage = {
	id: "message-1",
	room_id: "room-1",
	sender_id: "user-1",
	content: "hello",
	created_at: 1,
	status: "pending" as const,
};

describe("updateMessageDeliveryStatus", () => {
	test("marks the matching optimistic message as sent", () => {
		const result = updateMessageDeliveryStatus(
			[pendingMessage],
			"message-1",
			"sent",
			2
		);

		expect(result[0].status).toBe("sent");
		expect(result[0].created_at).toBe(2);
	});

	test("marks the matching optimistic message as failed", () => {
		const result = updateMessageDeliveryStatus(
			[pendingMessage],
			"message-1",
			"failed"
		);

		expect(result[0].status).toBe("failed");
	});

	test("preserves the array when the message is no longer present", () => {
		const messages = [pendingMessage];

		expect(
			updateMessageDeliveryStatus(messages, "missing", "sent")
		).toBe(messages);
	});
});

describe("reconcileAuthoritativeMessages", () => {
	test("replaces a pending message with its persisted representation", () => {
		const persisted = {
			...pendingMessage,
			content: "persisted",
			created_at: 2,
			status: undefined,
		};

		const result = reconcileAuthoritativeMessages(
			[pendingMessage],
			[persisted]
		);

		expect(result).toHaveLength(1);
		expect(result[0].content).toBe("persisted");
		expect(result[0].created_at).toBe(2);
		expect(result[0].status).toBe("sent");
	});

	test("adds authoritative messages that are not in local state", () => {
		const persisted = {
			...pendingMessage,
			id: "message-2",
			created_at: 2,
			status: undefined,
		};

		expect(
			reconcileAuthoritativeMessages([pendingMessage], [persisted]).map(
				(message) => message.id
			)
		).toEqual(["message-1", "message-2"]);
	});
});
