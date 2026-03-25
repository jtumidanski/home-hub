import { useState, useCallback } from "react";

const STORAGE_KEY = "sidebar-collapsed-groups";

function readCollapsedGroups(): Set<string> {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      return new Set(JSON.parse(stored) as string[]);
    }
  } catch {
    // ignore invalid localStorage data
  }
  return new Set();
}

function writeCollapsedGroups(collapsed: Set<string>) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify([...collapsed]));
}

export function useNavGroupState() {
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(readCollapsedGroups);

  const toggleGroup = useCallback((groupKey: string) => {
    setCollapsedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(groupKey)) {
        next.delete(groupKey);
      } else {
        next.add(groupKey);
      }
      writeCollapsedGroups(next);
      return next;
    });
  }, []);

  const isGroupOpen = useCallback(
    (groupKey: string, hasActiveRoute: boolean) => {
      if (hasActiveRoute) return true;
      return !collapsedGroups.has(groupKey);
    },
    [collapsedGroups],
  );

  return { toggleGroup, isGroupOpen };
}
