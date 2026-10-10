import { useMemo, useState } from 'react';
import { processFile, type FileDiffMetadata } from '@pierre/diffs';
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
      className="file"
      id={`file-${index}`}
      data-path={file.name}
      data-viewed={viewed}
    >
      <div className="file-header">
        <h2 className="file-name">
          {file.previousName
            ? `${file.previousName} → ${file.name}`
            : file.name}
        </h2>
        <span className="file-kind">
          {file.kind === 'untracked' ? 'Untracked' : file.kind}
        </span>
        {canExpand && (
          <button
            aria-expanded={expanded}
            disabled={expansionFailed}
            onClick={toggleExpanded}
          >
            {expansionFailed
              ? 'Expansion unavailable'
              : expanded
                ? 'Show changes only'
                : 'Show full file'}
          </button>
        )}
        <button
          className="mark-viewed"
          type="button"
          aria-pressed={viewed}
          aria-label={viewed ? 'Mark as unviewed' : 'Mark as viewed'}
          onClick={() => setViewed(!viewed)}
        >
          {viewed ? '✓ Viewed' : 'Mark as viewed'}
        </button>
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
            <div className="file-summary" hidden={expanded}>
              {file.patch.includes('Binary files ') ||
              file.patch.includes('GIT binary patch')
                ? 'Binary file changed'
                : 'File metadata changed'}
              <pre>{file.patch}</pre>
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
