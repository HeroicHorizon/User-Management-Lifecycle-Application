import type { ListParams, User, UserInput, UserList, UserStats, UserStatus } from "../types/user";
import { request } from "./client";

export function listUsers(p: ListParams): Promise<UserList> {
  const qs = new URLSearchParams({ page: String(p.page), pageSize: String(p.pageSize) });
  if (p.search.trim()) qs.set("search", p.search.trim());
  if (p.status) qs.set("status", p.status);
  return request<UserList>(`/api/v1/users?${qs.toString()}`);
}

export const getStats = () => request<UserStats>("/api/v1/users/stats");

export const getUser = (id: string) => request<User>(`/api/v1/users/${encodeURIComponent(id)}`);

export const createUser = (input: UserInput) =>
  request<User>("/api/v1/users", { method: "POST", body: JSON.stringify(input) });

export const updateUser = (id: string, input: UserInput) =>
  request<User>(`/api/v1/users/${encodeURIComponent(id)}`, { method: "PUT", body: JSON.stringify(input) });

export const setUserStatus = (id: string, status: UserStatus) =>
  request<User>(`/api/v1/users/${encodeURIComponent(id)}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });

export const deleteUser = (id: string) =>
  request<{ message: string }>(`/api/v1/users/${encodeURIComponent(id)}`, { method: "DELETE" });
