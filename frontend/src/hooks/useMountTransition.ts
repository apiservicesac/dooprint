import { useEffect, useState } from "react";

/**
 * Replacement for Vue's <transition> around a v-if: keeps the node mounted
 * while its leave transition plays, and delays the enter classes by a frame so
 * the browser has a "from" state to animate away from.
 */
export function useMountTransition(show: boolean, leaveDuration = 150) {
  const [mounted, setMounted] = useState(show);
  const [visible, setVisible] = useState(show);

  useEffect(() => {
    if (show) {
      setMounted(true);

      let inner = 0;
      const outer = requestAnimationFrame(() => {
        inner = requestAnimationFrame(() => setVisible(true));
      });

      return () => {
        cancelAnimationFrame(outer);
        cancelAnimationFrame(inner);
      };
    }

    setVisible(false);
    const timeout = setTimeout(() => setMounted(false), leaveDuration);
    return () => clearTimeout(timeout);
  }, [show, leaveDuration]);

  return { mounted, visible };
}
