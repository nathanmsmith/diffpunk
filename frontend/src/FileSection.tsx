import { cn } from 'cn';
import { useMemo, useState } from 'react';
import { hydratePartialDiff, processFile } from '@pierre/diffs';
import { Button } from './Button';
import { File, FileDiff } from '@pierre/diffs/react';
import type { FileSnapshot, ViewOptions } from './types';

interface Props {
  file: FileSnapshot;
  options: ViewOptions;
}

export function FileSection({ file, options }: Props) {
  const patch = useMemo(() => {
    try {
      return processFile(file.patch, {
        throwOnError: true,
      });
    } catch (error) {
      console.error(error);
      return undefined;
    }
  }, [file.patch, file.name]);
  const [expanded, setExpanded] = useState(false);
  const [review, setReview] = useState({ patch: file.patch, viewed: false });
  if (review.patch !== file.patch) {
    setReview({ patch: file.patch, viewed: false });
  }
  const viewed = review.patch === file.patch && review.viewed;
  const { full, expandablePatch } = useMemo(() => {
    if (!patch || !file.fullPatch) return {};
    try {
      const full = processFile(file.fullPatch, {
        throwOnError: true,
      });
      if (!full) return {};
      const expandablePatch = patch.isPartial
        ? hydratePartialDiff('clone', patch, {
            oldFile: {
              name: file.previousName ?? file.name,
              contents: full.deletionLines.join(''),
            },
            newFile: {
              name: file.name,
              contents: full.additionLines.join(''),
            },
          })
        : patch;
      return { full, expandablePatch };
    } catch (error) {
      console.error(error);
      return {};
    }
  }, [patch, file.fullPatch, file.name, file.previousName]);
  const [failedPatch, setFailedPatch] = useState<string>();
  const expansionFailed =
    failedPatch !== undefined && failedPatch === (file.fullPatch ?? '');
  const hasDiff = patch !== undefined && patch.hunks.length > 0;
  const canExpand = hasDiff
    ? Boolean(file.fullPatch && file.fullPatch !== file.patch)
    : file.fullContent !== undefined;
  const contents = useMemo(
    () => ({
      name: file.name,
      contents: file.fullContent ?? '',
    }),
    [file.name, file.fullContent],
  );

  function toggleExpanded() {
    try {
      if (hasDiff && !expanded && !full) {
        throw new Error('Full-file context is unavailable');
      }
      setExpanded(!expanded);
    } catch (error) {
      setFailedPatch(file.fullPatch ?? '');
      console.error(error);
    }
  }

  return (
    <section
      className={cn(
        'group',
        'mb-6',
        'scroll-mt-24',
        'overflow-clip',
        'rounded-lg',
        'border',
        'border-solid',
        'border-border',
        'bg-surface',
      )}
      id={`file-${file.name}`}
      data-path={file.name}
      data-viewed={viewed}
    >
      <div
        className={cn(
          'file-header',
          'sticky',
          'top-0',
          'z-1',
          'flex',
          'items-center',
          'gap-3',
          'border-0',
          'border-b',
          'border-solid',
          'border-border',
          'bg-surface',
          'px-4',
          'py-3',
          'group-viewed:border-b-0',
          'mobile:gap-2',
          'mobile:p-2.5',
        )}
      >
        <h2
          className={cn(
            'm-0',
            'flex-1',
            'font-mono',
            'text-sm',
            'font-semibold',
            'wrap-anywhere',
          )}
        >
          {file.previousName
            ? `${file.previousName} → ${file.name}`
            : file.name}
        </h2>
        <span className={cn('text-xs', 'text-muted')}>{file.kind}</span>
        {canExpand && (
          <Button
            className={cn('shrink-0', 'text-xs')}
            aria-expanded={expanded}
            disabled={expansionFailed}
            onClick={toggleExpanded}
          >
            {expansionFailed
              ? 'Expansion unavailable'
              : expanded
                ? 'Show changes only'
                : 'Show full file'}
          </Button>
        )}
        <Button
          className={cn(
            'mark-viewed',
            'shrink-0',
            'text-xs',
            'aria-pressed:border-accent',
            'aria-pressed:text-accent',
          )}
          type="button"
          aria-pressed={viewed}
          aria-label={viewed ? 'Mark as unviewed' : 'Mark as viewed'}
          onClick={() => setReview({ patch: file.patch, viewed: !viewed })}
        >
          {viewed ? '✓ Viewed' : 'Mark as viewed'}
        </Button>
      </div>
      <div hidden={viewed}>
        {hasDiff ? (
          <FileDiff
            className={cn('diff-host')}
            fileDiff={expanded && full ? full : (expandablePatch ?? patch)}
            options={options}
          />
        ) : (
          <>
            <div
              className={cn('p-6', 'text-center', 'text-muted')}
              hidden={expanded}
            >
              {file.patch.includes('Binary files ') ||
              file.patch.includes('GIT binary patch')
                ? 'Binary file changed'
                : 'File metadata changed'}
              <pre
                className={cn(
                  'mb-0',
                  'mt-3',
                  'whitespace-pre-wrap',
                  'text-left',
                  'font-mono',
                  'text-sm',
                  'leading-relaxed',
                  'text-fg',
                  'wrap-anywhere',
                )}
              >
                {file.patch}
              </pre>
            </div>
            {expanded && (
              <File
                className={cn('diff-host')}
                file={contents}
                options={options}
              />
            )}
          </>
        )}
      </div>
    </section>
  );
}
