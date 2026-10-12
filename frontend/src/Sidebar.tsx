import { useState } from 'react';
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
    <aside className="sticky top-toolbar h-sidebar overflow-auto border-0 border-r border-solid border-border px-3 py-6 mobile:hidden">
      <div className="flex justify-between px-2.5 pb-4 text-label font-label tracking-label text-muted">
        CHANGED FILES <span id="file-count">{snapshot?.files.length}</span>
      </div>
      <nav id="files" aria-label="Changed files">
        {snapshot?.files.map((file, index) => (
          <a
            key={file.name}
            href={`#file-${index}`}
            className={`flex items-start gap-2 rounded-control px-2.5 py-2 text-body text-fg no-underline hover:bg-hover ${activeFile === file.name ? 'active bg-hover' : ''}`}
            onClick={() => setActiveFile(file.name)}
          >
            <span className="shrink-0 font-status text-muted">
              {statuses[file.kind] ?? 'M'}
            </span>
            <span className="wrap-anywhere">{file.name}</span>
          </a>
        ))}
      </nav>
      <div
        className="mx-2.5 my-6 whitespace-pre-wrap text-meta leading-relaxed text-muted"
        id="snapshot-time"
      >
        {snapshot &&
          `Snapshot · ${new Date(snapshot.createdAt).toLocaleTimeString()}\n git diff ${snapshot.arguments.join(' ')}`}
      </div>
    </aside>
  );
}
