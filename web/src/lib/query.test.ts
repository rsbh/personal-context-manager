import { describe, expect, it } from "vitest";
import { shouldRetry, withStatus } from "./query";

describe("withStatus", () => {
  it("adds the HTTP status to an error body", () => {
    expect(withStatus({ error: "not found" }, new Response("", { status: 404 }))).toEqual({ error: "not found", status: 404 });
  });
  it("wraps a non-object error", () => {
    expect(withStatus("boom", new Response("", { status: 500 }))).toEqual({ error: "boom", status: 500 });
  });
});

describe("shouldRetry", () => {
  it("does not retry client errors", () => {
    expect(shouldRetry(0, { status: 404 })).toBe(false);
    expect(shouldRetry(0, { status: 400 })).toBe(false);
  });
  it("retries server and network errors twice", () => {
    expect(shouldRetry(0, { status: 500 })).toBe(true);
    expect(shouldRetry(1, new Error("network"))).toBe(true);
    expect(shouldRetry(2, { status: 503 })).toBe(false);
  });
});
