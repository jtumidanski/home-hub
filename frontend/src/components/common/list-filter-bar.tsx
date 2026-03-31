import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { useHouseholdMembers } from "@/lib/hooks/api/use-household-members";
import type { Member } from "@/types/models/member";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { X } from "lucide-react";

interface ListFilterBarProps {
  statusOptions: { value: string; label: string }[];
}

export function ListFilterBar({ statusOptions }: ListFilterBarProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const { data } = useHouseholdMembers();
  const members = useMemo(() => (data?.data ?? []) as Member[], [data]);

  const query = searchParams.get("q") ?? "";
  const status = searchParams.get("status") ?? "all";
  const owner = searchParams.get("owner") ?? "all";

  const hasFilters = query !== "" || status !== "all" || owner !== "all";

  const ownerLabel = useMemo(() => {
    if (owner === "all") return "All owners";
    if (owner === "everyone") return "Everyone (unassigned)";
    const member = members.find((m) => m.relationships.user.data.id === owner);
    return member?.attributes.displayName ?? owner;
  }, [owner, members]);

  const updateParam = (key: string, value: string) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (value === "" || value === "all") {
        next.delete(key);
      } else {
        next.set(key, value);
      }
      return next;
    });
  };

  const clearFilters = () => {
    setSearchParams({});
  };

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Input
        placeholder="Search by title..."
        value={query}
        onChange={(e) => updateParam("q", e.target.value)}
        className="w-48"
      />
      <Select value={status} onValueChange={(v) => updateParam("status", v ?? "all")}>
        <SelectTrigger className="w-36">
          <span>{status === "all" ? "All statuses" : statusOptions.find((o) => o.value === status)?.label ?? status}</span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All statuses</SelectItem>
          {statusOptions.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Select value={owner} onValueChange={(v) => updateParam("owner", v ?? "all")}>
        <SelectTrigger className="w-40">
          <span>{ownerLabel}</span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All owners</SelectItem>
          <SelectItem value="everyone">Everyone (unassigned)</SelectItem>
          {members.map((m) => (
            <SelectItem key={m.relationships.user.data.id} value={m.relationships.user.data.id}>
              {m.attributes.displayName}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {hasFilters && (
        <Button variant="ghost" size="sm" onClick={clearFilters}>
          <X className="mr-1 h-4 w-4" />
          Clear filters
        </Button>
      )}
    </div>
  );
}
