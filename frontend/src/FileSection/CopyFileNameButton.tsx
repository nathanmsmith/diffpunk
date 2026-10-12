import { cn } from 'cn';
import { useEffect, useState } from 'react';
import { Button } from '../Button';

interface Props {
  name: string;
}

export function CopyFileNameButton({ name }: Props) {
  const [copyStatus, setCopyStatus] = useState<'idle' | 'copied' | 'failed'>(
    'idle',
  );
  useEffect(() => {
    if (copyStatus === 'idle') return;
    const timeout = window.setTimeout(() => setCopyStatus('idle'), 2000);
    return () => window.clearTimeout(timeout);
  }, [copyStatus]);

  async function copyFileName() {
    try {
      await navigator.clipboard.writeText(name);
      setCopyStatus('copied');
    } catch (error) {
      console.error(error);
      setCopyStatus('failed');
    }
  }

  return (
    <Button
      className={cn('shrink-0', 'text-xs')}
      type="button"
      aria-label={`Copy file name: ${name}`}
      title={
        copyStatus === 'copied'
          ? 'Copied!'
          : copyStatus === 'failed'
            ? 'Copy failed'
            : 'Copy file name'
      }
      onClick={copyFileName}
    >
      <svg
        className={cn('block', 'size-4')}
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        {copyStatus === 'copied' ? (
          <path d="m5 12 4 4L19 6" />
        ) : copyStatus === 'failed' ? (
          <path d="m6 6 12 12M6 18 18 6" />
        ) : (
          <>
            <rect x="9" y="9" width="13" height="13" rx="2" />
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
          </>
        )}
      </svg>
      <span className={cn('sr-only')} role="status" aria-live="polite">
        {copyStatus === 'copied'
          ? 'Copied!'
          : copyStatus === 'failed'
            ? 'Copy failed'
            : 'Copy name'}
      </span>
    </Button>
  );
}
