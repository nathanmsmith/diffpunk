import { useLayoutEffect, useRef } from 'react';
import { Button } from './Button';
import type { Layout } from './types';

interface Props {
  repository: string;
  repositoryPath?: string;
  layout: Layout;
  onLayoutChange: (layout: Layout) => void;
}

export function Topnav({
  repository,
  repositoryPath,
  layout,
  onLayoutChange,
}: Props) {
  const toolbarRef = useRef<HTMLElement>(null);

  useLayoutEffect(() => {
    const toolbar = toolbarRef.current;
    if (!toolbar) return;

    const updateHeight = () => {
      document.documentElement.style.setProperty(
        '--toolbar-height',
        `${toolbar.getBoundingClientRect().height}px`,
      );
    };
    updateHeight();
    const observer = new ResizeObserver(updateHeight);
    observer.observe(toolbar);
    return () => {
      observer.disconnect();
      document.documentElement.style.removeProperty('--toolbar-height');
    };
  }, []);

  return (
    <header
      className="sticky top-0 z-10 flex h-toolbar items-center gap-6 border-0 border-b border-solid border-border bg-surface px-6 mobile:h-auto mobile:min-h-toolbar mobile:flex-wrap mobile:gap-3 mobile:px-3 mobile:py-2.5"
      ref={toolbarRef}
    >
      <a
        className="text-brand font-brand tracking-brand text-fg no-underline"
        href="#"
      >
        diff punk
      </a>
      <div
        className="flex-1 truncate text-muted mobile:min-w-repository mobile:text-meta"
        id="repository"
        title={repositoryPath}
      >
        {repository}
      </div>
      <div
        className="flex shrink-0 gap-0.5 rounded-panel border border-solid border-border bg-bg p-0.75"
        role="group"
        aria-label="Diff layout"
      >
        {(['split', 'unified'] as const).map((mode) => (
          <Button
            variant="layout"
            key={mode}
            id={mode}
            aria-pressed={mode === layout}
            onClick={() => onLayoutChange(mode)}
          >
            {mode === 'split' ? 'Side-by-side' : 'Unified'}
          </Button>
        ))}
      </div>
    </header>
  );
}
