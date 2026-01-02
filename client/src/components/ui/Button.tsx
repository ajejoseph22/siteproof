import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"
import { Loader2 } from "lucide-react"

/**
 * Industrial-grade button variants for SiteProof
 * Designed for high visibility and easy touch targets on construction sites
 */
const buttonVariants = cva(
  [
    // Base styles
    "inline-flex items-center justify-center gap-2.5",
    "whitespace-nowrap font-semibold",
    "transition-all duration-200",
    "disabled:pointer-events-none disabled:opacity-50",
    "[&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4",
    "shrink-0",
    // Active state feedback
    "active:scale-[0.97]",
    // Focus ring
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-safety-500 focus-visible:ring-offset-2",
    "dark:focus-visible:ring-offset-charcoal-900",
  ].join(" "),
  {
    variants: {
      variant: {
        // Primary - Safety Orange gradient with depth
        default: [
          "bg-gradient-to-b from-safety-500 to-safety-600",
          "text-white font-semibold",
          "shadow-md shadow-safety-500/25",
          "border border-safety-600/50",
          "hover:from-safety-400 hover:to-safety-500",
          "hover:shadow-lg hover:shadow-safety-500/30",
          "hover:-translate-y-0.5",
        ].join(" "),

        // Destructive - Bold red for dangerous actions
        destructive: [
          "bg-gradient-to-b from-red-500 to-red-600",
          "text-white font-semibold",
          "shadow-md shadow-red-500/25",
          "border border-red-700/50",
          "hover:from-red-400 hover:to-red-500",
          "hover:shadow-lg hover:shadow-red-500/30",
          "hover:-translate-y-0.5",
        ].join(" "),

        // Outline - Clean bordered button
        outline: [
          "border-2 border-charcoal-300 dark:border-charcoal-600",
          "bg-white dark:bg-charcoal-800",
          "text-charcoal-700 dark:text-charcoal-200",
          "shadow-sm",
          "hover:border-safety-500 hover:text-safety-600",
          "hover:bg-safety-50 dark:hover:bg-safety-950/20",
          "hover:shadow-md",
        ].join(" "),

        // Secondary - Subtle charcoal
        secondary: [
          "bg-charcoal-800 dark:bg-charcoal-700",
          "text-white",
          "shadow-md shadow-charcoal-900/20",
          "border border-charcoal-700 dark:border-charcoal-600",
          "hover:bg-charcoal-700 dark:hover:bg-charcoal-600",
          "hover:shadow-lg",
        ].join(" "),

        // Ghost - Minimal, no background
        ghost: [
          "text-charcoal-600 dark:text-charcoal-400",
          "hover:bg-charcoal-100 dark:hover:bg-charcoal-800",
          "hover:text-charcoal-900 dark:hover:text-white",
        ].join(" "),

        // Link - Text-only with underline
        link: [
          "text-safety-500 dark:text-safety-400",
          "underline-offset-4",
          "hover:underline",
          "hover:text-safety-600 dark:hover:text-safety-300",
        ].join(" "),

        // Success - Green for confirmations
        success: [
          "bg-gradient-to-b from-emerald-500 to-emerald-600",
          "text-white font-semibold",
          "shadow-md shadow-emerald-500/25",
          "border border-emerald-700/50",
          "hover:from-emerald-400 hover:to-emerald-500",
          "hover:shadow-lg hover:shadow-emerald-500/30",
          "hover:-translate-y-0.5",
        ].join(" "),
      },
      size: {
        default: "h-11 px-5 py-2.5 text-sm rounded-lg",
        sm: "h-9 px-4 py-2 text-xs rounded-lg",
        lg: "h-12 px-6 py-3 text-base rounded-xl",
        xl: "h-14 px-8 py-4 text-lg rounded-xl",
        icon: "h-11 w-11 rounded-lg",
        "icon-sm": "h-9 w-9 rounded-lg",
        "icon-lg": "h-12 w-12 rounded-xl",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
  isLoading?: boolean
  leftIcon?: React.ReactNode
  rightIcon?: React.ReactNode
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, isLoading, leftIcon, rightIcon, children, disabled, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || isLoading}
        {...props}
      >
        {isLoading ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          leftIcon
        )}
        {children}
        {!isLoading && rightIcon}
      </Comp>
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }
