export function errorText(err: unknown, fallback: string): string {
  if (typeof err === "string" && err) {
    return err;
  }

  if (err instanceof Error && err.message) {
    return err.message;
  }

  return fallback;
}
