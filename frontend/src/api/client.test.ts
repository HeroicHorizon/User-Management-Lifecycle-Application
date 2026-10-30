import { afterEach, describe, expect, it, vi } from "vitest";
import { fail, mockFetch, ok } from "../testUtils";
import { ApiError, request, UNREACHABLE_MESSAGE } from "./client";

afterEach(() => vi.unstubAllGlobals());

describe("api client", () => {
  it('network error shows "Unable to reach the server. Please try again later."', async () => {
    mockFetch(() => {
      throw new TypeError("Failed to fetch");
    });
    await expect(request("/x")).rejects.toMatchObject({ message: UNREACHABLE_MESSAGE });
    expect(UNREACHABLE_MESSAGE).toBe("Unable to reach the server. Please try again later.");
  });

  it("a 503 response shows the same message", async () => {
    mockFetch(() => fail(503, "SERVICE_UNAVAILABLE", "Service temporarily unavailable."));
    await expect(request("/x")).rejects.toMatchObject({ message: UNREACHABLE_MESSAGE, status: 503 });
  });

  it("returns the data field of a successful envelope", async () => {
    mockFetch(() => ok({ a: 1 }));
    await expect(request("/x")).resolves.toEqual({ a: 1 });
  });

  it("maps API errors to ApiError with status, code and details", async () => {
    mockFetch(() => fail(400, "VALIDATION_ERROR", "Validation failed.", { email: "bad" }));
    const err = await request("/x").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ status: 400, code: "VALIDATION_ERROR", details: { email: "bad" } });
  });

  it("handles a non-JSON error body", async () => {
    mockFetch(() => new Response("oops", { status: 500 }));
    await expect(request("/x")).rejects.toMatchObject({ status: 500, message: "Request failed." });
  });
});
