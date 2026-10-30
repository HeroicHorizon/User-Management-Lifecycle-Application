export const UNREACHABLE_MESSAGE = "Unable to reach the server. Please try again later.";

const BASE_URL: string = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
    public details?: Record<string, string>,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

interface Envelope<T> {
  success: boolean;
  data?: T;
  error?: { code: string; message: string; details?: Record<string, string> };
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${BASE_URL}${path}`, {
      ...init,
      headers: { "Content-Type": "application/json", ...(init.headers ?? {}) },
    });
  } catch {
    throw new ApiError(UNREACHABLE_MESSAGE, 0, "NETWORK");
  }
  if (res.status === 503) throw new ApiError(UNREACHABLE_MESSAGE, 503, "SERVICE_UNAVAILABLE");

  let body: Envelope<T> | undefined;
  try {
    body = (await res.json()) as Envelope<T>;
  } catch {
    body = undefined;
  }
  if (!res.ok) {
    const e = body?.error;
    throw new ApiError(e?.message ?? "Request failed.", res.status, e?.code, e?.details);
  }
  return body?.data as T;
}
