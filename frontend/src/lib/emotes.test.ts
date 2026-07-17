import { describe, expect, test } from "bun:test";

import { convertEmotes } from "./emotes";

describe("convertEmotes", () => {
  test("replaces supported emotes", () => {
    expect(convertEmotes("hello :fire:")).toBe("hello 🔥");
  });

  test("replaces repeated and different emotes", () => {
    expect(convertEmotes(":fire: :heart: :fire:")).toBe("🔥 ❤️ 🔥");
  });

  test("leaves unsupported text unchanged", () => {
    expect(convertEmotes("hello :unknown:")).toBe("hello :unknown:");
  });
});
