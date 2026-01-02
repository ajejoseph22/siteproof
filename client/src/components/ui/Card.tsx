import * as React from "react"
import { cn } from "@/lib/utils"

/**
 * Industrial Card Component
 * Clean surfaces with subtle depth for construction app readability
 */
const Card = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    variant?: 'default' | 'elevated' | 'outline' | 'ghost' | 'highlight'
    hover?: boolean
    padding?: 'none' | 'sm' | 'md' | 'lg'
  }
>(({ className, variant = 'default', hover = false, padding = 'md', ...props }, ref) => {
  const variants = {
    default: [
      "bg-white dark:bg-charcoal-900",
      "border-2 border-charcoal-200 dark:border-charcoal-800",
      "shadow-sm",
    ].join(" "),
    elevated: [
      "bg-white dark:bg-charcoal-900",
      "shadow-xl shadow-charcoal-900/10 dark:shadow-charcoal-950/50",
      "border border-charcoal-100 dark:border-charcoal-800",
    ].join(" "),
    outline: [
      "bg-transparent",
      "border-2 border-charcoal-300 dark:border-charcoal-700",
    ].join(" "),
    ghost: [
      "bg-charcoal-50 dark:bg-charcoal-900/50",
      "border border-charcoal-100 dark:border-charcoal-800/50",
    ].join(" "),
    highlight: [
      "bg-gradient-to-br from-safety-50 to-white dark:from-safety-950/20 dark:to-charcoal-900",
      "border-2 border-safety-200 dark:border-safety-800/50",
      "shadow-md shadow-safety-500/5",
    ].join(" "),
  }

  const paddings = {
    none: "",
    sm: "p-4",
    md: "p-5 sm:p-6",
    lg: "p-6 sm:p-8",
  }

  return (
    <div
      ref={ref}
      className={cn(
        "rounded-xl text-charcoal-900 dark:text-charcoal-50",
        variants[variant],
        paddings[padding],
        hover && [
          "cursor-pointer",
          "transition-all duration-200",
          "hover:shadow-xl hover:shadow-charcoal-900/10 dark:hover:shadow-charcoal-950/30",
          "hover:-translate-y-1",
          "hover:border-safety-300 dark:hover:border-safety-700/50",
        ].join(" "),
        className
      )}
      {...props}
    />
  )
})
Card.displayName = "Card"

const CardHeader = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("flex items-center justify-between mb-5", className)}
    {...props}
  />
))
CardHeader.displayName = "CardHeader"

const CardTitle = React.forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement> & { as?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6' }
>(({ className, as: Tag = 'h3', ...props }, ref) => (
  <Tag
    ref={ref}
    className={cn(
      "font-display text-xl tracking-wide",
      "text-charcoal-900 dark:text-white",
      className
    )}
    {...props}
  />
))
CardTitle.displayName = "CardTitle"

const CardDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <p
    ref={ref}
    className={cn(
      "text-sm text-charcoal-500 dark:text-charcoal-400",
      "leading-relaxed",
      className
    )}
    {...props}
  />
))
CardDescription.displayName = "CardDescription"

const CardContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div ref={ref} className={cn("", className)} {...props} />
))
CardContent.displayName = "CardContent"

const CardFooter = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      "flex items-center justify-end gap-3",
      "mt-6 pt-5",
      "border-t-2 border-charcoal-100 dark:border-charcoal-800",
      className
    )}
    {...props}
  />
))
CardFooter.displayName = "CardFooter"

export { Card, CardHeader, CardFooter, CardTitle, CardDescription, CardContent }
