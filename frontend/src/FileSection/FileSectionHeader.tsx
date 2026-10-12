import { cn } from 'cn';
import { Button } from '../Button';
import type { FileSnapshot } from '../types';
import { CopyFileNameButton } from './CopyFileNameButton';

interface Props {
  file: Pick<FileSnapshot, 'name' | 'previousName' | 'kind'>;
  canExpand: boolean;
  expanded: boolean;
  expansionFailed: boolean;
  viewed: boolean;
  onToggleExpanded: () => void;
  onToggleViewed: () => void;
}

export function FileSectionHeader({
  file,
  canExpand,
  expanded,
  expansionFailed,
  viewed,
  onToggleExpanded,
  onToggleViewed,
}: Props) {
  return (
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
      <div className={cn('flex', 'min-w-0', 'flex-1', 'items-center', 'gap-2')}>
        <h2
          className={cn(
            'm-0',
            'min-w-0',
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
        <CopyFileNameButton name={file.name} />
      </div>
      <span className={cn('text-xs', 'text-muted')}>{file.kind}</span>
      {canExpand && (
        <Button
          className={cn('shrink-0', 'text-xs')}
          aria-expanded={expanded}
          disabled={expansionFailed}
          onClick={onToggleExpanded}
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
        onClick={onToggleViewed}
      >
        {viewed ? '✓ Viewed' : 'Mark as viewed'}
      </Button>
    </div>
  );
}
