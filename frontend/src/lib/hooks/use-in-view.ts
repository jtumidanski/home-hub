import { useEffect, useState, type RefCallback } from "react";

/**
 * Reports whether an element has come within `rootMargin` of the viewport.
 * Latches true and stops observing once visible.
 */
export function useInView(rootMargin = "200px"): { ref: RefCallback<Element>; inView: boolean } {
  const [node, setNode] = useState<Element | null>(null);
  const [inView, setInView] = useState(false);

  useEffect(() => {
    if (!node || inView) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((e) => e.isIntersecting)) {
          setInView(true);
          observer.disconnect();
        }
      },
      { rootMargin },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [node, inView, rootMargin]);

  return { ref: setNode, inView };
}
