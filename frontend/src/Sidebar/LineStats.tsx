import { cn } from 'cn';
interface Props {
  additions?: number;
  deletions?: number;
  binary?: boolean;
}

export function LineStats({ additions, deletions, binary }: Props) {
  if (binary) {
    return (
      <span className={cn('shrink-0', 'text-xs', 'text-muted')}>Binary</span>
    );
  }

  if (additions === undefined || deletions === undefined) {
    return (
      <span
        className={cn('shrink-0', 'text-xs', 'text-muted')}
        title="Line counts unavailable"
      >
        —
      </span>
    );
  }

  return (
    <span
      className={cn(
        'flex',
        'shrink-0',
        'gap-2',
        'font-mono',
        'text-xs',
        'tabular-nums',
      )}
      aria-label={`${additions} lines added, ${deletions} lines deleted`}
    >
      <span className={cn('text-addition')}>+{additions}</span>
      <span className={cn('text-deletion')}>−{deletions}</span>
    </span>
  );
}
