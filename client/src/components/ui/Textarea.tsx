import * as React from "react"
import { cn } from "@/lib/utils"
import { Label } from "./label"

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string
  error?: string
  helperText?: string
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, label, error, helperText, id, ...props }, ref) => {
    const textareaId = id || label?.toLowerCase().replace(/\s+/g, '-')

    return (
      <div className="w-full">
        {label && (
          <Label htmlFor={textareaId} className="mb-1.5">
            {label}
          </Label>
        )}
        <textarea
          id={textareaId}
          className={cn(
            "flex min-h-[120px] w-full rounded-lg border-2 border-charcoal-200 bg-white px-4 py-3 text-base transition-colors",
            "placeholder:text-charcoal-400",
            "focus-visible:outline-none focus-visible:border-safety-500 focus-visible:ring-2 focus-visible:ring-safety-500/20",
            "disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-charcoal-50",
            "dark:border-charcoal-700 dark:bg-charcoal-800 dark:text-charcoal-100 dark:placeholder:text-charcoal-500 dark:disabled:bg-charcoal-900",
            "resize-y",
            error && "border-red-500 focus-visible:border-red-500 focus-visible:ring-red-500/20",
            className
          )}
          ref={ref}
          {...props}
        />
        {error && (
          <p className="mt-1.5 text-sm text-red-600 dark:text-red-400">{error}</p>
        )}
        {helperText && !error && (
          <p className="mt-1.5 text-sm text-charcoal-500 dark:text-charcoal-400">{helperText}</p>
        )}
      </div>
    )
  }
)
Textarea.displayName = "Textarea"

export { Textarea }
