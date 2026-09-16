import type { ReactNode } from 'react'

export default function Stat({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div>
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className="mt-1.5 text-2xl font-semibold tracking-tight text-foreground">{value}</div>
    </div>
  )
}
