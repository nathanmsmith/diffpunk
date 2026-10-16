import { useEffect, useMemo, useRef } from 'react';
import { cn } from 'cn';
import type { ProcessedSnapshot } from '../types';
import { LineStats } from './LineStats';
import { fileId } from '../fileId';

interface Props {
  snapshot?: ProcessedSnapshot;
  activeFile?: string;
  includeUntracked: boolean;
  untrackedCount: number;
  onIncludeUntrackedChange: (include: boolean) => void;
}

const statuses: Record<string, string> = {
  added: 'A',
  deleted: 'D',
  renamed: 'R',
  copied: 'C',
  untracked: '?',
};

export function Sidebar({
  snapshot,
  activeFile,
  includeUntracked,
  untrackedCount,
  onIncludeUntrackedChange,
}: Props) {
  const sidebar = useRef<HTMLElement>(null);
  useEffect(() => {
    const container = sidebar.current;
    const active = container?.querySelector<HTMLElement>('[aria-current]');
    if (!container || !active) return;
    const bounds = container.getBoundingClientRect();
    const row = active.getBoundingClientRect();
    const top = bounds.top + container.clientTop;
    const bottom = top + container.clientHeight;
    if (row.top < top) container.scrollTop += row.top - top;
    else if (row.bottom > bottom) container.scrollTop += row.bottom - bottom;
  }, [activeFile, snapshot?.files]);
  const { files, additions, deletions } = useMemo(() => {
    let additions = 0;
    let deletions = 0;
    const files = (snapshot?.files ?? []).map((file) => {
      const binary = /^(Binary files |GIT binary patch\r?$)/m.test(file.patch);
      if (binary) return { file, binary };
      const patch = file.diff;
      if (!patch) return { file, binary };
      const stats = patch.hunks.reduce(
        (total, hunk) => ({
          additions: total.additions + hunk.additionLines,
          deletions: total.deletions + hunk.deletionLines,
        }),
        { additions: 0, deletions: 0 },
      );
      additions += stats.additions;
      deletions += stats.deletions;
      return { file, binary, stats };
    });
    return { files, additions, deletions };
  }, [snapshot?.files]);

  return (
    <aside
      ref={sidebar}
      id="sidebar"
      className={cn(
        'min-h-0',
        'min-w-0',
        'w-full',
        'flex-1',
        'overflow-auto',
        'border-0',
        'border-r',
        'border-solid',
        'border-border',
        'px-3',
        'py-6',
        'mobile:w-full',
        'mobile:border-r-0',
        'mobile:border-b',
        'mobile:py-3',
        'mobile:shrink',
      )}
    >
      <div className={cn('px-2.5', 'pb-4')}>
        {snapshot?.repository && (
          <div
            className={cn(
              'mb-4',
              'font-mono',
              'text-sm',
              'text-muted',
              'wrap-anywhere',
            )}
            id="repository"
            title={snapshot.repository}
          >
            {snapshot.repositoryDisplay ?? snapshot.repository}
          </div>
        )}
        <div
          className={cn(
            'flex',
            'justify-between',
            'text-xs',
            'font-semibold',
            'tracking-widest',
            'text-muted',
          )}
        >
          FILES <span id="file-count">{snapshot?.files.length}</span>
        </div>
        <div
          className={cn(
            'flex',
            'justify-between',
            'text-xs',
            'font-semibold',
            'tracking-widest',
            'text-muted',
          )}
        >
          LINES <LineStats additions={additions} deletions={deletions} />
        </div>
        {snapshot?.untracked && (
          <label
            className={cn(
              'mt-4',
              'flex',
              'cursor-pointer',
              'items-center',
              'justify-between',
              'gap-3',
              'text-sm',
              'text-fg',
            )}
          >
            <span
              className={cn(
                'group',
                'relative',
                'inline-flex',
                'w-11',
                'shrink-0',
                'rounded-full',
                'bg-border',
                'p-0.5',
                'inset-ring',
                'inset-ring-fg/5',
                'outline-offset-2',
                'outline-accent',
                'transition-colors',
                'duration-200',
                'ease-in-out',
                'has-checked:bg-accent',
                'has-focus-visible:outline-2',
              )}
            >
              <span
                aria-hidden="true"
                className={cn(
                  'size-5',
                  'rounded-full',
                  'bg-white',
                  'shadow-xs',
                  'ring-1',
                  'ring-black/5',
                  'transition-transform',
                  'duration-200',
                  'ease-in-out',
                  'group-has-checked:translate-x-5',
                )}
              />
              <input
                type="checkbox"
                role="switch"
                aria-labelledby="include-untracked-label"
                checked={includeUntracked}
                onChange={(event) =>
                  onIncludeUntrackedChange(event.target.checked)
                }
                className={cn(
                  'absolute',
                  'inset-0',
                  'size-full',
                  'cursor-pointer',
                  'appearance-none',
                  'focus:outline-hidden',
                )}
              />
            </span>
            <span id="include-untracked-label" className={cn('font-medium')}>
              Include {untrackedCount} untracked{' '}
              {untrackedCount === 1 ? 'file' : 'files'}
            </span>
          </label>
        )}
      </div>
      <nav id="files" aria-label="Changed files">
        {files.map(({ file, binary, stats }) => (
          <a
            key={fileId(file)}
            href={`#${encodeURIComponent(fileId(file))}`}
            aria-current={activeFile === fileId(file) ? 'location' : undefined}
            className={cn(
              'flex',
              'flex-col',
              'items-start',
              'rounded-md',
              'px-2.5',
              'py-2',
              'text-sm',
              'text-fg',
              'no-underline',
              'transition-colors',
              'duration-200',
              'ease-out',
              'hover:bg-hover',
              activeFile === fileId(file) && 'active',
              activeFile === fileId(file) && 'bg-hover',
            )}
          >
            <div className={cn('min-w-0', 'flex-1', 'wrap-anywhere')}>
              {file.name}
            </div>
            <div className={cn('flex', 'items-baseline', 'gap-2')}>
              <span className={cn('font-mono', 'text-xs', 'text-muted')}>
                {statuses[file.kind] ?? 'M'}
              </span>
              <LineStats binary={binary} {...stats} />
            </div>
          </a>
        ))}
      </nav>
    </aside>
  );
}
