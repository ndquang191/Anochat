import { describe, expect, test } from "bun:test";

import { prependUniqueMessages } from "./message-history";

describe("prependUniqueMessages", () => {
	test("prepends older messages in API order", () => {
		const current = [
			{
				id: "3",
				room_id: "room-1",
				sender_id: "user-1",
				content: "new",
				created_at: 3,
			},
		];
		const older = [
			{
				id: "1",
				room_id: "",
				sender_id: "user-2",
				content: "oldest",
				created_at: 1,
			},
			{
				id: "2",
				room_id: "",
				sender_id: "user-1",
				content: "older",
				created_at: 2,
			},
		];

		const result = prependUniqueMessages(current, older, "room-1");

		expect(result.map((message) => message.id)).toEqual(["1", "2", "3"]);
		expect(result[0].room_id).toBe("room-1");
	});

	test("does not insert duplicate message ids", () => {
		const current = [
			{
				id: "2",
				room_id: "room-1",
				sender_id: "user-1",
				content: "existing",
				created_at: 2,
			},
		];
		const older = [
			{
				id: "1",
				room_id: "room-1",
				sender_id: "user-2",
				content: "old",
				created_at: 1,
			},
			{
				id: "2",
				room_id: "room-1",
				sender_id: "user-1",
				content: "duplicate",
				created_at: 2,
			},
		];

		expect(
			prependUniqueMessages(current, older, "room-1").map(
				(message) => message.id
			)
		).toEqual(["1", "2"]);
	});
});
