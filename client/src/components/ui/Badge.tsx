import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn, getStatusLabel } from "@/lib/utils"
import type { TicketStatus } from "@/types"

/**
 * Industrial Badge Component
 * High contrast for easy visibility on job sites
 */
const badgeVariants = cva(
  [
    "inline-flex items-center justify-center",
    "rounded-full border-2",
    "font-semibold tracking-wide",
    "transition-colors",
    "focus:outline-none focus:ring-2 focus:ring-safety-500 focus:ring-offset-2",
  ].join(" "),
  {
    variants: {
      variant: {
        default: [
          "border-safety-500/30 bg-safety-500",
          "text-white",
        ].join(" "),
        secondary: [
          "border-charcoal-200 dark:border-charcoal-700",
          "bg-charcoal-100 dark:bg-charcoal-800",
          "text-charcoal-700 dark:text-charcoal-300",
        ].join(" "),
        destructive: [
          "border-red-500/30 bg-red-500",
          "text-white",
        ].join(" "),
        outline: [
          "border-charcoal-300 dark:border-charcoal-600",
          "bg-transparent",
          "text-charcoal-700 dark:text-charcoal-300",
        ].join(" "),
        success: [
          "border-emerald-500/30 bg-emerald-500",
          "text-white",
        ].join(" "),
        warning: [
          "border-amber-500/30 bg-amber-500",
          "text-white",
        ].join(" "),
        error: [
          "border-red-500/30 bg-red-100 dark:bg-red-900/40",
          "text-red-700 dark:text-red-300",
        ].join(" "),
        info: [
          "border-blue-500/30 bg-blue-500",
          "text-white",
        ].join(" "),
        // Status-specific variants
        draft: [
          "border-charcoal-300 dark:border-charcoal-600",
          "bg-charcoal-100 dark:bg-charcoal-800",
          "text-charcoal-600 dark:text-charcoal-400",
        ].join(" "),
        pending: [
          "border-amber-400/50 bg-amber-100 dark:bg-amber-900/30",
          "text-amber-700 dark:text-amber-400",
        ].join(" "),
        approved: [
          "border-emerald-400/50 bg-emerald-100 dark:bg-emerald-900/30",
          "text-emerald-700 dark:text-emerald-400",
        ].join(" "),
        rejected: [
          "border-red-400/50 bg-red-100 dark:bg-red-900/30",
          "text-red-700 dark:text-red-400",
        ].join(" "),
        locked: [
          "border-charcoal-400/50 bg-charcoal-700 dark:bg-charcoal-600",
          "text-white",
        ].join(" "),
      },
      size: {
        default: "px-3 py-1 text-xs",
        sm: "px-2.5 py-0.5 text-xs",
        lg: "px-4 py-1.5 text-sm",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, size, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant, size }), className)} {...props} />
  )
}

interface StatusBadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  status: TicketStatus
  size?: 'sm' | 'default' | 'lg'
}

function StatusBadge({ status, size = 'default', className, ...props }: StatusBadgeProps) {
  // Map ticket status to badge variant
  const variantMap: Record<TicketStatus, VariantProps<typeof badgeVariants>['variant']> = {
    draft: 'draft',
    pending_review: 'pending',
    approved: 'approved',
    rejected: 'rejected',
    locked: 'locked',
  }

  return (
    <Badge
      variant={variantMap[status]}
      size={size}
      className={className}
      {...props}
    >
      {getStatusLabel(status)}
    </Badge>
  )
}

export { Badge, StatusBadge, badgeVariants }
