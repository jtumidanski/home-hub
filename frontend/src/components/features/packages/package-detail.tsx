import { usePackage } from "@/lib/hooks/api/use-packages";
import { StatusBadge } from "./status-badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ExternalLink } from "lucide-react";

interface PackageDetailProps {
  packageId: string;
  carrier: string;
  trackingNumber: string | null;
}

const CARRIER_URLS: Record<string, (tn: string) => string> = {
  usps: (tn) => `https://tools.usps.com/go/TrackConfirmAction?tLabels=${tn}`,
  ups: (tn) => `https://www.ups.com/track?tracknum=${tn}`,
  fedex: (tn) => `https://www.fedex.com/fedextrack/?trknbr=${tn}`,
};

export function PackageDetail({ packageId, carrier, trackingNumber }: PackageDetailProps) {
  const { data, isLoading } = usePackage(packageId);

  if (isLoading) {
    return (
      <div className="mt-3 pt-3 border-t space-y-2">
        <Skeleton className="h-4 w-48" />
        <Skeleton className="h-4 w-64" />
        <Skeleton className="h-4 w-56" />
      </div>
    );
  }

  const carrierUrl = trackingNumber && CARRIER_URLS[carrier]
    ? CARRIER_URLS[carrier](trackingNumber)
    : null;

  const events = data?.data?.attributes?.trackingEvents ?? [];

  return (
    <div className="mt-3 pt-3 border-t">
      {carrierUrl && (
        <a
          href={carrierUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 text-xs text-primary hover:underline mb-3"
        >
          View on carrier website <ExternalLink className="h-3 w-3" />
        </a>
      )}

      <div className="space-y-2">
        <h4 className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          Tracking History
        </h4>
        {events.length === 0 ? (
          <p className="text-xs text-muted-foreground">No tracking events yet.</p>
        ) : (
          <div className="space-y-0">
            {events.map((evt, i) => (
              <div key={i} className="flex gap-3 py-1.5">
                <div className="flex flex-col items-center">
                  <div className={`w-2 h-2 rounded-full mt-1.5 ${i === 0 ? "bg-primary" : "bg-muted-foreground/30"}`} />
                  {i < events.length - 1 && <div className="w-px flex-1 bg-muted-foreground/20" />}
                </div>
                <div className="min-w-0 flex-1 pb-2">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-xs font-medium">{evt.description}</span>
                    <StatusBadge status={evt.status} />
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                    <span>{new Date(evt.timestamp).toLocaleString()}</span>
                    {evt.location && <span>· {evt.location}</span>}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
