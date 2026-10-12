import type { ComponentProps } from 'react';
import { cn } from './cn';

interface Props extends ComponentProps<'button'> {
  variant?: 'default' | 'layout';
}

export function Button({
  variant = 'default',
  className = '',
  ...props
}: Props) {
  const styles =
    variant === 'layout'
      ? 'border-0 bg-transparent px-2.5 py-1.5 text-muted aria-pressed:bg-surface aria-pressed:text-fg aria-pressed:shadow-sm mobile:px-2 mobile:text-xs'
      : 'border border-solid border-border bg-surface px-2.5 py-1.5 text-fg hover:bg-hover';

  return (
    <button
      {...props}
      className={cn('cursor-pointer rounded-md', styles, className)}
    />
  );
}
