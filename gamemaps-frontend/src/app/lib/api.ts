export type ApiPayload<T> = {
  status: "success" | "error";
  data?: T;
  error?: string;
};

const FALLBACK_ERROR = "Request failed. Please try again.";

async function requestApi<T>(
  path: string,
  init: RequestInit
): Promise<ApiPayload<T>> {
  try {
    const res = await fetch(`/api${path}`, {
      headers: { "Content-Type": "application/json" },
      cache: "no-store",
      ...init,
    });

    let payload: ApiPayload<T> | null = null;
    try {
      payload = (await res.json()) as ApiPayload<T>;
    } catch {
      payload = null;
    }

    if (payload && (payload.status === "success" || payload.status === "error")) {
      return payload;
    }

    if (!res.ok) {
      return { status: "error", error: `Request failed (${res.status}).` };
    }

    return { status: "error", error: FALLBACK_ERROR };
  } catch {
    return { status: "error", error: "Network error. Check your connection and try again." };
  }
}

export async function apiGet<T>(path: string): Promise<ApiPayload<T>> {
  return requestApi<T>(path, {
    method: "GET",
  });
}

export async function apiSend<T>(
  path: string,
  body?: unknown,
  method: "POST" | "PATCH" | "DELETE" = "POST"
): Promise<ApiPayload<T>> {
  return requestApi<T>(path, {
    method,
    body: body ? JSON.stringify(body) : undefined,
  });
}
