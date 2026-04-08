const MESSAGES: Record<string, string> = {
  token_revoked:
    "Access was revoked from your Google account. Reconnect to resume syncing.",
  refresh_unauthorized:
    "Google rejected your saved credentials. Reconnect to resume syncing.",
  token_decrypt_failed:
    "We can't read your stored credentials. Reconnect to resume syncing.",
  refresh_http_error: "Couldn't reach Google. Retrying automatically.",
  unknown: "Sync is failing. Retrying automatically.",
};

export function errorCodeToMessage(code: string | null): string | null {
  if (!code) return null;
  return MESSAGES[code] ?? null;
}
