import {
  Sun,
  CloudSun,
  Cloud,
  CloudFog,
  CloudDrizzle,
  CloudRain,
  Snowflake,
  CloudLightning,
  type LucideProps,
} from "lucide-react";
import type { FC } from "react";
import { cn } from "@/lib/utils";

const colorMap: Record<string, string> = {
  sun: "text-chart-3",
  "cloud-sun": "text-chart-3",
  cloud: "text-chart-4",
  "cloud-fog": "text-chart-4",
  "cloud-drizzle": "text-chart-4",
  "cloud-rain": "text-chart-4",
  snowflake: "text-chart-4",
  "cloud-lightning": "text-chart-1",
};

const iconMap: Record<string, FC<LucideProps>> = {
  sun: Sun,
  "cloud-sun": CloudSun,
  cloud: Cloud,
  "cloud-fog": CloudFog,
  "cloud-drizzle": CloudDrizzle,
  "cloud-rain": CloudRain,
  snowflake: Snowflake,
  "cloud-lightning": CloudLightning,
};

interface WeatherIconProps extends LucideProps {
  icon: string;
}

/** Decorative by default; pass `aria-label` to expose it as a named image. */
export function WeatherIcon({ icon, className, ...props }: WeatherIconProps) {
  const IconComponent = iconMap[icon] ?? Cloud;
  const labelled = props["aria-label"] !== undefined;
  return (
    <IconComponent
      aria-hidden={labelled ? undefined : true}
      role={labelled ? "img" : undefined}
      className={cn(colorMap[icon] ?? "text-chart-4", className)}
      {...props}
    />
  );
}
