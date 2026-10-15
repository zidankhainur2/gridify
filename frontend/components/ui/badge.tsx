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
    default: 'bg-surface-elevated text-text-secondary border-border-default',
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
    />
  );
}

export function TournamentStatusBadge({
  status,
  className,
  size = 'sm',
}: {
  status: string;
  className?: string;
  size?: 'sm' | 'md';
}) {
  const norm = (status || '').toUpperCase();

  const sizeClasses = {
    sm: 'text-[10px] px-2.5 py-0.5',
    md: 'text-xs px-3 py-1',
  }[size];

  if (norm === 'ONGOING') {
    return (
      <span
        className={twMerge(
          clsx(
            'inline-flex items-center gap-1.5 rounded-full font-mono font-bold uppercase tracking-wider',
            'bg-emerald-950/60 text-emerald-400 border border-emerald-500/30 shadow-xs',
            sizeClasses,
            className
          )
        )}
      >
        <span className="relative flex h-1.5 w-1.5">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
          <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-emerald-500" />
        </span>
        <span>Ongoing</span>
      </span>
    );
  }

  if (norm === 'COMPLETED') {
    return (
      <span
        className={twMerge(
          clsx(
            'inline-flex items-center gap-1.5 rounded-full font-mono font-bold uppercase tracking-wider',
            'bg-surface-elevated text-text-secondary border border-border-default shadow-xs',
            sizeClasses,
            className
          )
        )}
      >
        <span className="h-1.5 w-1.5 rounded-full bg-text-muted" />
        <span>Selesai</span>
      </span>
    );
  }

  // DRAFT or fallback
  return (
    <span
      className={twMerge(
        clsx(
          'inline-flex items-center gap-1.5 rounded-full font-mono font-bold uppercase tracking-wider',
          'bg-amber-950/60 text-amber-400 border border-amber-500/30 shadow-xs',
          sizeClasses,
          className
        )
      )}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-amber-400" />
      <span>Draft</span>
    </span>
  );
}

