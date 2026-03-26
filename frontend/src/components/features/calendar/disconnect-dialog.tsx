import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useDisconnectCalendar } from "@/lib/hooks/api/use-calendar";
import { getErrorMessage } from "@/lib/api/errors";

interface DisconnectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  connectionId: string;
  email: string;
}

export function DisconnectDialog({ open, onOpenChange, connectionId, email }: DisconnectDialogProps) {
  const disconnect = useDisconnectCalendar();

  const handleDisconnect = async () => {
    try {
      await disconnect.mutateAsync(connectionId);
      toast.success("Calendar disconnected");
      onOpenChange(false);
    } catch (error) {
      toast.error(getErrorMessage(error, "Failed to disconnect"));
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Disconnect Calendar</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">
          This will disconnect <strong>{email}</strong> and remove all synced events from the household calendar. You can reconnect at any time.
        </p>
        <div className="flex justify-end gap-2 pt-4">
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={handleDisconnect} disabled={disconnect.isPending}>
            {disconnect.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Disconnect
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
