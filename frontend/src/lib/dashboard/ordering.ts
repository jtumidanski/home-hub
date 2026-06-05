import type { Dashboard } from "@/types/models/dashboard";

/**
 * Sort dashboards by sortOrder ASC, then createdAt ASC as a stable tiebreaker.
 */
export function sortDashboards(list: Dashboard[]): Dashboard[] {
  return [...list].sort((a, b) => {
    if (a.attributes.sortOrder !== b.attributes.sortOrder) {
      return a.attributes.sortOrder - b.attributes.sortOrder;
    }
    return a.attributes.createdAt.localeCompare(b.attributes.createdAt);
  });
}

/**
 * Given a sorted list and active/over ids from a dnd-kit drag-end, returns
 * the reorder payload with 0-indexed sortOrder. Returns null when the drag
 * is a no-op.
 */
export function computeReorderEntries(
  sorted: Dashboard[],
  activeId: string,
  overId: string,
): Array<{ id: string; sortOrder: number }> | null {
  if (activeId === overId) return null;
  const fromIdx = sorted.findIndex((d) => d.id === activeId);
  const toIdx = sorted.findIndex((d) => d.id === overId);
  if (fromIdx < 0 || toIdx < 0) return null;
  const next = [...sorted];
  const [moved] = next.splice(fromIdx, 1);
  if (!moved) return null;
  next.splice(toIdx, 0, moved);
  return next.map((d, i) => ({ id: d.id, sortOrder: i }));
}
