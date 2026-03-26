import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useConnectGoogleCalendar } from "@/lib/hooks/api/use-calendar";

export function ConnectCalendarButton() {
  const connect = useConnectGoogleCalendar();

  const handleConnect = () => {
    const redirectUri = `${window.location.origin}/api/v1/calendar/connections/google/callback`;
    connect.mutate(redirectUri);
  };

  return (
    <Button size="sm" onClick={handleConnect} disabled={connect.isPending}>
      {connect.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
      Connect Google Calendar
    </Button>
  );
}
