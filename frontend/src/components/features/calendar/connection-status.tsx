import { RefreshCw, Unlink } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { CalendarConnection } from "@/types/models/calendar";
import { useReauthorizeCalendar, useTriggerSync } from "@/lib/hooks/api/use-calendar";
import { DisconnectDialog } from "./disconnect-dialog";
import { errorCodeToMessage } from "./error-code-message";

interface ConnectionStatusProps {
  connection: CalendarConnection;
}

function formatRelative(iso: string | null): string | null {
  if (!iso) return null;
  const then = new Date(iso).getTime();
  const diffMs = Date.now() - then;
  const diffSec = Math.round(diffMs / 1000);
  const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
  const abs = Math.abs(diffSec);
  if (abs < 60) return rtf.format(-diffSec, "second");
  if (abs < 3600) return rtf.format(-Math.round(diffSec / 60), "minute");
  if (abs < 86400) return rtf.format(-Math.round(diffSec / 3600), "hour");
  return rtf.format(-Math.round(diffSec / 86400), "day");
}

export function ConnectionStatus({ connection }: ConnectionStatusProps) {
  const [showDisconnect, setShowDisconnect] = useState(false);
  const triggerSync = useTriggerSync();
  const reauthorize = useReauthorizeCalendar();
  const { attributes: attrs } = connection;

  const isDisconnected = attrs.status === "disconnected";
  const isError = attrs.status === "error";

  const failureMessage =
    errorCodeToMessage(attrs.errorCode) ??
    (isDisconnected ? "This calendar is disconnected." : null);

  const lastSync = attrs.lastSyncAt
    ? `Last synced ${new Date(attrs.lastSyncAt).toLocaleString([], { month: "short", day: "numeric", hour: "numeric", minute: "2-digit" })}`
    : "Never synced";

  // When attempt and success diverge, show "Tried X ago · Last success Y ago".
  const showAttemptSubline =
    !!attrs.lastSyncAttemptAt &&
    attrs.lastSyncAttemptAt !== attrs.lastSyncAt;
  const triedRel = formatRelative(attrs.lastSyncAttemptAt);
  const successRel = attrs.lastSyncAt ? formatRelative(attrs.lastSyncAt) : null;

  return (
    <div className="flex flex-col gap-1 text-sm">
      <div className="flex items-center gap-3">
        <div
          className="w-3 h-3 rounded-full flex-shrink-0"
          style={{ backgroundColor: attrs.userColor }}
        />
        {isError && (
          <Badge
            variant="outline"
            className="bg-amber-500/10 text-amber-700 border-amber-200"
          >
            Sync issues
          </Badge>
        )}
        {isDisconnected && (
          <Badge
            variant="outline"
            className="bg-red-500/10 text-red-700 border-red-200"
          >
            Disconnected
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
            <RefreshCw
              className={`h-4 w-4 ${triggerSync.isPending ? "animate-spin" : ""}`}
            />
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
      </div>

      {failureMessage && (isError || isDisconnected) && (
        <div className="ml-6 flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-3">
          <span
            className={
              isDisconnected ? "text-red-700" : "text-amber-700"
            }
          >
            {failureMessage}
          </span>
          {showAttemptSubline && triedRel && (
            <span className="text-xs text-muted-foreground">
              Tried {triedRel}
              {successRel ? ` · Last success ${successRel}` : ""}
            </span>
          )}
          <Button
            variant={isDisconnected ? "default" : "outline"}
            size="sm"
            onClick={() =>
              reauthorize.mutate(window.location.origin + "/app/calendar")
            }
            disabled={reauthorize.isPending}
            className="sm:ml-auto"
          >
            {isDisconnected ? "Reconnect" : "Reconnect anyway"}
          </Button>
        </div>
      )}

      <DisconnectDialog
        open={showDisconnect}
        onOpenChange={setShowDisconnect}
        connectionId={connection.id}
        email={attrs.email}
      />
    </div>
  );
}
