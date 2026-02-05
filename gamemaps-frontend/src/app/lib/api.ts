export type ApiPayload<T> = {
  status: "success" | "error";
  data?: T;
  error?: string;
};

export async function apiGet<T>(path: string): Promise<ApiPayload<T>> {
  const res = await fetch(`/api${path}`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
    cache: "no-store",
  });
  return (await res.json()) as ApiPayload<T>;
}

export async function apiSend<T>(
  path: string,
  body?: any,
  method: "POST" | "PATCH" | "DELETE" = "POST"
): Promise<ApiPayload<T>> {
  const res = await fetch(`/api${path}`, {
    method,
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
    cache: "no-store",
  });
  return (await res.json()) as ApiPayload<T>;
}
