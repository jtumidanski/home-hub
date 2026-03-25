import { Home } from "lucide-react";
import { type Household } from "@/types/models/household";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface HouseholdCardProps {
  household: Household;
  isActive: boolean;
}

export function HouseholdCard({ household, isActive }: HouseholdCardProps) {
  const { attributes } = household;

  return (
    <Card>
      <CardContent className="flex items-center gap-3 p-4">
        <Home className="h-5 w-5 shrink-0 text-muted-foreground" />
        <div className="min-w-0 flex-1">
          <p className="font-medium">{attributes.name}</p>
          <p className="text-xs text-muted-foreground">
            {attributes.timezone} &middot; {attributes.units}
          </p>
        </div>
        {isActive && <Badge>Active</Badge>}
      </CardContent>
    </Card>
  );
}
