import { Button } from './Button';
import { cn } from './cn';
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
  return (
    <header
      className={cn(
        'sticky',
        'top-0',
        'z-10',
        'flex',
        'shrink-0',
        'items-center',
        'h-18',
        'px-6',
        'mobile:px-3',
        'gap-6',
        'mobile:gap-3',
        'border-0',
        'border-b',
        'border-solid',
        'border-border',
        'bg-surface',
      )}
    >
      <h1
        className={cn(
          'shrink-0',
          'text-xl',
          'leading-none',
          'font-extrabold',
          'tracking-tight',
          'text-fg',
          'no-underline',
          'font-mono',
        )}
      >
        diff
        <br />
        punk
      </h1>
      <div
        className={cn(
          'min-w-0',
          'flex-1',
          'truncate',
          'text-muted',
          'mobile:text-xs',
        )}
        id="repository"
        title={repositoryPath}
      >
        {repository}
      </div>
      <div
        className={cn(
          'flex',
          'shrink-0',
          'gap-0.5',
          'rounded-lg',
          'border',
          'border-solid',
          'border-border',
          'bg-bg',
          'p-0.75',
          'whitespace-nowrap',
        )}
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
