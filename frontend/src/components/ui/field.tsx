import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export function Field({ label, hint, children }: { label: string; hint?: string; children: ReactNode }) {
  return (
    <label className="block">
      <span className="mb-2 block text-sm font-medium text-foreground">{label}</span>
      {children}
      {hint && <span className="mt-2 block text-sm text-muted-foreground">{hint}</span>}
    </label>
  )
}

interface RadioCardProps {
  checked: boolean
  onChange: () => void
  title: string
  description: string
}

/** Large option with an explanation, used to choose the connection mode. */
export function RadioCard({ checked, onChange, title, description }: RadioCardProps) {
  return (
    <label
      className={cn(
        'flex cursor-pointer items-start gap-3 rounded-lg border p-4 transition-colors',
        checked ? 'border-brand-orange bg-brand-orange-light/60' : 'border-border hover:bg-muted/50',
      )}
    >
      <input type="radio" checked={checked} onChange={onChange} className="mt-1 accent-brand-orange" />
      <span>
        <span className="block text-[15px] font-medium text-foreground">{title}</span>
        <span className="mt-1 block text-sm leading-6 text-muted-foreground">{description}</span>
      </span>
    </label>
  )
}
