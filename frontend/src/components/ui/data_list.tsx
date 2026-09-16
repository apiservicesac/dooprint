import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface DataListProps {
  rows: [string, ReactNode][]
  columns?: 1 | 2
}

/** Label/value list for technical details. A value can be text or a badge. */
export default function DataList({ rows, columns = 2 }: DataListProps) {
  return (
    <dl className={cn('grid gap-x-8 gap-y-3 text-sm', columns === 2 && 'sm:grid-cols-2')}>
      {rows.map(([label, value]) => (
        <div key={label} className="flex items-center justify-between gap-4 border-b border-border pb-2">
          <dt className="whitespace-nowrap text-muted-foreground">{label}</dt>
          <dd className="break-all text-right font-medium text-foreground">
            {value === '' || value === null || value === undefined ? '—' : value}
          </dd>
        </div>
      ))}
    </dl>
  )
}
