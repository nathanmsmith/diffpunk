import { useLayoutEffect, useRef } from 'react';
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
    <header className="toolbar" ref={toolbarRef}>
      <a className="brand" href="#">
        diff punk
      </a>
      <div className="repository" id="repository" title={repositoryPath}>
        {repository}
      </div>
      <div className="layouts" role="group" aria-label="Diff layout">
        {(['split', 'unified'] as const).map((mode) => (
          <button
            key={mode}
            id={mode}
            aria-pressed={mode === layout}
            onClick={() => onLayoutChange(mode)}
          >
            {mode === 'split' ? 'Side-by-side' : 'Unified'}
          </button>
        ))}
      </div>
    </header>
  );
}
