import { cn } from 'cn';
import { useMemo, useRef, useState } from 'react';
import { FileSection } from './FileSection/FileSection';
import { Sidebar } from './Sidebar/Sidebar';
import { SidebarPanel } from './Sidebar/SidebarPanel';
import { Header } from './Header';
import type { Layout, ViewOptions } from './types';
import { useSnapshot } from './hooks/useSnapshot';
import { useActiveFile } from './hooks/useActiveFile';
import { fileId } from './fileId';

export function App() {
  const { load, refreshError } = useSnapshot();
  const [untrackedChoice, setUntrackedChoice] = useState<boolean>();
  const [layout, setLayout] = useState<Layout>(() =>
    matchMedia('(max-width: 800px)').matches ? 'unified' : 'split',
  );
  const options = useMemo<ViewOptions>(
    () => ({
      diffStyle: layout,
      theme: { light: 'pierre-light', dark: 'pierre-dark' },
      disableFileHeader: true,
      overflow: 'scroll',
    }),
    [layout],
  );

  const loadedSnapshot = load.status === 'loaded' ? load.snapshot : undefined;
  const untrackedCount =
    loadedSnapshot?.files.filter((file) => file.kind === 'untracked').length ??
    0;
  const includeUntracked =
    untrackedChoice ?? loadedSnapshot?.untracked?.defaultIncluded ?? false;
  const snapshot = useMemo(
    () =>
      loadedSnapshot && {
        ...loadedSnapshot,
        files: loadedSnapshot.files.filter(
          (file) => file.kind !== 'untracked' || includeUntracked,
        ),
      },
    [loadedSnapshot, includeUntracked],
  );
  const main = useRef<HTMLElement>(null);
  const activeFile = useActiveFile(main, snapshot?.files);

  return (
    <div className={cn('flex', 'h-dvh', 'flex-col')}>
      <Header layout={layout} onLayoutChange={setLayout} />
      <div className={cn('flex', 'min-h-0', 'flex-1', 'mobile:flex-col')}>
        <SidebarPanel>
          <Sidebar
            snapshot={snapshot}
            activeFile={activeFile}
            includeUntracked={includeUntracked}
            untrackedCount={untrackedCount}
            onIncludeUntrackedChange={setUntrackedChoice}
          />
        </SidebarPanel>
        <main
          ref={main}
          className={cn(
            'min-h-0',
            'min-w-0',
            'flex-1',
            'overflow-auto',
            'px-6',
            'mobile:px-3',
          )}
        >
          <div className={cn('py-6', 'mobile:py-3')}>
            {refreshError && (
              <div
                className={cn('p-6', 'text-center', 'text-muted')}
                role="status"
              >
                {refreshError} Keeping the last diff; retrying automatically.
              </div>
            )}
            {load.status === 'loading' && (
              <div
                className={cn('p-6', 'text-center', 'text-muted')}
                id="loading"
              >
                Loading diff…
              </div>
            )}
            {load.status === 'error' && (
              <div className={cn('p-6', 'text-center', 'text-muted')}>
                {load.message}
              </div>
            )}
            {snapshot?.files.map((file) => (
              <FileSection key={fileId(file)} file={file} options={options} />
            ))}
            {snapshot?.files.length === 0 && (
              <div className={cn('p-6', 'text-center', 'text-muted')}>
                No changes.
              </div>
            )}
          </div>
        </main>
      </div>
    </div>
  );
}
