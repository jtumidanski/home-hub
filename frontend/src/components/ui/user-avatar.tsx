import { useState } from "react";
import { cn } from "@/lib/utils";

type AvatarSize = "sm" | "md" | "lg";

const sizeClasses: Record<AvatarSize, string> = {
  sm: "h-8 w-8 text-xs",
  md: "h-10 w-10 text-sm",
  lg: "h-20 w-20 text-xl",
};

interface UserAvatarProps {
  avatarUrl?: string;
  providerAvatarUrl?: string;
  displayName: string;
  userId: string;
  size?: AvatarSize;
  className?: string;
}

function resolveDiceBearUrl(descriptor: string): string {
  const parts = descriptor.split(":");
  if (parts.length !== 3) return "";
  const [, style, seed] = parts;
  return `https://api.dicebear.com/9.x/${style}/svg?seed=${seed}`;
}

function getInitials(displayName: string): string {
  const parts = displayName.trim().split(/\s+/);
  const first = parts[0] ?? "";
  const last = parts.length >= 2 ? (parts[parts.length - 1] ?? "") : "";
  if (first && last) {
    return (first[0]! + last[0]!).toUpperCase();
  }
  return displayName.slice(0, 2).toUpperCase();
}

function getColorFromId(userId: string): string {
  let hash = 0;
  for (let i = 0; i < userId.length; i++) {
    hash = userId.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 65%, 50%)`;
}

export function UserAvatar({
  avatarUrl,
  providerAvatarUrl,
  displayName,
  userId,
  size = "md",
  className,
}: UserAvatarProps) {
  const [imgError, setImgError] = useState(false);

  const imageUrl = avatarUrl?.startsWith("dicebear:")
    ? resolveDiceBearUrl(avatarUrl)
    : avatarUrl || providerAvatarUrl;

  const showImage = imageUrl && !imgError;

  if (showImage) {
    return (
      <img
        src={imageUrl}
        alt={displayName}
        onError={() => setImgError(true)}
        className={cn(
          "shrink-0 rounded-full object-cover",
          sizeClasses[size],
          className,
        )}
      />
    );
  }

  return (
    <div
      role="img"
      aria-label={displayName}
      className={cn(
        "flex shrink-0 items-center justify-center rounded-full font-medium text-white",
        sizeClasses[size],
        className,
      )}
      style={{ backgroundColor: getColorFromId(userId) }}
    >
      {getInitials(displayName)}
    </div>
  );
}
