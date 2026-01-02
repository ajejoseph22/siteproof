import * as React from "react"
import { cn } from "@/lib/utils"
import { Label } from "./label"

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  helperText?: string
  leftIcon?: React.ReactNode
  rightIcon?: React.ReactNode
}

/**
 * Industrial Input Component
 * Large touch targets, high contrast for field use
 */
const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, label, error, helperText, leftIcon, rightIcon, id, ...props }, ref) => {
    const inputId = id || label?.toLowerCase().replace(/\s+/g, '-')

    return (
      <div className="w-full">
        {label && (
          <Label htmlFor={inputId} className="mb-2">
            {label}
          </Label>
        )}
        <div className="relative">
          {leftIcon && (
            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-charcoal-400 dark:text-charcoal-500 pointer-events-none">
              {leftIcon}
            </div>
          )}
          <input
            type={type}
            id={inputId}
            className={cn(
              // Base styles
              "flex h-12 w-full rounded-xl px-4 py-3",
              "text-base font-medium",
              "transition-all duration-200",
              // Border & background
              "border-2 border-charcoal-200 dark:border-charcoal-700",
              "bg-white dark:bg-charcoal-800",
              "text-charcoal-900 dark:text-charcoal-100",
              // Placeholder
              "placeholder:text-charcoal-400 dark:placeholder:text-charcoal-500",
              // Focus state
              "focus-visible:outline-none",
              "focus-visible:border-safety-500",
              "focus-visible:ring-4 focus-visible:ring-safety-500/20",
              // File input
              "file:border-0 file:bg-transparent file:text-sm file:font-semibold",
              "file:text-safety-600 file:mr-4",
              // Disabled state
              "disabled:cursor-not-allowed disabled:opacity-50",
              "disabled:bg-charcoal-100 dark:disabled:bg-charcoal-900",
              // Icon padding
              leftIcon && "pl-12",
              rightIcon && "pr-12",
              // Error state
              error && [
                "border-red-500 dark:border-red-500",
                "focus-visible:border-red-500",
                "focus-visible:ring-red-500/20",
              ].join(" "),
              className
            )}
            ref={ref}
            {...props}
          />
          {rightIcon && (
            <div className="absolute right-4 top-1/2 -translate-y-1/2 text-charcoal-400 dark:text-charcoal-500">
              {rightIcon}
            </div>
          )}
        </div>
        {error && (
          <p className="mt-2 text-sm font-medium text-red-600 dark:text-red-400 flex items-center gap-1.5">
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            {error}
          </p>
        )}
        {helperText && !error && (
          <p className="mt-2 text-sm text-charcoal-500 dark:text-charcoal-400">{helperText}</p>
        )}
      </div>
    )
  }
)
Input.displayName = "Input"

export { Input }
