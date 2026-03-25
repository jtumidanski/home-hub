import { Card, CardContent } from "@/components/ui/card";

interface ErrorCardProps {
  message: string;
}

export function ErrorCard({ message }: ErrorCardProps) {
  return (
    <Card className="border-destructive">
      <CardContent className="py-3">
        <p className="text-sm text-destructive">{message}</p>
      </CardContent>
    </Card>
  );
}
