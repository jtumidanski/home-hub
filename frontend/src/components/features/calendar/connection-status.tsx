import { RefreshCw, Unlink } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { CalendarConnection } from "@/types/models/calendar";
import { useTriggerSync } from "@/lib/hooks/api/use-calendar";
import { DisconnectDialog } from "./disconnect-dialog";

interface ConnectionStatusProps {
  connection: CalendarConnection;
}

export function ConnectionStatus({ connection }: ConnectionStatusProps) {
  const [showDisconnect, setShowDisconnect] = useState(false);
  const triggerSync = useTriggerSync();
  const { attributes: attrs } = connection;

  const isHealthy = attrs.status === "connected" || attrs.status === "syncing";

  const lastSync = attrs.lastSyncAt
    ? `Last synced ${new Date(attrs.lastSyncAt).toLocaleString([], { month: "short", day: "numeric", hour: "numeric", minute: "2-digit" })}`
    : "Never synced";

  return (
    <div className="flex items-center gap-3 text-sm">
      <div
        className="w-3 h-3 rounded-full flex-shrink-0"
        style={{ backgroundColor: attrs.userColor }}
      />
      {!isHealthy && (
        <Badge variant="outline" className="bg-red-500/10 text-red-700 border-red-200">
          {attrs.status}
        </Badge>
      )}
      <span className="text-muted-foreground">{attrs.userDisplayName}</span>
      <span className="text-muted-foreground">·</span>
      <span className="text-muted-foreground">{lastSync}</span>

      <div className="flex gap-1 ml-auto">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => triggerSync.mutate(connection.id)}
          disabled={triggerSync.isPending || attrs.status !== "connected"}
          title="Sync now"
        >
          <RefreshCw className={`h-4 w-4 ${triggerSync.isPending ? "animate-spin" : ""}`} />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setShowDisconnect(true)}
          title="Disconnect"
        >
          <Unlink className="h-4 w-4" />
        </Button>
      </div>

      <DisconnectDialog
        open={showDisconnect}
        onOpenChange={setShowDisconnect}
        connectionId={connection.id}
        email={attrs.email}
      />
    </div>
  );
}
