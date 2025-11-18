'use client';

import { useState, useRef, useEffect, ReactNode } from 'react';
import { RefreshCw } from 'lucide-react';

interface PullToRefreshProps {
  onRefresh: () => Promise<void>;
  children: ReactNode;
}

export function PullToRefresh({ onRefresh, children }: PullToRefreshProps) {
  const [pullDistance, setPullDistance] = useState(0);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [startY, setStartY] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);

  const threshold = 80; // Minimum pull distance to trigger refresh
  const maxPull = 150; // Maximum visual pull distance

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let touchStartY = 0;
    let scrollTop = 0;

    const handleTouchStart = (e: TouchEvent) => {
      // Only start tracking if we're at the top of the scroll
      const target = e.target as HTMLElement;
      const scrollableParent = findScrollableParent(target);
      scrollTop = scrollableParent ? scrollableParent.scrollTop : window.scrollY;

      if (scrollTop === 0 && !isRefreshing) {
        touchStartY = e.touches[0].clientY;
        setStartY(touchStartY);
      }
    };

    const handleTouchMove = (e: TouchEvent) => {
      if (isRefreshing || touchStartY === 0) return;

      const scrollableParent = findScrollableParent(e.target as HTMLElement);
      const currentScrollTop = scrollableParent ? scrollableParent.scrollTop : window.scrollY;

      // Only allow pull if still at top
      if (currentScrollTop > 0) {
        touchStartY = 0;
        setPullDistance(0);
        return;
      }

      const currentY = e.touches[0].clientY;
      const diff = currentY - touchStartY;

      if (diff > 0) {
        // Prevent default scroll behavior
        e.preventDefault();

        // Apply rubber band effect (diminishing returns as you pull further)
        const pull = Math.min(diff * 0.5, maxPull);
        setPullDistance(pull);
      }
    };

    const handleTouchEnd = async () => {
      if (isRefreshing || touchStartY === 0) return;

      if (pullDistance >= threshold) {
        setIsRefreshing(true);
        try {
          await onRefresh();
        } finally {
          setIsRefreshing(false);
          setPullDistance(0);
        }
      } else {
        setPullDistance(0);
      }

      touchStartY = 0;
      setStartY(0);
    };

    const findScrollableParent = (element: HTMLElement | null): HTMLElement | null => {
      if (!element) return null;

      const style = window.getComputedStyle(element);
      const overflowY = style.overflowY;

      if (overflowY === 'auto' || overflowY === 'scroll') {
        return element;
      }

      return findScrollableParent(element.parentElement);
    };

    container.addEventListener('touchstart', handleTouchStart, { passive: true });
    container.addEventListener('touchmove', handleTouchMove, { passive: false });
    container.addEventListener('touchend', handleTouchEnd, { passive: true });

    return () => {
      container.removeEventListener('touchstart', handleTouchStart);
      container.removeEventListener('touchmove', handleTouchMove);
      container.removeEventListener('touchend', handleTouchEnd);
    };
  }, [pullDistance, isRefreshing, onRefresh, threshold, maxPull, startY]);

  const showIndicator = pullDistance > 0 || isRefreshing;
  const triggerRefresh = pullDistance >= threshold;

  return (
    <div ref={containerRef} className="relative">
      {/* Pull Indicator */}
      {showIndicator && (
        <div
          className="absolute top-0 left-0 right-0 flex items-center justify-center transition-opacity"
          style={{
            transform: `translateY(${isRefreshing ? 60 : Math.min(pullDistance, maxPull)}px)`,
            opacity: isRefreshing ? 1 : Math.min(pullDistance / threshold, 1),
            height: 60,
            marginTop: -60,
          }}
        >
          <div
            className={`flex items-center gap-2 px-4 py-2 bg-card border border-border rounded-full shadow-lg ${
              isRefreshing ? 'animate-pulse' : ''
            }`}
          >
            <RefreshCw
              className={`h-5 w-5 text-blue-600 dark:text-blue-500 ${
                isRefreshing || triggerRefresh ? 'animate-spin' : ''
              }`}
            />
            <span className="text-sm font-medium text-card-foreground">
              {isRefreshing
                ? 'Refreshing...'
                : triggerRefresh
                ? 'Release to refresh'
                : 'Pull to refresh'}
            </span>
          </div>
        </div>
      )}

      {/* Content */}
      <div
        style={{
          transform: `translateY(${isRefreshing ? 60 : pullDistance}px)`,
          transition: isRefreshing || pullDistance === 0 ? 'transform 0.3s ease' : 'none',
        }}
      >
        {children}
      </div>
    </div>
  );
}
