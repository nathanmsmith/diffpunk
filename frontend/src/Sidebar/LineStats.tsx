interface Props {
  additions: number;
  deletions: number;
}

export function LineStats({ additions, deletions }: Props) {
  return (
    <span
      className="flex shrink-0 gap-2 font-mono text-xs tabular-nums"
      aria-label={`${additions} lines added, ${deletions} lines deleted`}
    >
      <span className="text-addition">+{additions}</span>
      <span className="text-deletion">−{deletions}</span>
    </span>
  );
}
