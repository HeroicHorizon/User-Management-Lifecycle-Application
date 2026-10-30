import { render } from "@testing-library/react";
import type { ReactElement } from "react";
import { MemoryRouter } from "react-router-dom";
import { vi } from "vitest";
import { NotificationProvider } from "./components/NotificationProvider";
import type { User } from "./types/user";

export function makeUser(n: number, over: Partial<User> = {}): User {
  return {
    id: `64b7f0c2a1b2c3d4e5f6070${n}`,
    firstName: `First${n}`,
    lastName: `Last${n}`,
    email: `user${n}@example.com`,
    phone: "+911234567890",
    department: "Engineering",
    designation: "Engineer",
    status: "ACTIVE",
    createdAt: "2026-10-30T10:00:00Z",
    updatedAt: "2026-10-30T10:00:00Z",
    ...over,
  };
}

export const ok = (data: unknown, status = 200) =>
  new Response(JSON.stringify({ success: true, data }), { status, headers: { "Content-Type": "application/json" } });

export const fail = (status: number, code: string, message: string, details?: Record<string, string>) =>
  new Response(JSON.stringify({ success: false, error: { code, message, details } }), {
    status,
    headers: { "Content-Type": "application/json" },
  });

export type Handler = (url: string, init: RequestInit) => Response | Promise<Response>;

/** Replaces global fetch; every call is recorded as "METHOD url". */
export function mockFetch(handler: Handler) {
  const fn = vi.fn((url: RequestInfo | URL, init: RequestInit = {}) => {
    try {
      return Promise.resolve(handler(String(url), init));
    } catch (e) {
      return Promise.reject(e);
    }
  });
  vi.stubGlobal("fetch", fn);
  const calls = () => fn.mock.calls.map(([u, i]) => `${(i as RequestInit | undefined)?.method ?? "GET"} ${String(u)}`);
  return { fn, calls };
}

export function renderWithProviders(ui: ReactElement, route = "/") {
  return render(
    <MemoryRouter initialEntries={[route]}>
      <NotificationProvider>{ui}</NotificationProvider>
    </MemoryRouter>,
  );
}
