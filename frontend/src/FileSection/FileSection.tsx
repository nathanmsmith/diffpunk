import { cn } from 'cn';
import { useMemo, useState } from 'react';
import { hydratePartialDiff, processFile } from '@pierre/diffs';
import { FileSectionHeader } from './FileSectionHeader';
import { File, FileDiff } from '@pierre/diffs/react';
import type { ProcessedFileSnapshot, ViewOptions } from '../types';

interface Props {
  file: ProcessedFileSnapshot;
  options: ViewOptions;
}

export function FileSection({ file, options }: Props) {
  const patch = file.diff;
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
      <FileSectionHeader
        file={file}
        canExpand={canExpand}
        expanded={expanded}
        expansionFailed={expansionFailed}
        viewed={viewed}
        onToggleExpanded={toggleExpanded}
        onToggleViewed={() => setReview({ patch: file.patch, viewed: !viewed })}
      />
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
