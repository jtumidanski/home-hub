import { useMemo } from "react";
import type { Member } from "@/types/models/member";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";

interface OwnerSelectProps {
  value: string;
  onChange: (value: string) => void;
  members: Member[];
}

const EVERYONE_VALUE = "__everyone__";

export function OwnerSelect({ value, onChange, members }: OwnerSelectProps) {
  const displayLabel = useMemo(() => {
    if (!value || value === EVERYONE_VALUE) return "Everyone";
    const member = members.find((m) => m.relationships.user.data.id === value);
    return member?.attributes.displayName ?? value;
  }, [value, members]);

  return (
    <Select value={value || EVERYONE_VALUE} onValueChange={(v) => onChange((v ?? EVERYONE_VALUE) === EVERYONE_VALUE ? "" : v!)}>
      <SelectTrigger>
        <span>{displayLabel}</span>
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={EVERYONE_VALUE}>Everyone</SelectItem>
        {members.map((m) => (
          <SelectItem key={m.relationships.user.data.id} value={m.relationships.user.data.id}>
            {m.attributes.displayName}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
