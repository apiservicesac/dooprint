import type { ReactNode } from 'react'

export default function Stat({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="rounded-sm border border-border bg-muted/40 px-4 py-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 text-lg font-medium text-foreground">{value}</div>
    </div>
  )
}
