import * as React from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'purple';
  size?: 'sm' | 'md';
}

export function Badge({
  className,
  variant = 'default',
  size = 'sm',
  children,
  ...props
}: BadgeProps) {
  const variants = {
    default: 'bg-zinc-800 text-zinc-300 border-zinc-700/60',
    success: 'bg-emerald-950/60 text-emerald-400 border-emerald-800/50',
    warning: 'bg-amber-950/60 text-amber-400 border-amber-800/50',
    danger: 'bg-rose-950/60 text-rose-400 border-rose-800/50',
    info: 'bg-sky-950/60 text-sky-400 border-sky-800/50',
    purple: 'bg-violet-950/60 text-violet-400 border-violet-800/50',
  };

  const sizes = {
    sm: 'text-[11px] px-2 py-0.5 rounded',
    md: 'text-xs px-2.5 py-1 rounded-md',
  };

  return (
    <div
      className={twMerge(
        clsx(
          'inline-flex items-center font-medium border select-none tracking-wide',
          variants[variant],
          sizes[size],
          className
        )
      )}
      {...props}
    >
      {children}
    </div>
  );
}
