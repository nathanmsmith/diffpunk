import { useMemo, useState } from 'react';
import { processFile } from '@pierre/diffs';
import { cn } from 'cn';
import type { Snapshot } from '../types';
import { LineStats } from './LineStats';

interface Props {
  snapshot?: Snapshot;
}

const statuses: Record<string, string> = {
  added: 'A',
  deleted: 'D',
  renamed: 'R',
  copied: 'C',
};

export function Sidebar({ snapshot }: Props) {
  const [activeFile, setActiveFile] = useState<string>();
  const { files, additions, deletions } = useMemo(() => {
    let additions = 0;
    let deletions = 0;
    const files = (snapshot?.files ?? []).map((file) => {
      const binary = /^(Binary files |GIT binary patch\r?$)/m.test(file.patch);
      if (binary) return { file, binary };
      try {
        const patch = processFile(file.patch, { throwOnError: true });
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
      } catch (error) {
        console.error(error);
        return { file, binary };
      }
    });
    return { files, additions, deletions };
  }, [snapshot?.files]);

  return (
    <aside
      className={cn(
        'min-h-0',
        'w-60',
        'shrink-0',
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
      </div>
      <nav id="files" aria-label="Changed files">
        {files.map(({ file, binary, stats }) => (
          <a
            key={file.name}
            href={`#${encodeURIComponent(`file-${file.name}`)}`}
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
              'hover:bg-hover',
              activeFile === file.name && 'active',
              activeFile === file.name && 'bg-hover',
            )}
            onClick={() => setActiveFile(file.name)}
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
      <div
        className={cn(
          'mx-2.5',
          'my-6',
          'whitespace-pre-wrap',
          'text-xs',
          'leading-relaxed',
          'text-muted',
        )}
        id="snapshot-time"
      >
        {snapshot &&
          `${snapshot.live ? 'Live' : 'Snapshot'} · ${new Date(snapshot.createdAt).toLocaleTimeString()}\n git diff ${snapshot.arguments.join(' ')}`}
      </div>
    </aside>
  );
}
