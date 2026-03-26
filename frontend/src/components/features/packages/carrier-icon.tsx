import { CARRIER_LABELS } from "@/types/models/package";

interface CarrierIconProps {
  carrier: string;
  className?: string;
}

export function CarrierIcon({ carrier, className }: CarrierIconProps) {
  const label = CARRIER_LABELS[carrier] ?? carrier.toUpperCase();

  const colors: Record<string, string> = {
    usps: "bg-blue-600",
    ups: "bg-amber-700",
    fedex: "bg-purple-600",
  };

  return (
    <span
      className={`inline-flex items-center justify-center rounded px-1.5 py-0.5 text-[10px] font-bold text-white ${colors[carrier] ?? "bg-gray-500"} ${className ?? ""}`}
    >
      {label}
    </span>
  );
}
