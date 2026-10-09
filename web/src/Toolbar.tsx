import type { Layout } from './types';

export function Toolbar({
  repository,
  repositoryPath,
  layout,
  onLayoutChange,
}: {
  repository: string;
  repositoryPath?: string;
  layout: Layout;
  onLayoutChange: (layout: Layout) => void;
}) {
  return (
    <header className="toolbar">
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
