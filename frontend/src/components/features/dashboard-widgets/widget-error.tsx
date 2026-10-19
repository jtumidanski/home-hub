import { AlertTriangle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

/**
 * Shared error body for dashboard widgets: icon plus text so status is never
 * conveyed by colour alone.
 */
export function WidgetError({ message }: { message: string }) {
  return (
    <Card className="@container h-full border-destructive">
      <CardContent role="alert" className="flex items-center gap-2 py-3">
        <AlertTriangle aria-hidden="true" className="size-5 shrink-0 text-destructive @xl:size-6" />
        <p className="text-sm @xl:text-xl text-foreground">{message}</p>
      </CardContent>
    </Card>
  );
}
