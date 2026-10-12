import type { ComponentProps } from 'react';

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
      ? 'border-0 bg-transparent px-2.5 py-control-y text-muted aria-pressed:bg-surface aria-pressed:text-fg aria-pressed:shadow-layout mobile:px-control-mobile-x mobile:text-meta'
      : 'border border-solid border-border bg-surface px-2.5 py-control-y text-fg hover:bg-hover';

  return (
    <button
      {...props}
      className={`cursor-pointer rounded-control ${styles} ${className}`}
    />
  );
}
