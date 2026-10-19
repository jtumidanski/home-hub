import { ShieldAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { StatusBanner } from "@/components/ui/status-banner";
import { useReauthorizeCalendar } from "@/lib/hooks/api/use-calendar";

export function ReauthorizeBanner() {
  const reauthorize = useReauthorizeCalendar();

  const handleUpgrade = () => {
    reauthorize.mutate(window.location.origin + "/app/calendar");
  };

  return (
    <StatusBanner variant="warning" icon={ShieldAlert} className="items-center">
      <div className="flex items-center gap-3">
        <span className="flex-1">Upgrade your calendar connection to add and edit events.</span>
        <Button
          variant="outline"
          size="sm"
          onClick={handleUpgrade}
          disabled={reauthorize.isPending}
        >
          Upgrade Access
        </Button>
      </div>
    </StatusBanner>
  );
}
