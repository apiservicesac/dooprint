import { cn } from '@/lib/utils'

type Tone = 'success' | 'danger' | 'warning' | 'muted'

const TONES: Record<Tone, string> = {
  success: 'bg-emerald-500',
  danger: 'bg-red-500',
  warning: 'bg-amber-500',
  muted: 'bg-zinc-300 dark:bg-zinc-600',
}

/** Status at a glance: a colored dot and its label. */
export default function StatusDot({ tone, label, className }: { tone: Tone; label: string; className?: string }) {
  return (
    <span className={cn('inline-flex items-center gap-2 whitespace-nowrap text-sm text-muted-foreground', className)}>
      <span className={cn('inline-block h-2 w-2 rounded-full', TONES[tone])} />
      {label}
    </span>
  )
}
