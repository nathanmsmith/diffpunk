import { useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { cn } from 'cn';

interface Props {
  children: ReactNode;
}

const storageKey = 'diffpunk.sidebar-width';
const minWidth = 160;
const mobileQuery = '(max-width: 800px)';

function clamp(width: number, max: number) {
  return Math.min(max, Math.max(minWidth, width));
}

function initialSize() {
  const max = Math.max(minWidth, window.innerWidth / 2);
  const defaultWidth =
    15 * parseFloat(getComputedStyle(document.documentElement).fontSize);
  let width = defaultWidth;
  try {
    const saved = Number(localStorage.getItem(storageKey));
    if (Number.isFinite(saved) && saved >= minWidth) width = saved;
  } catch {
    // Resizing still works when browser storage is unavailable.
  }
  return {
    width: matchMedia(mobileQuery).matches ? width : clamp(width, max),
    max,
  };
}

export function SidebarPanel({ children }: Props) {
  const panel = useRef<HTMLDivElement>(null);
  const drag = useRef<{ pointerId: number; x: number; width: number } | null>(
    null,
  );
  const [size, setSize] = useState(initialSize);
  const [resizing, setResizing] = useState(false);

  useEffect(() => {
    const container = panel.current?.parentElement;
    if (!container) return;
    const media = matchMedia(mobileQuery);
    const updateBounds = () => {
      if (media.matches) {
        drag.current = null;
        setResizing(false);
        return;
      }
      const max = Math.max(
        minWidth,
        container.getBoundingClientRect().width / 2,
      );
      setSize((current) => ({ width: clamp(current.width, max), max }));
    };
    const observer = new ResizeObserver(updateBounds);
    observer.observe(container);
    media.addEventListener('change', updateBounds);
    updateBounds();
    return () => {
      observer.disconnect();
      media.removeEventListener('change', updateBounds);
    };
  }, []);

  useEffect(() => {
    try {
      localStorage.setItem(storageKey, String(size.width));
    } catch {
      // Resizing still works when browser storage is unavailable.
    }
  }, [size.width]);

  useEffect(() => {
    if (!resizing) return;
    const classes = [
      'cursor-col-resize!',
      '**:cursor-col-resize!',
      'select-none!',
      '**:select-none!',
    ].filter((name) => !document.body.classList.contains(name));
    document.body.classList.add(...classes);
    return () => document.body.classList.remove(...classes);
  }, [resizing]);

  const stopDragging = () => {
    drag.current = null;
    setResizing(false);
  };

  return (
    <div
      ref={panel}
      className={cn(
        'relative',
        'flex',
        'shrink-0',
        'min-h-0',
        'mobile:w-full!',
        'mobile:shrink',
      )}
      style={{ width: size.width }}
    >
      {children}
      <div
        className={cn(
          'absolute',
          'top-0',
          '-right-1',
          'bottom-0',
          'z-10',
          'w-2',
          'cursor-col-resize',
          'touch-none',
          'mobile:hidden',
          'outline-none',
        )}
        role="separator"
        tabIndex={0}
        aria-label="Resize sidebar"
        aria-controls="sidebar"
        aria-orientation="vertical"
        aria-valuemin={minWidth}
        aria-valuemax={Math.floor(size.max)}
        aria-valuenow={Math.round(size.width)}
        aria-valuetext={`${Math.round(size.width)} pixels`}
        onPointerDown={(event) => {
          if (!event.isPrimary || event.button !== 0) return;
          event.preventDefault();
          event.currentTarget.setPointerCapture(event.pointerId);
          drag.current = {
            pointerId: event.pointerId,
            x: event.clientX,
            width: size.width,
          };
          setResizing(true);
        }}
        onPointerMove={(event) => {
          const current = drag.current;
          if (!current || current.pointerId !== event.pointerId) return;
          const width = current.width + event.clientX - current.x;
          setSize((previous) => ({
            ...previous,
            width: clamp(width, previous.max),
          }));
        }}
        onPointerUp={(event) => {
          if (drag.current?.pointerId !== event.pointerId) return;
          event.currentTarget.releasePointerCapture(event.pointerId);
          stopDragging();
        }}
        onPointerCancel={stopDragging}
        onLostPointerCapture={stopDragging}
        onKeyDown={(event) => {
          const step = event.shiftKey ? 50 : 10;
          let width: number;
          switch (event.key) {
            case 'ArrowLeft':
              width = size.width - step;
              break;
            case 'ArrowRight':
              width = size.width + step;
              break;
            case 'Home':
              width = minWidth;
              break;
            case 'End':
              width = size.max;
              break;
            default:
              return;
          }
          event.preventDefault();
          setSize((current) => ({
            ...current,
            width: clamp(width, current.max),
          }));
        }}
      />
    </div>
  );
}
