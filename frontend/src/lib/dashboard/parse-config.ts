import { z } from "zod";

export type ParsedConfig<T> = { config: T; lossy: boolean };

export function parseConfig<T>(
  def: { defaultConfig: T; configSchema: z.ZodType<T> },
  raw: unknown,
): ParsedConfig<T> {
  const merged = { ...(def.defaultConfig as object), ...(isRecord(raw) ? raw : {}) };
  const result = def.configSchema.safeParse(merged);
  if (result.success) return { config: result.data, lossy: false };
  return { config: def.defaultConfig, lossy: true };
}

function isRecord(x: unknown): x is Record<string, unknown> {
  return typeof x === "object" && x !== null && !Array.isArray(x);
}
