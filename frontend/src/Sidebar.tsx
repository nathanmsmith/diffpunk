import { useState } from 'react';
import { cn } from './cn';
import type { Snapshot } from './types';

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
  return (
    <aside className="min-h-0 w-60 shrink-0 overflow-auto border-0 border-r border-solid border-border px-3 py-6 mobile:w-full mobile:border-r-0 mobile:border-b mobile:py-3 mobile:shrink">
      <div className="flex justify-between px-2.5 pb-4 text-xs font-semibold tracking-widest text-muted">
        CHANGED FILES <span id="file-count">{snapshot?.files.length}</span>
      </div>
      <nav id="files" aria-label="Changed files">
        {snapshot?.files.map((file) => (
          <a
            key={file.name}
            href={`#${encodeURIComponent(`file-${file.name}`)}`}
            className={cn(
              'flex items-start gap-2 rounded-md px-2.5 py-2 text-sm text-fg no-underline hover:bg-hover',
              activeFile === file.name && 'active bg-hover',
            )}
            onClick={() => setActiveFile(file.name)}
          >
            <span className="shrink-0 font-mono text-muted">
              {statuses[file.kind] ?? 'M'}
            </span>
            <span className="wrap-anywhere">{file.name}</span>
          </a>
        ))}
      </nav>
      <div
        className="mx-2.5 my-6 whitespace-pre-wrap text-xs leading-relaxed text-muted"
        id="snapshot-time"
      >
        {snapshot &&
          `${snapshot.live ? 'Live' : 'Snapshot'} · ${new Date(snapshot.createdAt).toLocaleTimeString()}\n git diff ${snapshot.arguments.join(' ')}`}
      </div>
    </aside>
  );
}
