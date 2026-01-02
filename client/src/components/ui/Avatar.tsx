import { type HTMLAttributes, forwardRef } from 'react';
import { cn } from '@/lib/utils';

interface AvatarProps extends HTMLAttributes<HTMLDivElement> {
  src?: string | null;
  name: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

/**
 * Industrial Avatar Component
 * Bold initials with construction brand colors
 */
const Avatar = forwardRef<HTMLDivElement, AvatarProps>(
  ({ src, name, size = 'md', className, ...props }, ref) => {
    const initials = name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);

    const sizes = {
      sm: 'w-8 h-8 text-xs',
      md: 'w-10 h-10 text-sm',
      lg: 'w-12 h-12 text-base',
      xl: 'w-16 h-16 text-lg',
    };

    return (
      <div
        ref={ref}
        className={cn(
          'relative flex items-center justify-center rounded-xl overflow-hidden',
          'bg-gradient-to-br from-charcoal-700 to-charcoal-800',
          'dark:from-charcoal-600 dark:to-charcoal-700',
          'text-white font-bold tracking-wide',
          'border-2 border-charcoal-600 dark:border-charcoal-500',
          'shadow-md',
          sizes[size],
          className
        )}
        {...props}
      >
        {src ? (
          <img
            src={src}
            alt={name}
            className="w-full h-full object-cover"
          />
        ) : (
          <span className="font-display">{initials}</span>
        )}
        {/* Online indicator slot */}
        <div className="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-emerald-500 rounded-full border-2 border-white dark:border-charcoal-900 hidden" />
      </div>
    );
  }
);

Avatar.displayName = 'Avatar';

export { Avatar };
