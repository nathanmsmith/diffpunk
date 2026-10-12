import { useEffect, useMemo, useState } from 'react';
import { FileSection } from './FileSection';
import { Sidebar } from './Sidebar';
import { Topnav } from './Topnav';
import type { Layout, Snapshot, ViewOptions } from './types';

type LoadState =
  | { status: 'loading' }
  | { status: 'loaded'; snapshot: Snapshot }
  | { status: 'error'; message: string };

export function App() {
  const [load, setLoad] = useState<LoadState>({ status: 'loading' });
  const [refreshError, setRefreshError] = useState<string>();
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
    let session: EventSource | undefined;
    let retry: ReturnType<typeof setTimeout> | undefined;
    let fetching = false;
    let pending = false;
    let loaded = false;

    async function loadSnapshot() {
      if (controller.signal.aborted) return;
      if (fetching) {
        pending = true;
        return;
      }
      fetching = true;
      clearTimeout(retry);
      try {
        const response = await fetch('./snapshot.json', {
          signal: controller.signal,
        });
        if (!response.ok)
          throw new Error(`Could not load snapshot (${response.status})`);
        const snapshot: Snapshot = await response.json();
        if (!controller.signal.aborted) {
          loaded = true;
          setLoad({ status: 'loaded', snapshot });
          setRefreshError(snapshot.refreshError);
        }
      } catch (error) {
        if (!controller.signal.aborted) {
          const message =
            error instanceof Error ? error.message : 'Could not load diff.';
          if (loaded) setRefreshError(message);
          else setLoad({ status: 'error', message });
          retry = setTimeout(() => void loadSnapshot(), 1000);
        }
      } finally {
        fetching = false;
        if (pending && !controller.signal.aborted) {
          pending = false;
          void loadSnapshot();
        }
      }
    }

    const connect = () => {
      if (session) return;
      session = new EventSource('./session');
      session.addEventListener('snapshot', () => void loadSnapshot());
      session.addEventListener('open', () => void loadSnapshot());
      session.addEventListener('error', () => {
        if (loaded) setRefreshError('Connection lost. Retrying…');
      });
    };
    const disconnect = () => {
      session?.close();
      session = undefined;
    };
    connect();
    void loadSnapshot();
    window.addEventListener('pagehide', disconnect);
    window.addEventListener('pageshow', connect);
    return () => {
      window.removeEventListener('pagehide', disconnect);
      window.removeEventListener('pageshow', connect);
      controller.abort();
      clearTimeout(retry);
      disconnect();
    };
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
