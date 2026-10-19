import { StatusBanner } from "@/components/ui/status-banner";

interface ErrorCardProps {
  message: string;
}

export function ErrorCard({ message }: ErrorCardProps) {
  return <StatusBanner variant="destructive">{message}</StatusBanner>;
}
