import * as React from 'react';
import Link from 'next/link';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export interface EmptyStateActionConfig {
  label: string;
  href?: string;
  onClick?: () => void;
}

export interface EmptyStateProps {
  icon?: React.ReactNode | React.ComponentType<{ size?: number | string; className?: string }>;
  title: string;
  description?: string;
  action?: React.ReactNode | EmptyStateActionConfig;
  className?: string;
}

export function EmptyState({
  icon: IconProp,
  title,
  description,
  action,
  className,
}: EmptyStateProps) {
  let renderedIcon: React.ReactNode = null;
  if (IconProp) {
    if (typeof IconProp === 'function' || (typeof IconProp === 'object' && 'render' in (IconProp as any))) {
      const Component = IconProp as React.ComponentType<{ size?: number | string; className?: string }>;
      renderedIcon = <Component size={32} className="text-text-muted" />;
    } else {
      renderedIcon = IconProp as React.ReactNode;
    }
  }

  let renderedAction: React.ReactNode = null;
  if (action) {
    if (React.isValidElement(action)) {
      renderedAction = action;
    } else if (typeof action === 'object' && 'label' in action) {
      const config = action as EmptyStateActionConfig;
      if (config.href) {
        renderedAction = (
          <Link
            href={config.href}
            className="inline-flex items-center justify-center px-4 py-2 rounded-lg bg-accent text-accent-contrast font-medium text-xs hover:bg-accent-hover transition-colors shadow-xs"
          >
            {config.label}
          </Link>
        );
      } else {
        renderedAction = (
          <button
            type="button"
            onClick={config.onClick}
            className="inline-flex items-center justify-center px-4 py-2 rounded-lg bg-accent text-accent-contrast font-medium text-xs hover:bg-accent-hover transition-colors shadow-xs"
          >
            {config.label}
          </button>
        );
      }
    } else {
      renderedAction = action as React.ReactNode;
    }
  }

  return (
    <div
      className={twMerge(
        clsx(
          'flex flex-col items-center justify-center rounded-xl border border-dashed border-border-default bg-surface-raised/40 px-6 py-12 text-center',
          className
        )
      )}
    >
      {renderedIcon && (
        <div className="mb-3 text-text-muted flex items-center justify-center">
          {renderedIcon}
        </div>
      )}
      <h3 className="text-sm font-semibold text-text-primary">{title}</h3>
      {description && (
        <p className="mt-1 max-w-sm text-xs text-text-muted leading-relaxed">
          {description}
        </p>
      )}
      {renderedAction && <div className="mt-4">{renderedAction}</div>}
    </div>
  );
}
