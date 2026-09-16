import { cn } from '@/lib/utils'

type BadgeVariant = 'default' | 'success' | 'warning' | 'danger' | 'muted'

const variants: Record<BadgeVariant, string> = {
  default: 'bg-brand-orange-light text-brand-orange-dark ring-brand-orange/20',
  success: 'bg-brand-teal-light text-brand-teal ring-brand-teal/20',
  warning: 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-400 ring-yellow-600/20',
  danger:  'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 ring-red-600/20',
  muted:   'bg-muted dark:bg-zinc-800 text-muted-foreground dark:text-zinc-400 ring-zinc-500/20',
}

interface BadgeProps {
  variant?:  BadgeVariant
  className?: string
  children:  React.ReactNode
}

export function Badge({ variant = 'default', className, children }: BadgeProps) {
  return (
    <span className={cn(
      'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset',
      variants[variant],
      className,
    )}>
      {children}
    </span>
  )
}
