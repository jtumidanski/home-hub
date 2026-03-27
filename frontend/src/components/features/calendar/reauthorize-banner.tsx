import { ShieldAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useReauthorizeCalendar } from "@/lib/hooks/api/use-calendar";

export function ReauthorizeBanner() {
  const reauthorize = useReauthorizeCalendar();

  const handleUpgrade = () => {
    reauthorize.mutate(window.location.origin + "/app/calendar");
  };

  return (
    <div className="flex items-center gap-3 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm dark:border-amber-800 dark:bg-amber-950/30">
      <ShieldAlert className="h-4 w-4 text-amber-600 flex-shrink-0" />
      <span className="flex-1 text-amber-800 dark:text-amber-200">
        Upgrade your calendar connection to add and edit events.
      </span>
      <Button
        variant="outline"
        size="sm"
        onClick={handleUpgrade}
        disabled={reauthorize.isPending}
      >
        Upgrade Access
      </Button>
    </div>
  );
}
