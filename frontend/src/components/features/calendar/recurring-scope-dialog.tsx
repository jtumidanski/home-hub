import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

interface RecurringScopeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  action: "edit" | "delete";
  onSelect: (scope: "single" | "all") => void;
}

export function RecurringScopeDialog({ open, onOpenChange, action, onSelect }: RecurringScopeDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{action === "edit" ? "Edit Recurring Event" : "Delete Recurring Event"}</DialogTitle>
        </DialogHeader>
        <div className="space-y-2">
          <Button variant="outline" className="w-full justify-start" onClick={() => onSelect("single")}>
            This event only
          </Button>
          <Button variant="outline" className="w-full justify-start" onClick={() => onSelect("all")}>
            All events in the series
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
