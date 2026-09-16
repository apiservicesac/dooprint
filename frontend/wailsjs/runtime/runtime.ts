// Minimal runtime helpers for the web version.

// The backend sends no events: the interface refreshes by polling.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function EventsOn(_event: string, _callback: (...data: any[]) => void): () => void {
  return () => {};
}

export function BrowserOpenURL(url: string): void {
  window.open(url, "_blank", "noopener");
}
