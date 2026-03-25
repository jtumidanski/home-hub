import { Loader2 } from "lucide-react";
import { useMobile } from "@/lib/hooks/use-mobile";
import { usePullToRefresh } from "@/lib/hooks/use-pull-to-refresh";

interface PullToRefreshProps {
  onRefresh: () => Promise<void>;
  children: React.ReactNode;
}

export function PullToRefresh({ onRefresh, children }: PullToRefreshProps) {
  const isMobile = useMobile();
  const { pullDistance, isRefreshing, handlers } = usePullToRefresh({
    onRefresh,
    enabled: isMobile,
  });

  if (!isMobile) {
    return <>{children}</>;
  }

  return (
    <div
      className="relative min-h-0 flex-1 overflow-auto"
      {...handlers}
    >
      {(pullDistance > 0 || isRefreshing) && (
        <div
          className="flex items-center justify-center transition-[height] duration-150"
          style={{ height: isRefreshing ? 48 : pullDistance }}
        >
          <Loader2
            className="h-5 w-5 animate-spin text-muted-foreground"
            style={{
              opacity: isRefreshing ? 1 : Math.min(pullDistance / 60, 1),
            }}
          />
        </div>
      )}
      {children}
    </div>
  );
}
