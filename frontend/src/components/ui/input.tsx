import * as React from 'react'
import { cn } from '@/lib/utils'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  icon?: React.ReactElement
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(({ className, type, icon, ...props }, ref) => (
  <div className="relative flex items-center w-full">
    {icon && <div className="absolute left-3 text-muted-foreground pointer-events-none flex items-center">{icon}</div>}
    <input
      type={type}
      className={cn(
        'flex h-11 w-full rounded-md border border-neutral-300 dark:border-neutral-600 bg-background px-3.5 py-2 text-[15px]',
        'placeholder:text-muted-foreground focus-visible:outline-none transition',
        'read-only:text-neutral-700 dark:read-only:text-neutral-400 read-only:ring-0! read-only:border-neutral-300! dark:read-only:border-neutral-600!',
        'focus:border-neutral-500 dark:focus:border-neutral-400 focus:ring-4 focus:ring-neutral-200 dark:focus:ring-neutral-700',
        'disabled:cursor-not-allowed disabled:opacity-50',
        icon && 'pl-8',
        className,
      )}
      ref={ref}
      name={props.id}
      {...props}
    />
  </div>
))

Input.displayName = 'Input'
export default Input
