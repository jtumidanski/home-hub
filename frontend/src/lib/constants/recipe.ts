export const CLASSIFICATIONS = ["breakfast", "lunch", "dinner", "snack", "side"] as const;

export const UNIT_FAMILIES = ["", "count", "weight", "volume"] as const;

export type Classification = (typeof CLASSIFICATIONS)[number];

export function isClassification(tag: string): tag is Classification {
  return CLASSIFICATIONS.includes(tag as Classification);
}

export function extractClassification(tags: string[]): string | undefined {
  return tags.find(isClassification);
}

export function filterClassificationTags(tags: string[]): string[] {
  return tags.filter((t) => !isClassification(t));
}
