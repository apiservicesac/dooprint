import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export function Field({ label, hint, children }: { label: string; hint?: string; children: ReactNode }) {
  return (
    <label className="block text-sm">
      <span className="mb-1 block font-medium text-foreground">{label}</span>
      {children}
      {hint && <span className="mt-1 block text-xs text-muted-foreground">{hint}</span>}
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
        'flex cursor-pointer items-start gap-3 rounded-sm border p-3 transition-colors',
        checked ? 'border-brand-orange bg-brand-orange-light' : 'border-border hover:border-neutral-400',
      )}
    >
      <input type="radio" checked={checked} onChange={onChange} className="mt-1 accent-brand-orange" />
      <span className="text-sm">
        <span className="font-medium text-foreground">{title}</span>
        <span className="mt-0.5 block text-muted-foreground">{description}</span>
      </span>
    </label>
  )
}
