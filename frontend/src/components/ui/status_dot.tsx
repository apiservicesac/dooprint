import { cn } from '@/lib/utils'

type Tone = 'success' | 'danger' | 'warning' | 'muted'

const TONES: Record<Tone, string> = {
  success: 'bg-brand-teal',
  danger: 'bg-danger',
  warning: 'bg-yellow-500',
  muted: 'bg-muted-foreground/40',
}

/** Status at a glance: a colored dot and its label. */
export default function StatusDot({ tone, label, className }: { tone: Tone; label: string; className?: string }) {
  return (
    <span className={cn('inline-flex items-center gap-2 text-sm text-muted-foreground whitespace-nowrap', className)}>
      <span className={cn('inline-block h-2.5 w-2.5 rounded-full', TONES[tone])} />
      {label}
    </span>
  )
}
