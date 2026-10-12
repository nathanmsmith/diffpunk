import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
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
  const anchor = useRef<{ path: string; top: number } | undefined>(undefined);
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
          const sections = document.querySelectorAll<HTMLElement>(
            '#diffs section[data-path]',
          );
          const headerBottom =
            document.querySelector('header')?.getBoundingClientRect().bottom ??
            0;
          const current = Array.from(sections).find(
            (section) => section.getBoundingClientRect().bottom > headerBottom,
          );
          anchor.current = current
            ? {
                path: current.dataset.path!,
                top: current.getBoundingClientRect().top,
              }
            : undefined;
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

  useLayoutEffect(() => {
    const saved = anchor.current;
    anchor.current = undefined;
    if (!saved) return;
    const section = document.getElementById(`file-${saved.path}`);
    if (!section) return;
    const restore = () =>
      window.scrollBy(0, section.getBoundingClientRect().top - saved.top);
    restore();
    const frame = requestAnimationFrame(restore);
    // Syntax highlighting can change heights after React commits the update.
    const observer = new ResizeObserver(restore);
    const diffs = document.getElementById('diffs');
    if (diffs) observer.observe(diffs);
    observer.observe(section);
    const stop = () => {
      observer.disconnect();
      cancelAnimationFrame(frame);
      window.removeEventListener('wheel', stop);
      window.removeEventListener('touchstart', stop);
      window.removeEventListener('pointerdown', stop);
      window.removeEventListener('keydown', stop);
    };
    window.addEventListener('wheel', stop, { passive: true });
    window.addEventListener('touchstart', stop, { passive: true });
    window.addEventListener('pointerdown', stop);
    window.addEventListener('keydown', stop);
    const timeout = setTimeout(stop, 1000);
    return () => {
      clearTimeout(timeout);
      stop();
    };
  }, [load]);

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
      <Topnav
        repository={repository}
        repositoryPath={snapshot?.repository}
        layout={layout}
        onLayoutChange={setLayout}
      />
      <div className="workspace grid grid-cols-workspace mobile:grid-cols-1">
        <Sidebar snapshot={snapshot} />
        <main className="min-w-0 p-6 mobile:p-3" id="diffs" aria-label="Diffs">
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
    </>
  );
}
