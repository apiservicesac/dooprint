// Bridge between the frontend and the server API served by the binary (internal/webui).
export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!response.ok) {
    throw new Error((await response.text()) || response.statusText);
  }
  return (await response.json()) as T;
}
