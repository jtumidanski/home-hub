import { AlertTriangle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface UnknownWidgetPlaceholderProps {
  type: string;
}

export function UnknownWidgetPlaceholder({ type }: UnknownWidgetPlaceholderProps) {
  return (
    <Card className="@container h-full border-dashed">
      <CardHeader className="flex flex-row items-center gap-2 pb-2">
        <AlertTriangle aria-hidden="true" className="size-5 @xl:size-6 text-warning-text" />
        <CardTitle className="font-heading text-lg @xl:text-2xl">Unknown widget</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm @xl:text-xl text-muted-foreground">
          This dashboard references a widget type (<code>{type}</code>) that is not
          available in this app version.
        </p>
      </CardContent>
    </Card>
  );
}
