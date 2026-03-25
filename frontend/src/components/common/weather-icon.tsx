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

export function WeatherIcon({ icon, ...props }: WeatherIconProps) {
  const IconComponent = iconMap[icon] ?? Cloud;
  return <IconComponent {...props} />;
}
