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
    <aside>
      <div className="sidebar-label">
        CHANGED FILES <span id="file-count">{snapshot?.files.length}</span>
      </div>
      <nav id="files" aria-label="Changed files">
        {snapshot?.files.map((file, index) => (
          <a
            key={file.name}
            href={`#file-${index}`}
            className={activeFile === file.name ? 'active' : undefined}
            onClick={() => setActiveFile(file.name)}
          >
            <span className="status">{statuses[file.kind] ?? 'M'}</span>
            <span className="name">{file.name}</span>
          </a>
        ))}
      </nav>
      <div className="snapshot-label" id="snapshot-time">
        {snapshot &&
          `Snapshot · ${new Date(snapshot.createdAt).toLocaleTimeString()}\n git diff ${snapshot.arguments.join(' ')}`}
      </div>
    </aside>
  );
}
