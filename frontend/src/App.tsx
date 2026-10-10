import { useEffect, useMemo, useState } from 'react';
import { FileSection } from './FileSection';
import { Sidebar } from './Sidebar';
import { Toolbar } from './Toolbar';
import type { Layout, Snapshot, ViewOptions } from './types';

type LoadState =
  | { status: 'loading' }
  | { status: 'loaded'; snapshot: Snapshot }
  | { status: 'error'; message: string };

export function App() {
  const [load, setLoad] = useState<LoadState>({ status: 'loading' });
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

  useEffect(() => {
    const controller = new AbortController();
    async function loadSnapshot() {
      try {
        const response = await fetch('./snapshot.json', {
          signal: controller.signal,
        });
        if (!response.ok)
          throw new Error(`Could not load snapshot (${response.status})`);
        const snapshot: Snapshot = await response.json();
        if (!controller.signal.aborted) setLoad({ status: 'loaded', snapshot });
      } catch (error) {
        if (!controller.signal.aborted)
          setLoad({
            status: 'error',
            message:
              error instanceof Error ? error.message : 'Could not load diff.',
          });
      }
    }
    void loadSnapshot();
    return () => controller.abort();
  }, []);

  const snapshot = load.status === 'loaded' ? load.snapshot : undefined;
  const repository = snapshot
    ? (snapshot.repository.split('/').filter(Boolean).at(-1) ??
      snapshot.repository)
    : '';
  useEffect(() => {
    if (snapshot) document.title = `${repository} · diff punk`;
  }, [snapshot, repository]);

  return (
    <>
      <Toolbar
        repository={repository}
        repositoryPath={snapshot?.repository}
        layout={layout}
        onLayoutChange={setLayout}
      />
      <div className="workspace">
        <Sidebar snapshot={snapshot} />
        <main id="diffs" aria-label="Diffs">
          {load.status === 'loading' && (
            <div className="empty" id="loading">
              Loading diff…
            </div>
          )}
          {load.status === 'error' && (
            <div className="empty">{load.message}</div>
          )}
          {snapshot?.files.map((file, index) => (
            <FileSection
              key={file.name}
              file={file}
              index={index}
              options={options}
            />
          ))}
          {snapshot?.files.length === 0 && (
            <div className="empty">No changes.</div>
          )}
        </main>
      </div>
    </>
  );
}
