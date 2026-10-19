import { useCallback, useMemo, useState } from "react";
import { useAuth } from "@/components/providers/auth-provider";
import { useThemeToggle } from "@/lib/hooks/use-theme-toggle";
import { useUpdateMe } from "@/lib/hooks/api/use-auth";
import { Button } from "@/components/ui/button";
import { ErrorCard } from "@/components/common/error-card";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { UserAvatar } from "@/components/ui/user-avatar";
import { Check, Moon, Sun, Shuffle, RotateCcw, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { PageHeader } from "@/components/common/page-header";
import { membershipRoleLabelMap } from "@/types/models/membership";
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
      <div className="flex gap-2" role="group" aria-label="Avatar style">
        {AVATAR_STYLES.map((style) => (
          <Button
            key={style}
            variant={activeStyle === style ? "default" : "outline"}
            size="sm"
            aria-pressed={activeStyle === style}
            onClick={() => setActiveStyle(style)}
          >
            {STYLE_LABELS[style]}
          </Button>
        ))}
      </div>

      <div className="grid grid-cols-4 gap-3 sm:grid-cols-6">
        {seeds.map((seed, index) => {
          const descriptor = `dicebear:${activeStyle}:${seed}`;
          const isSelected = currentAvatarUrl === descriptor;
          return (
            <button
              key={seed}
              type="button"
              disabled={isPending}
              aria-pressed={isSelected}
              aria-label={`Choose avatar ${index + 1}`}
              onClick={() => onSelect(descriptor)}
              className={cn(
                "relative rounded-lg border-2 p-1.5 transition-colors hover:border-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-50",
                isSelected ? "border-primary bg-primary/10" : "border-transparent",
              )}
            >
              <img
                src={`https://api.dicebear.com/9.x/${activeStyle}/svg?seed=${seed}`}
                alt=""
                width={48}
                height={48}
                loading="lazy"
                className="h-12 w-12 rounded-md"
              />
              {isSelected && (
                <span className="absolute right-0.5 top-0.5 flex size-4 items-center justify-center rounded-full bg-primary text-primary-foreground">
                  <Check aria-hidden="true" className="size-3" />
                </span>
              )}
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
      <PageHeader title="Settings" />
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
  const resolvedRole = appContext.attributes.resolvedRole;
  const roleLabel = membershipRoleLabelMap[resolvedRole];

  return (
    <div className="p-4 md:p-6 space-y-6">
      <PageHeader title="Settings" />

      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Your account information</CardDescription>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
            <dt className="text-muted-foreground">Name</dt>
            <dd>{user.attributes.displayName}</dd>
            <dt className="text-muted-foreground">Email</dt>
            <dd>{user.attributes.email}</dd>
            <dt className="text-muted-foreground">Role</dt>
            <dd>{roleLabel}</dd>
          </dl>
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
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <p className="text-sm font-medium">Theme</p>
              <p className="text-sm text-muted-foreground">
                Current theme: {theme === "light" ? "Light" : "Dark"}
              </p>
            </div>
            <Button variant="outline" onClick={toggleTheme}>
              {theme === "light" ? <Moon className="mr-2 h-4 w-4" /> : <Sun className="mr-2 h-4 w-4" />}
              {theme === "light" ? "Switch to Dark Mode" : "Switch to Light Mode"}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Data retention</CardTitle>
          <CardDescription>Configure how long Home Hub keeps each kind of data.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" render={<a href="/app/settings/data-retention" />}>
            Open data retention settings
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
