import { useEffect, useState } from 'react';
import type { Snapshot } from '../types';

type LoadState =
  | { status: 'loading' }
  | { status: 'loaded'; snapshot: Snapshot }
  | { status: 'error'; message: string };

export function useSnapshot() {
  const [load, setLoad] = useState<LoadState>({ status: 'loading' });
  const [refreshError, setRefreshError] = useState<string>();

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

  return { load, refreshError };
}
