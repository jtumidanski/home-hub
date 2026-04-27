import { AlertTriangle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface UnknownWidgetPlaceholderProps {
  type: string;
}

export function UnknownWidgetPlaceholder({ type }: UnknownWidgetPlaceholderProps) {
  return (
    <Card className="h-full border-dashed">
      <CardHeader className="flex flex-row items-center gap-2 pb-2">
        <AlertTriangle className="h-4 w-4 text-muted-foreground" />
        <CardTitle className="text-sm font-medium">Unknown widget</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-xs text-muted-foreground">
          This dashboard references a widget type (<code>{type}</code>) that is not
          available in this app version.
        </p>
      </CardContent>
    </Card>
  );
}
