import { type ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description?: string;
  action?: ReactNode;
  className?: string;
}

export function EmptyState({ icon, title, description, action, className }: EmptyStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-12 px-4 text-center', className)}>
      {icon && (
        <div className="w-16 h-16 rounded-full bg-charcoal-100 dark:bg-charcoal-800 flex items-center justify-center mb-4 text-charcoal-400 dark:text-charcoal-500">
          {icon}
        </div>
      )}
      <h3 className="font-display text-xl tracking-wide text-charcoal-900 dark:text-charcoal-100 mb-2">
        {title}
      </h3>
      {description && (
        <p className="text-charcoal-500 dark:text-charcoal-400 max-w-sm mb-6">
          {description}
        </p>
      )}
      {action}
    </div>
  );
}
