import { Home, MapPin } from "lucide-react";
import { type Household } from "@/types/models/household";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LocationSearch } from "@/components/features/weather/location-search";
import { useUpdateHousehold } from "@/lib/hooks/api/use-households";

interface HouseholdCardProps {
  household: Household;
  isActive: boolean;
}

export function HouseholdCard({ household, isActive }: HouseholdCardProps) {
  const { attributes } = household;
  const updateHousehold = useUpdateHousehold();

  const handleLocationSelect = (place: { name: string; latitude: number; longitude: number }) => {
    updateHousehold.mutate({
      householdId: household.id,
      attrs: {
        name: attributes.name,
        timezone: attributes.timezone,
        units: attributes.units,
        latitude: place.latitude,
        longitude: place.longitude,
        locationName: place.name,
      },
    });
  };

  const handleLocationClear = () => {
    updateHousehold.mutate({
      householdId: household.id,
      attrs: {
        name: attributes.name,
        timezone: attributes.timezone,
        units: attributes.units,
        latitude: null,
        longitude: null,
        locationName: null,
      },
    });
  };

  return (
    <Card className="overflow-visible">
      <CardContent className="p-4 space-y-3 overflow-visible">
        <div className="flex items-center gap-3">
          <Home className="h-5 w-5 shrink-0 text-muted-foreground" />
          <div className="min-w-0 flex-1">
            <p className="font-medium">{attributes.name}</p>
            <p className="text-xs text-muted-foreground">
              {attributes.timezone} &middot; {attributes.units}
            </p>
          </div>
          {isActive && <Badge>Active</Badge>}
        </div>
        <div>
          <div className="flex items-center gap-1 mb-1">
            <MapPin className="h-3 w-3 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">Location</span>
          </div>
          <LocationSearch
            value={attributes.locationName}
            onSelect={handleLocationSelect}
            onClear={handleLocationClear}
            isPending={updateHousehold.isPending}
          />
        </div>
      </CardContent>
    </Card>
  );
}
