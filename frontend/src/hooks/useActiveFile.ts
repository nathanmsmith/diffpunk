import { useEffect, useState } from 'react';
import type { RefObject } from 'react';
import type { FileSnapshot } from '../types';

export function useActiveFile(
  main: RefObject<HTMLElement | null>,
  files?: FileSnapshot[],
) {
  const [activeFile, setActiveFile] = useState<string>();

  useEffect(() => {
    const root = main.current;
    if (!root) return;
    const sections = Array.from(
      root.querySelectorAll<HTMLElement>('section[data-path]'),
    );
    let frame: number | undefined;

    const update = () => {
      frame = undefined;
      const top = root.getBoundingClientRect().top + root.clientTop;
      let current = sections[0]?.dataset.path;
      for (const section of sections) {
        if (section.getBoundingClientRect().top > top + 1) break;
        current = section.dataset.path;
      }
      setActiveFile(current);
    };
    const schedule = () => {
      if (frame === undefined) frame = requestAnimationFrame(update);
    };

    const observer = new IntersectionObserver(schedule, { root });
    const resizeObserver = new ResizeObserver(schedule);
    resizeObserver.observe(root);
    for (const section of sections) {
      observer.observe(section);
      resizeObserver.observe(section);
    }
    // Scroll updates locate the sticky boundary even within a tall section,
    // or when a fast scroll jumps past an entire file or the gap after it.
    root.addEventListener('scroll', schedule, { passive: true });
    schedule();

    return () => {
      observer.disconnect();
      resizeObserver.disconnect();
      root.removeEventListener('scroll', schedule);
      if (frame !== undefined) cancelAnimationFrame(frame);
    };
  }, [main, files]);

  return activeFile;
}
