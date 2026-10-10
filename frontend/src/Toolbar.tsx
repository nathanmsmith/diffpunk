import type { Layout } from './types';

interface Props {
  repository: string;
  repositoryPath?: string;
  layout: Layout;
  untrackedAvailable?: boolean;
  showUntracked: boolean;
  onUntrackedChange: (show: boolean) => void;
  onLayoutChange: (layout: Layout) => void;
}

export function Toolbar({
  repository,
  repositoryPath,
  layout,
  onLayoutChange,
  untrackedAvailable,
  showUntracked,
  onUntrackedChange,
}: Props) {
  return (
    <header className="toolbar">
      <a className="brand" href="#">
        diff punk
      </a>
      <div className="repository" id="repository" title={repositoryPath}>
        {repository}
      </div>
      {untrackedAvailable && (
        <button
          id="show-untracked"
          className="untracked-toggle"
          aria-pressed={showUntracked}
          onClick={() => onUntrackedChange(!showUntracked)}
        >
          Show untracked files
        </button>
      )}
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
