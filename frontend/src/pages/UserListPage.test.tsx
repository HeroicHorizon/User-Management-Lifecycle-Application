import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { makeUser, mockFetch, ok, renderWithProviders } from "../testUtils";
import type { UserList } from "../types/user";
import { UserListPage } from "./UserListPage";

afterEach(() => vi.unstubAllGlobals());

const list = (items = [makeUser(1), makeUser(2)], total = items.length): UserList => ({
  items,
  page: 1,
  pageSize: 10,
  total,
});

describe("UserListPage", () => {
  it("renders rows with all columns", async () => {
    mockFetch(() => ok(list()));
    renderWithProviders(<UserListPage />);
    expect(await screen.findAllByTestId("user-row")).toHaveLength(2);
    for (const h of [
      "User ID",
      "First Name",
      "Last Name",
      "Email Address",
      "Phone Number",
      "Department",
      "Designation",
      "Status",
      "Created Date",
      "Updated Date",
    ]) {
      expect(screen.getByRole("columnheader", { name: h })).toBeInTheDocument();
    }
    expect(screen.getByText("user1@example.com")).toBeInTheDocument();
  });

  it('shows "No users found." with zero data rows', async () => {
    mockFetch(() => ok(list([])));
    renderWithProviders(<UserListPage />);
    expect(await screen.findByText("No users found.")).toBeInTheDocument();
    expect(screen.queryAllByTestId("user-row")).toHaveLength(0);
  });

  it("search and status filter send the right query params", async () => {
    const { calls } = mockFetch(() => ok(list()));
    const user = userEvent.setup();
    renderWithProviders(<UserListPage />);
    await screen.findAllByTestId("user-row");

    await user.type(screen.getByLabelText("Search users"), "john");
    await user.click(screen.getByRole("button", { name: "Search" }));
    await waitFor(() => expect(calls().some((c) => c.includes("search=john"))).toBe(true));

    await user.selectOptions(screen.getByLabelText("Filter by status"), "INACTIVE");
    await waitFor(() => {
      const last = calls().at(-1)!;
      expect(last).toContain("status=INACTIVE");
      expect(last).toContain("search=john");
      expect(last).toContain("page=1");
    });
  });

  it("next page requests page 2", async () => {
    const { calls } = mockFetch(() => ok(list([makeUser(1)], 25)));
    const user = userEvent.setup();
    renderWithProviders(<UserListPage />);
    await screen.findAllByTestId("user-row");
    expect(screen.getByText("Page 1 of 3")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Next" }));
    await waitFor(() => expect(calls().at(-1)).toContain("page=2"));
  });

  it("activate/deactivate calls the status API", async () => {
    const u = makeUser(1);
    const { calls } = mockFetch((_url, init) =>
      init.method === "PATCH" ? ok({ ...u, status: "INACTIVE" }) : ok(list([u])),
    );
    const user = userEvent.setup();
    renderWithProviders(<UserListPage />);
    await user.click(await screen.findByRole("button", { name: "Deactivate" }));
    await waitFor(() => expect(calls().some((c) => c.startsWith("PATCH") && c.endsWith(`/${u.id}/status`))).toBe(true));
    const patch = (vi.mocked(fetch).mock.calls.find(([, i]) => i?.method === "PATCH")![1] as RequestInit).body;
    expect(JSON.parse(patch as string)).toEqual({ status: "INACTIVE" });
  });

  it("an inactive user shows Activate and sends ACTIVE", async () => {
    const u = makeUser(1, { status: "INACTIVE" });
    mockFetch((_url, init) => (init.method === "PATCH" ? ok({ ...u, status: "ACTIVE" }) : ok(list([u]))));
    const user = userEvent.setup();
    renderWithProviders(<UserListPage />);
    await user.click(await screen.findByRole("button", { name: "Activate" }));
    await waitFor(() => {
      const call = vi.mocked(fetch).mock.calls.find(([, i]) => i?.method === "PATCH");
      expect(JSON.parse((call![1] as RequestInit).body as string)).toEqual({ status: "ACTIVE" });
    });
  });

  it("cancel sends no DELETE request", async () => {
    const { calls } = mockFetch(() => ok(list([makeUser(1)])));
    const user = userEvent.setup();
    renderWithProviders(<UserListPage />);
    await user.click(await screen.findByRole("button", { name: "Delete" }));
    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(calls().some((c) => c.startsWith("DELETE"))).toBe(false);
  });

  it("confirm sends DELETE and reloads the list", async () => {
    const u = makeUser(1);
    let deleted = false;
    const { calls } = mockFetch((_url, init) => {
      if (init.method === "DELETE") {
        deleted = true;
        return ok({ message: "User deleted." });
      }
      return ok(list(deleted ? [] : [u]));
    });
    const user = userEvent.setup();
    renderWithProviders(<UserListPage />);
    await user.click(await screen.findByRole("button", { name: "Delete" }));
    await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Confirm" }));
    await waitFor(() => expect(calls().some((c) => c.startsWith("DELETE") && c.endsWith(`/${u.id}`))).toBe(true));
    expect(await screen.findByText("No users found.")).toBeInTheDocument();
  });

  it("shows the unreachable-server notification when the list cannot load", async () => {
    mockFetch(() => {
      throw new TypeError("Failed to fetch");
    });
    renderWithProviders(<UserListPage />);
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Unable to reach the server. Please try again later.",
    );
  });
});
