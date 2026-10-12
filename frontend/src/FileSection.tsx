import { useMemo, useState } from 'react';
import { processFile, type FileDiffMetadata } from '@pierre/diffs';
import { Button } from './Button';
import { File, FileDiff } from '@pierre/diffs/react';
import type { FileSnapshot, ViewOptions } from './types';

interface Props {
  file: FileSnapshot;
  index: number;
  options: ViewOptions;
}

export function FileSection({ file, index, options }: Props) {
  const patch = useMemo(() => {
    try {
      return processFile(file.patch, {
        throwOnError: true,
        cacheKey: `patch-${file.name}`,
      });
    } catch (error) {
      console.error(error);
      return undefined;
    }
  }, [file.patch, file.name]);
  const [expanded, setExpanded] = useState(false);
  const [viewed, setViewed] = useState(false);
  const [full, setFull] = useState<FileDiffMetadata>();
  const [expansionFailed, setExpansionFailed] = useState(false);
  const hasDiff = patch !== undefined && patch.hunks.length > 0;
  const canExpand = hasDiff
    ? Boolean(file.fullPatch && file.fullPatch !== file.patch)
    : file.fullContent !== undefined;
  const contents = useMemo(
    () => ({
      name: file.name,
      contents: file.fullContent ?? '',
      cacheKey: `metadata-${file.name}`,
    }),
    [file.name, file.fullContent],
  );

  function toggleExpanded() {
    try {
      if (hasDiff && !expanded && !full) {
        const parsed = processFile(file.fullPatch!, {
          throwOnError: true,
          cacheKey: `full-${file.name}`,
        });
        if (!parsed) throw new Error('Full-file context is unavailable');
        setFull(parsed);
      }
      setExpanded(!expanded);
    } catch (error) {
      setExpansionFailed(true);
      console.error(error);
    }
  }

  return (
    <section
      className="group mb-6 scroll-mt-24 overflow-clip rounded-lg border border-solid border-border bg-surface"
      id={`file-${index}`}
      data-path={file.name}
      data-viewed={viewed}
    >
      <div className="file-header sticky top-18 z-1 flex items-center gap-3 border-0 border-b border-solid border-border bg-surface px-4 py-3 group-viewed:border-b-0 mobile:gap-2 mobile:p-2.5">
        <h2 className="m-0 flex-1 font-mono text-sm font-semibold wrap-anywhere">
          {file.previousName
            ? `${file.previousName} → ${file.name}`
            : file.name}
        </h2>
        <span className="text-xs text-muted">{file.kind}</span>
        {canExpand && (
          <Button
            className="shrink-0 text-xs"
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
          className="mark-viewed shrink-0 text-xs aria-pressed:border-accent aria-pressed:text-accent"
          type="button"
          aria-pressed={viewed}
          aria-label={viewed ? 'Mark as unviewed' : 'Mark as viewed'}
          onClick={() => setViewed(!viewed)}
        >
          {viewed ? '✓ Viewed' : 'Mark as viewed'}
        </Button>
      </div>
      <div hidden={viewed}>
        {hasDiff ? (
          <FileDiff
            className="diff-host"
            fileDiff={expanded && full ? full : patch}
            options={options}
          />
        ) : (
          <>
            <div className="p-6 text-muted" hidden={expanded}>
              {file.patch.includes('Binary files ') ||
              file.patch.includes('GIT binary patch')
                ? 'Binary file changed'
                : 'File metadata changed'}
              <pre className="mb-0 mt-3 whitespace-pre-wrap font-mono text-sm leading-relaxed text-fg wrap-anywhere">
                {file.patch}
              </pre>
            </div>
            {expanded && (
              <File className="diff-host" file={contents} options={options} />
            )}
          </>
        )}
      </div>
    </section>
  );
}
