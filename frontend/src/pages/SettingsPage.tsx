import { useCallback, useMemo, useState } from "react";
import { useAuth } from "@/components/providers/auth-provider";
import { useThemeToggle } from "@/lib/hooks/use-theme-toggle";
import { useUpdateMe } from "@/lib/hooks/api/use-auth";
import { Button } from "@/components/ui/button";
import { ErrorCard } from "@/components/common/error-card";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { UserAvatar } from "@/components/ui/user-avatar";
import { Moon, Sun, Shuffle, RotateCcw, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

const AVATAR_STYLES = ["adventurer", "bottts", "fun-emoji"] as const;
type AvatarStyle = (typeof AVATAR_STYLES)[number];

const STYLE_LABELS: Record<AvatarStyle, string> = {
  adventurer: "Adventurer",
  bottts: "Robots",
  "fun-emoji": "Emoji",
};

function generateSeeds(count: number, offset: number): string[] {
  return Array.from({ length: count }, (_, i) => `seed${offset + i}`);
}

function AvatarPicker({
  currentAvatarUrl,
  onSelect,
  isPending,
}: {
  currentAvatarUrl: string;
  onSelect: (avatarUrl: string) => void;
  isPending: boolean;
}) {
  const [activeStyle, setActiveStyle] = useState<AvatarStyle>("adventurer");
  const [seedOffset, setSeedOffset] = useState(0);

  const seeds = useMemo(() => generateSeeds(12, seedOffset), [seedOffset]);

  const handleShuffle = useCallback(() => {
    setSeedOffset((prev) => prev + 12);
  }, []);

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        {AVATAR_STYLES.map((style) => (
          <Button
            key={style}
            variant={activeStyle === style ? "default" : "outline"}
            size="sm"
            onClick={() => setActiveStyle(style)}
          >
            {STYLE_LABELS[style]}
          </Button>
        ))}
      </div>

      <div className="grid grid-cols-4 gap-3 sm:grid-cols-6">
        {seeds.map((seed) => {
          const descriptor = `dicebear:${activeStyle}:${seed}`;
          const isSelected = currentAvatarUrl === descriptor;
          return (
            <button
              key={seed}
              type="button"
              disabled={isPending}
              onClick={() => onSelect(descriptor)}
              className={cn(
                "rounded-lg border-2 p-1.5 transition-colors hover:border-primary",
                isSelected ? "border-primary bg-primary/10" : "border-transparent",
              )}
            >
              <img
                src={`https://api.dicebear.com/9.x/${activeStyle}/svg?seed=${seed}`}
                alt={`${STYLE_LABELS[activeStyle]} avatar ${seed}`}
                className="h-12 w-12 rounded-md"
              />
            </button>
          );
        })}
      </div>

      <Button variant="outline" size="sm" onClick={handleShuffle}>
        <Shuffle className="mr-2 h-4 w-4" />
        Shuffle
      </Button>
    </div>
  );
}

function SettingsPageSkeleton() {
  return (
    <div className="p-4 md:p-6 space-y-6" role="status" aria-label="Loading">
      <Skeleton className="h-8 w-32" />
      <Skeleton className="h-40 w-full" />
      <Skeleton className="h-24 w-full" />
    </div>
  );
}

export function SettingsPage() {
  const { user, appContext, isLoading } = useAuth();
  const { theme, toggleTheme } = useThemeToggle();
  const updateMe = useUpdateMe();

  const handleSelectAvatar = useCallback(
    (avatarUrl: string) => {
      updateMe.mutate(
        { avatarUrl },
        {
          onSuccess: () => toast.success("Avatar updated"),
          onError: () => toast.error("Failed to update avatar"),
        },
      );
    },
    [updateMe],
  );

  const handleResetToProvider = useCallback(() => {
    updateMe.mutate(
      { avatarUrl: "" },
      {
        onSuccess: () => toast.success("Avatar reset to provider image"),
        onError: () => toast.error("Failed to reset avatar"),
      },
    );
  }, [updateMe]);

  const handleRemoveAvatar = useCallback(() => {
    updateMe.mutate(
      { avatarUrl: "" },
      {
        onSuccess: () => toast.success("Avatar removed"),
        onError: () => toast.error("Failed to remove avatar"),
      },
    );
  }, [updateMe]);

  if (isLoading) {
    return <SettingsPageSkeleton />;
  }

  if (!user || !appContext) {
    return (
      <div className="p-4 md:p-6">
        <ErrorCard message="Failed to load settings. Try refreshing the page." />
      </div>
    );
  }

  const { avatarUrl, providerAvatarUrl } = user.attributes;
  const hasUserSelectedAvatar = avatarUrl.startsWith("dicebear:");
  const hasProviderAvatar = !!providerAvatarUrl;

  return (
    <div className="p-4 md:p-6 space-y-6">
      <h1 className="text-xl md:text-2xl font-semibold">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Your account information</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2">
          <div><span className="text-sm font-medium">Name:</span> <span className="text-sm">{user.attributes.displayName}</span></div>
          <div><span className="text-sm font-medium">Email:</span> <span className="text-sm">{user.attributes.email}</span></div>
          <div><span className="text-sm font-medium">Role:</span> <span className="text-sm">{appContext.attributes.resolvedRole}</span></div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Avatar</CardTitle>
          <CardDescription>Choose how you appear across Home Hub</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center gap-4">
            <UserAvatar
              avatarUrl={avatarUrl}
              providerAvatarUrl={providerAvatarUrl}
              displayName={user.attributes.displayName}
              userId={user.id}
              size="lg"
            />
            <div className="flex flex-col gap-2">
              {hasUserSelectedAvatar && hasProviderAvatar && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleResetToProvider}
                  disabled={updateMe.isPending}
                >
                  <RotateCcw className="mr-2 h-4 w-4" />
                  Reset to provider image
                </Button>
              )}
              {(hasUserSelectedAvatar || hasProviderAvatar) && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleRemoveAvatar}
                  disabled={updateMe.isPending}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Remove avatar
                </Button>
              )}
            </div>
          </div>

          <AvatarPicker
            currentAvatarUrl={avatarUrl}
            onSelect={handleSelectAvatar}
            isPending={updateMe.isPending}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>Customize how Home Hub looks</CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" onClick={toggleTheme}>
            {theme === "light" ? <Moon className="mr-2 h-4 w-4" /> : <Sun className="mr-2 h-4 w-4" />}
            {theme === "light" ? "Switch to Dark Mode" : "Switch to Light Mode"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
