import { screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { mockFetch, ok, renderWithProviders } from "../testUtils";
import { Dashboard } from "./Dashboard";

afterEach(() => vi.unstubAllGlobals());

describe("Dashboard", () => {
  it("shows total, active and inactive counts", async () => {
    const { calls } = mockFetch(() => ok({ total: 7, active: 5, inactive: 2 }));
    renderWithProviders(<Dashboard />);
    await waitFor(() => expect(screen.getByTestId("stat-total-users")).toHaveTextContent("7"));
    expect(screen.getByTestId("stat-active-users")).toHaveTextContent("5");
    expect(screen.getByTestId("stat-inactive-users")).toHaveTextContent("2");
    expect(calls()[0]).toMatch(/GET .*\/api\/v1\/users\/stats$/);
  });

  it("shows the unreachable-server notification when the backend is down", async () => {
    mockFetch(() => {
      throw new TypeError("Failed to fetch");
    });
    renderWithProviders(<Dashboard />);
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Unable to reach the server. Please try again later.",
    );
  });
});
