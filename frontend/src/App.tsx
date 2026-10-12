import { useMemo, useState } from 'react';
import { FileSection } from './FileSection';
import { Sidebar } from './Sidebar';
import { Topnav } from './Topnav';
import type { Layout, ViewOptions } from './types';
import { useSnapshot } from './hooks/useSnapshot';

export function App() {
  const { load, refreshError } = useSnapshot();
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

  const snapshot = load.status === 'loaded' ? load.snapshot : undefined;
  const repository = snapshot
    ? (snapshot.repository.split('/').filter(Boolean).at(-1) ??
      snapshot.repository)
    : '';

  return (
    <div className="flex min-h-dvh flex-col">
      <Topnav
        repository={repository}
        repositoryPath={snapshot?.repository}
        layout={layout}
        onLayoutChange={setLayout}
      />
      <div className="flex flex-1 mobile:flex-col">
        <Sidebar snapshot={snapshot} />
        <main className="min-w-0 flex-1 p-6 mobile:p-3">
          {refreshError && (
            <div className="p-6 text-muted" role="status">
              {refreshError} Keeping the last diff; retrying automatically.
            </div>
          )}
          {load.status === 'loading' && (
            <div className="p-6 text-muted" id="loading">
              Loading diff…
            </div>
          )}
          {load.status === 'error' && (
            <div className="p-6 text-muted">{load.message}</div>
          )}
          {snapshot?.files.map((file) => (
            <FileSection key={file.name} file={file} options={options} />
          ))}
          {snapshot?.files.length === 0 && (
            <div className="p-6 text-muted">No changes.</div>
          )}
        </main>
      </div>
    </div>
  );
}
