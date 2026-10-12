import type { ComponentProps } from 'react';
import { cn } from 'cn';

interface Props extends ComponentProps<'button'> {
  variant?: 'default' | 'layout';
}

export function Button({
  variant = 'default',
  className = '',
  ...props
}: Props) {
  return (
    <button
      {...props}
      className={cn(
        'cursor-pointer',
        'rounded-md',
        variant === 'layout' ? 'border-0' : 'border',
        variant === 'default' && 'border-solid',
        variant === 'default' && 'border-border',
        variant === 'layout' ? 'bg-transparent' : 'bg-surface',
        'px-2.5',
        'py-1.5',
        variant === 'layout' ? 'text-muted' : 'text-fg',
        variant === 'layout' && 'aria-pressed:bg-surface',
        variant === 'layout' && 'aria-pressed:text-fg',
        variant === 'layout' && 'aria-pressed:shadow-sm',
        variant === 'layout' && 'mobile:px-2',
        variant === 'layout' && 'mobile:text-xs',
        variant === 'default' && 'hover:bg-hover',
        className,
      )}
    />
  );
}
